package handler

import (
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"singleOpenVpn/internal/config"
	"singleOpenVpn/internal/model"
	"singleOpenVpn/internal/vpn"
)

// parseStatusConnectedAt 解析 status 中的 Connected Since，无时区字符串按东八区理解
func parseStatusConnectedAt(s string) time.Time {
	return vpn.ParseConnectedSince(s)
}

// isValidVPNUsername 仅允许英文字母、数字、下划线
func isValidVPNUsername(s string) bool {
	for _, r := range s {
		if r != '_' && (r < '0' || r > '9') && (r < 'A' || r > 'Z') && (r < 'a' || r > 'z') {
			return false
		}
	}
	return true
}

// parseOptionalDate 解析可选日期 YYYY-MM-DD，空或无效返回 nil（按东八区当天 00:00:00 入库）
func parseOptionalDate(s string) *model.NaiveTime {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	t, err := time.ParseInLocation("2006-01-02", s, model.Shanghai)
	if err != nil {
		return nil
	}
	nt := model.NT(t)
	return &nt
}

// UserListRow 用户列表行（含在线与流量信息）
type UserListRow struct {
	model.VPNUser
	HasPassword  bool   `json:"has_password"`
	CurrentIP    string `json:"current_ip"` // 当前获取到的 VPN IP（在线时来自 status，否则为空）
	Online       bool   `json:"online"`
	ConnectedAt  string `json:"connected_at"`  // 上线时间（在线时）
	DurationSecs int64  `json:"duration_secs"` // 在线时长（秒）
	BytesRecv    int64  `json:"bytes_recv"`    // 上传流量（服务器收到）
	BytesSent    int64  `json:"bytes_sent"`    // 下载流量（服务器发出）
}

func ListUsers(db *gorm.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var users []model.VPNUser
		if err := db.Find(&users).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		usageByCN := make(map[string]vpn.ClientUsage)
		if path := cfg.StatusFilePath; path != "" {
			if list, err := vpn.ParseStatusFile(path); err == nil {
				for _, u := range list {
					if u.CommonName != "" {
						usageByCN[u.CommonName] = u
					}
				}
			}
		}
		rows := make([]UserListRow, 0, len(users))
		for _, u := range users {
			r := UserListRow{VPNUser: u, HasPassword: u.PasswordHash != ""}
			if stat, ok := usageByCN[u.Name]; ok {
				r.CurrentIP = stat.VirtualIP
				r.Online = true
				if t := vpn.ParseConnectedSince(stat.ConnectedAt); !t.IsZero() {
					r.ConnectedAt = t.In(model.Shanghai).Format(model.NaiveTimeLayout)
					// 与上线时间同源计算时长，避免 status 解析阶段 DurationSecs 为 0 时前端一直显示 '-'
					r.DurationSecs = int64(time.Since(t).Seconds())
				} else {
					r.ConnectedAt = stat.ConnectedAt
					r.DurationSecs = stat.DurationSecs
				}
				r.BytesRecv = stat.BytesRecv
				r.BytesSent = stat.BytesSent
			} else if u.StaticIP != "" {
				r.CurrentIP = u.StaticIP
			}
			rows = append(rows, r)
		}
		c.JSON(http.StatusOK, rows)
	}
}

