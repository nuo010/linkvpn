package handler

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"singleOpenVpn/internal/config"
	"singleOpenVpn/internal/model"
	"singleOpenVpn/internal/vpn"
)

// NeedInitialClientConfig 返回是否需要首次设置「客户端下载配置」（服务器地址与端口）。
// 挂载目录为空或从未保存过时 need 为 true，前端应弹窗强制用户设置后再使用。
func NeedInitialClientConfig(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var sc model.ServerConfig
		err := db.Where("key = ?", configKeyClientRemoteHost).First(&sc).Error
		need := err != nil || strings.TrimSpace(sc.Value) == ""
		c.JSON(http.StatusOK, gin.H{"need": need})
	}
}

func GetServerConfig(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var list []model.ServerConfig
		if err := db.Find(&list).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		m := make(map[string]string)
		for _, item := range list {
			m[item.Key] = item.Value
		}
		c.JSON(http.StatusOK, m)
	}
}

func SetServerConfig(db *gorm.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body map[string]string
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
			return
		}
		for k, v := range body {
			var sc model.ServerConfig
			db.Where("key = ?", k).FirstOrInit(&sc)
			sc.Key = k
			sc.Value = v
			if err := db.Save(&sc).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	}
}

func GetServerConfigFile(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := cfg.VPNBasePath + "/server.conf"
		data, err := os.ReadFile(path)
		content := string(data)
		// 文件不存在或为空时返回默认配置，便于用户直接保存使用
		if err != nil || len(content) == 0 || len(strings.TrimSpace(content)) == 0 {
			content = vpn.DefaultServerConfig(cfg.VPNBasePath)
		}
		c.JSON(http.StatusOK, gin.H{"content": content, "path": path})
	}
}