func hashPassword(pass string) (string, error) {
	if pass == "" {
		return "", nil
	}
	b, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func CreateUser(db *gorm.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body struct {
			Name        string `json:"name"`
			Kind        string `json:"kind"` // user | client
			Email       string `json:"email"`
			Remark      string `json:"remark"`
			Password    string `json:"password"`
			StaticIP    string `json:"static_ip"`
			RouteNopull bool   `json:"route_nopull"` // 下载的配置中是否添加 route-nopull
			ExpiresAt   string `json:"expires_at"`   // 可选，日期 YYYY-MM-DD，空表示永不过期
			Enabled     bool   `json:"enabled"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
			return
		}
		if body.Name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "用户名为空"})
			return
		}
		if !isValidVPNUsername(body.Name) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "VPN 用户名仅限英文字母、数字和下划线，不能含中文或特殊字符"})
			return
		}
		hash, err := hashPassword(body.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
			return
		}
		// 静态 IP / CCD 不预填：仅当请求中明确传入非空 static_ip 时才写入 CCD
		body.StaticIP = strings.TrimSpace(body.StaticIP)
		kind := body.Kind
		if kind != model.KindClient {
			kind = model.KindUser
		}
		expAt := parseOptionalDate(body.ExpiresAt)
		u := model.VPNUser{
			Name:         body.Name,
			Kind:         kind,
			Email:        body.Email,
			Remark:       strings.TrimSpace(body.Remark),
			PasswordHash: hash,
			StaticIP:     body.StaticIP,
			RouteNopull:  body.RouteNopull,
			ExpiresAt:    expAt,
			Enabled:      body.Enabled,
		}
		if err := db.Create(&u).Error; err != nil {
			c.JSON(http.StatusConflict, gin.H{"error": "用户名已存在"})
			return
		}
		if err := vpn.GenClientCert(cfg.VPNBasePath, u.Name); err != nil {
			c.JSON(http.StatusOK, gin.H{"message": "用户已创建，但证书生成失败，请检查 easy-rsa 环境", "user": u})
			return
		}
		if u.StaticIP != "" {
			_ = vpn.WriteCCD(cfg.VPNBasePath, u.Name, u.StaticIP)
		}
		c.JSON(http.StatusOK, u)
	}
}

func UpdateUser(db *gorm.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body struct {
			ID          uint   `json:"id"`
			Name        string `json:"name"`
			Kind        string `json:"kind"` // user | client
			Email       string `json:"email"`
			Remark      string `json:"remark"`
			Password    string `json:"password"` // 可选，有则更新密码
			StaticIP    string `json:"static_ip"`
			RouteNopull bool   `json:"route_nopull"` // 下载的配置中是否添加 route-nopull
			ExpiresAt   string `json:"expires_at"`   // 可选，空表示永不过期
			Enabled     bool   `json:"enabled"`
			CCDContent  string `json:"ccd_content"` // 可选，有则写入 CCD 文件（优先于 static_ip 单行）
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
			return
		}
		if body.ID == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "缺少 id"})
			return
		}
		kind := body.Kind
		if kind != model.KindClient {
			kind = model.KindUser
		}
		updates := map[string]interface{}{
			"name": body.Name, "kind": kind, "email": body.Email, "remark": strings.TrimSpace(body.Remark),
			"enabled": body.Enabled, "static_ip": body.StaticIP, "route_nopull": body.RouteNopull, "expires_at": parseOptionalDate(body.ExpiresAt),
		}
		if body.Password != "" {
			hash, err := hashPassword(body.Password)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
				return
			}
			updates["password_hash"] = hash
		}
		if err := db.Model(&model.VPNUser{}).Where("id = ?", body.ID).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		// Updates(map) 不会走 NaiveTime.Valuer，驱动会把 updated_at 写成 RFC3339Nano；强制写成 naive 字符串
		_ = db.Model(&model.VPNUser{}).Where("id = ?", body.ID).UpdateColumn("updated_at", model.NowNaive().String())
		// 同步 CCD：若传了 ccd_content 则写完整内容；否则有 static_ip 写单行，无则删
		var u model.VPNUser
		_ = db.First(&u, body.ID).Error
		if body.CCDContent != "" {
			_ = vpn.WriteCCDContent(cfg.VPNBasePath, u.Name, body.CCDContent)
		} else if strings.TrimSpace(body.StaticIP) != "" {
			_ = vpn.WriteCCD(cfg.VPNBasePath, u.Name, strings.TrimSpace(body.StaticIP))
		} else {
			_ = vpn.RemoveCCD(cfg.VPNBasePath, u.Name)
		}
		// 清除该客户端在 ipp.txt 中的持久化 IP，使下次连接时按新 CCD 分配 IP
		_ = vpn.RemoveClientFromIPPPool(cfg.VPNBasePath, u.Name)
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	}
}

func DeleteUser(db *gorm.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var u model.VPNUser
		if err := db.First(&u, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
			return
		}
		_ = vpn.RevokeClient(cfg.VPNBasePath, u.Name)
		_ = vpn.RemoveCCD(cfg.VPNBasePath, u.Name)
		if err := db.Delete(&u).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "已删除"})
	}
}

// GetUserCCD 获取该用户的 CCD 文件内容；无文件时返回空，由前端展示空白表单
func GetUserCCD(db *gorm.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var u model.VPNUser
		if err := db.First(&u, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
			return
		}
		content, err := vpn.ReadCCD(cfg.VPNBasePath, u.Name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "读取 CCD 失败: " + err.Error()})
			return
		}
		// 无文件时返回空，由前端展示空白表单，不预填默认值
		c.JSON(http.StatusOK, gin.H{"content": content, "username": u.Name})
	}
}

// SetUserCCD 保存该用户的 CCD 文件内容（高级配置）
func SetUserCCD(db *gorm.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var u model.VPNUser
		if err := db.First(&u, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
			return
		}
		var body struct {
			Content string `json:"content"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
			return
		}
		if err := vpn.WriteCCDContent(cfg.VPNBasePath, u.Name, body.Content); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "写入 CCD 失败: " + err.Error()})
			return
		}
		_ = vpn.RemoveClientFromIPPPool(cfg.VPNBasePath, u.Name)
		c.JSON(http.StatusOK, gin.H{"message": "已保存，客户端重连 VPN 后生效"})
	}
}

// 客户端 .ovpn 中 remote 使用的配置 key（存于 server_configs 表）
const (
	configKeyClientRemoteHost = "client_remote_host"
	configKeyClientRemotePort = "client_remote_port"
)

// DownloadClientConfig 下载客户端 .ovpn 文件
func DownloadClientConfig(db *gorm.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var u model.VPNUser
		if err := db.First(&u, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
			return
		}
		serverAddr := c.Query("server")
		if serverAddr == "" {
			var sc model.ServerConfig
			if err := db.Where("key = ?", configKeyClientRemoteHost).First(&sc).Error; err == nil && sc.Value != "" {
				serverAddr = sc.Value
			}
			if serverAddr == "" {
				serverAddr = "127.0.0.1"
			}
		}
		port := 1194
		if p := c.Query("port"); p != "" {
			if v, err := strconv.Atoi(p); err == nil {
				port = v
			}
		} else {
			var sc model.ServerConfig
			if err := db.Where("key = ?", configKeyClientRemotePort).First(&sc).Error; err == nil && sc.Value != "" {
				if v, err := strconv.Atoi(sc.Value); err == nil && v > 0 {
					port = v
				}
			}
		}
		ovpnPath, err := vpn.GenClientOVPN(cfg.VPNBasePath, u.Name, serverAddr, port, u.RouteNopull)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "生成配置失败: " + err.Error()})
			return
		}
		name := filepath.Base(ovpnPath)
		c.Header("Content-Disposition", "attachment; filename="+name)
		c.Header("Content-Type", "application/x-openvpn-profile")
		c.File(ovpnPath)
	}
}