func PutServerConfigFile(db *gorm.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body struct {
			Content string `json:"content"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
			return
		}
		if err := vpn.WriteServerConfig(cfg.VPNBasePath, body.Content); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "已保存"})
	}
}

func InitVPN(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := vpn.InitPKI(cfg.VPNBasePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "init-pki: " + err.Error()})
			return
		}
		if err := vpn.BuildCA(cfg.VPNBasePath, cfg.ServerName); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "build-ca: " + err.Error()})
			return
		}
		if err := vpn.GenDH(cfg.VPNBasePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "gen-dh: " + err.Error()})
			return
		}
		if err := vpn.GenServerCert(cfg.VPNBasePath, "server"); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "gen-server-cert: " + err.Error()})
			return
		}
		content := vpn.DefaultServerConfig(cfg.VPNBasePath)
		if err := vpn.WriteServerConfig(cfg.VPNBasePath, content); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "VPN 服务端 PKI 与配置已初始化"})
	}
}

// GetDefaultServerConfig 返回默认 server.conf 内容（用于「加载默认配置」）
func GetDefaultServerConfig(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		content := vpn.DefaultServerConfig(cfg.VPNBasePath)
		c.JSON(http.StatusOK, gin.H{"content": content})
	}
}

// RestartVPNService 重启 OpenVPN：与 start.sh 一致——先杀掉本机 openvpn 进程，再 vpn.StartOpenVPN 追加写入 openvpn.log（仅 Linux 使用）
func RestartVPNService(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		configPath := filepath.Join(cfg.VPNBasePath, "server.conf")
		if _, err := os.Stat(configPath); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "未找到 server.conf，请先初始化 PKI 并保存配置"})
			return
		}
		for i := 0; i < 40; i++ {
			pid, ok := vpn.FindOpenVPNPID()
			if !ok {
				break
			}
			_ = syscall.Kill(pid, syscall.SIGKILL)
			time.Sleep(150 * time.Millisecond)
		}
		if _, still := vpn.FindOpenVPNPID(); still {
			_ = exec.Command("pkill", "-9", "openvpn").Run()
			time.Sleep(500 * time.Millisecond)
		}
		time.Sleep(400 * time.Millisecond)
		if err := vpn.StartOpenVPN(cfg.VPNBasePath); err != nil {
			log.Printf("RestartVPNService: StartOpenVPN failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "结束旧进程后启动失败: " + err.Error() + "（请确认面板进程有权限启动 openvpn，或手动在容器内执行 openvpn --config ...）"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "OpenVPN 已重启，日志写入 " + filepath.Join(cfg.VPNBasePath, "openvpn.log")})
	}
}

func GetVPNStatus(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := filepath.Join(cfg.VPNBasePath, "easy-rsa", "pki", "ca.crt")
		hasCA := false
		if _, err := os.Stat(path); err == nil {
			hasCA = true
		}
		c.JSON(http.StatusOK, gin.H{
			"pki_initialized": hasCA,
			"base_path":       cfg.VPNBasePath,
		})
	}
}

// GetOpenVPNParams 获取 OpenVPN 参数配置
func GetOpenVPNParams(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var sc model.ServerConfig
		err := db.Where("key = ?", model.OVPNParamsKey).First(&sc).Error
		p := model.DefaultOpenVPNParams()
		if err == nil && sc.Value != "" {
			_ = json.Unmarshal([]byte(sc.Value), &p)
		}
		c.JSON(http.StatusOK, p)
	}
}

// SetOpenVPNParams 保存 OpenVPN 参数；若 auto_apply_to_config 为 true 则同时写入 server.conf
func SetOpenVPNParams(db *gorm.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var p model.OpenVPNParams
		if err := c.ShouldBindJSON(&p); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
			return
		}
		if err := validateOpenVPNParams(&p); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		raw, err := json.Marshal(p)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "序列化失败"})
			return
		}
		var sc model.ServerConfig
		db.Where("key = ?", model.OVPNParamsKey).FirstOrInit(&sc)
		sc.Key = model.OVPNParamsKey
		sc.Value = string(raw)
		if err := db.Save(&sc).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if p.AutoApplyToConfig {
			content := vpn.BuildServerConfigFromParams(
				cfg.VPNBasePath, p.Port, p.Protocol, p.Device, p.Topology, p.MaxConnections, p.Subnet, p.PushRoutes, p.Management,
				p.PushDNS1, p.PushDNS2, p.Keepalive, p.Cipher, p.Auth, p.RunUser, p.RunGroup, p.Verb,
				p.VPNGateway, p.ClientToClient, p.IPv6, p.PersistKey, p.PersistTun, p.ExplicitExitNotify, p.IPv6Subnet,
			)
			if err := vpn.WriteServerConfig(cfg.VPNBasePath, content); err != nil {
				c.JSON(http.StatusOK, gin.H{"message": "参数已保存，但写入 server.conf 失败: " + err.Error()})
				return
			}
		}
		c.JSON(http.StatusOK, gin.H{"message": "已保存"})
	}
}

func validateOpenVPNParams(p *model.OpenVPNParams) error {
	if p == nil {
		return nil
	}
	if port, err := strconv.Atoi(strings.TrimSpace(p.Port)); err != nil || port < 1 || port > 65535 {
		return errInvalid("端口必须为 1-65535 的整数")
	}
	if strings.TrimSpace(p.MaxConnections) != "" {
		if n, err := strconv.Atoi(strings.TrimSpace(p.MaxConnections)); err != nil || n < 1 {
			return errInvalid("最大连接数必须为大于 0 的整数")
		}
	}
	if err := validateCIDRorIPv4Network(strings.TrimSpace(p.Subnet), true, "虚拟网段"); err != nil {
		return err
	}
	if p.IPv6 {
		if err := validateIPv6CIDR(strings.TrimSpace(p.IPv6Subnet), "IPv6 网段"); err != nil {
			return err
		}
	}
	if err := validateKeepalive(strings.TrimSpace(p.Keepalive)); err != nil {
		return err
	}
	if err := validateIPAddress(strings.TrimSpace(p.PushDNS1), "推送 DNS 1"); err != nil {
		return err
	}
	if err := validateIPAddress(strings.TrimSpace(p.PushDNS2), "推送 DNS 2"); err != nil {
		return err
	}
	if err := validatePushRoutes(strings.TrimSpace(p.PushRoutes)); err != nil {
		return err
	}
	if strings.TrimSpace(p.Verb) != "" {
		if n, err := strconv.Atoi(strings.TrimSpace(p.Verb)); err != nil || n < 0 || n > 11 {
			return errInvalid("日志级别必须为 0-11 的整数")
		}
	}
	return nil
}

func validateCIDRorIPv4Network(raw string, allowBare bool, field string) error {
	if raw == "" {
		return errInvalid(field + "不能为空")
	}
	if strings.Contains(raw, "/") {
		if _, _, err := net.ParseCIDR(raw); err != nil {
			return errInvalid(field + "格式错误，应为如 10.8.8.0/24")
		}
		return nil
	}
	if allowBare && net.ParseIP(raw) != nil {
		return nil
	}
	return errInvalid(field + "格式错误，应为如 10.8.8.0/24")
}

func validateIPv6CIDR(raw, field string) error {
	if raw == "" {
		return errInvalid(field + "不能为空")
	}
	ip, netw, err := net.ParseCIDR(raw)
	if err != nil || ip == nil || netw == nil || ip.To4() != nil {
		return errInvalid(field + "格式错误，应为如 fd00:8::/64")
	}
	return nil
}

func validateKeepalive(raw string) error {
	parts := strings.Fields(raw)
	if len(parts) != 2 {
		return errInvalid("保活参数必须为两个整数，如 10 120")
	}
	for _, part := range parts {
		n, err := strconv.Atoi(part)
		if err != nil || n <= 0 {
			return errInvalid("保活参数必须为两个大于 0 的整数，如 10 120")
		}
	}
	return nil
}

func validateIPAddress(raw, field string) error {
	if raw == "" {
		return nil
	}
	if net.ParseIP(raw) == nil {
		return errInvalid(field + "必须是合法的 IPv4 或 IPv6 地址")
	}
	return nil
}

func validatePushRoutes(raw string) error {
	if raw == "" {
		return nil
	}
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "push ") {
			continue
		}
		if _, _, err := net.ParseCIDR(line); err != nil {
			return errInvalid("推送路由每行必须是 CIDR，如 192.168.10.0/24")
		}
	}
	return nil
}

func errInvalid(msg string) error {
	return &paramError{msg: msg}
}

type paramError struct {
	msg string
}

func (e *paramError) Error() string { return e.msg }

// ApplyOpenVPNParams 将当前保存的参数应用到 server.conf 文件
func ApplyOpenVPNParams(db *gorm.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var sc model.ServerConfig
		err := db.Where("key = ?", model.OVPNParamsKey).First(&sc).Error
		p := model.DefaultOpenVPNParams()
		if err == nil && sc.Value != "" {
			_ = json.Unmarshal([]byte(sc.Value), &p)
		}
		content := vpn.BuildServerConfigFromParams(
			cfg.VPNBasePath, p.Port, p.Protocol, p.Device, p.Topology, p.MaxConnections, p.Subnet, p.PushRoutes, p.Management,
			p.PushDNS1, p.PushDNS2, p.Keepalive, p.Cipher, p.Auth, p.RunUser, p.RunGroup, p.Verb,
			p.VPNGateway, p.ClientToClient, p.IPv6, p.PersistKey, p.PersistTun, p.ExplicitExitNotify, p.IPv6Subnet,
		)
		if err := vpn.WriteServerConfig(cfg.VPNBasePath, content); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "已应用到 OpenVPN 配置文件，重启 OpenVPN 服务后生效"})
	}
}

// GetUsageStats 获取使用统计（从 OpenVPN status 文件解析当前连接）
func GetUsageStats(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := cfg.StatusFilePath
		if path == "" {
			path = filepath.Join(cfg.VPNBasePath, "openvpn-status.log")
		}
		list, err := vpn.ParseStatusFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				c.JSON(http.StatusOK, gin.H{"list": []interface{}{}, "message": "status 文件不存在，请确认 OpenVPN 已配置 status 并运行"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "读取 status 失败: " + err.Error()})
			return
		}
		// 在线统计：仅统计当前在线且登录成功的用户（忽略 CommonName 为空或 UNDEF 的记录）
		var filtered []vpn.ClientUsage
		for _, u := range list {
			name := strings.TrimSpace(u.CommonName)
			if name == "" || strings.EqualFold(name, "UNDEF") {
				continue
			}
			filtered = append(filtered, u)
		}
		c.JSON(http.StatusOK, gin.H{"list": filtered})
	}
}
