package handler

import (
	"bufio"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"singleOpenVpn/internal/config"
	"singleOpenVpn/internal/model"
	"singleOpenVpn/internal/vpn"
)

// GetLoginLogs 获取登录日志，支持按日期筛选
func GetLoginLogs(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		dateStr := c.Query("date")
		q := db.Model(&model.LoginLog{}).Order("id DESC").Limit(500)
		if dateStr != "" {
			t, err := time.ParseInLocation("2006-01-02", dateStr, model.Shanghai)
			if err == nil {
				start := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, model.Shanghai)
				end := start.Add(24 * time.Hour)
				q = q.Where("created_at >= ? AND created_at < ?", model.NT(start).String(), model.NT(end).String())
			}
		}
		var list []model.LoginLog
		if err := q.Find(&list).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"list": list})
	}
}

// ClearLoginLogs 清空面板登录日志
func ClearLoginLogs(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := db.Where("1 = 1").Delete(&model.LoginLog{}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "已清空"})
	}
}

// GetVPNConnectionLogs 获取 VPN 用户连接记录，支持按日期筛选；请求时会先根据 status 同步一次连接状态
func GetVPNConnectionLogs(db *gorm.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		dateStr := c.Query("date")
		baseQ := db.Model(&model.VPNConnectionLog{}).Order("connected_at DESC")
		if dateStr != "" {
			t, err := time.ParseInLocation("2006-01-02", dateStr, model.Shanghai)
			if err == nil {
				start := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, model.Shanghai)
				end := start.Add(24 * time.Hour)
				baseQ = baseQ.Where("connected_at >= ? AND connected_at < ?", model.NT(start).String(), model.NT(end).String())
			}
		}
		// 分页
		page := 1
		pageSize := 20
		if p := c.Query("page"); p != "" {
			if v, err := strconv.Atoi(p); err == nil && v > 0 {
				page = v
			}
		}
		if ps := c.Query("page_size"); ps != "" {
			if v, err := strconv.Atoi(ps); err == nil && v > 0 && v <= 100 {
				pageSize = v
			}
		}
		var total int64
		if err := baseQ.Count(&total).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		offset := (page - 1) * pageSize
		var list []model.VPNConnectionLog
		if err := baseQ.Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"list": list, "total": total})
	}
}

// ClearVPNConnectionLogs 清空 VPN 连接记录
// 同时会截断 openvpn.log，避免再次从历史日志中重新同步相同记录。
func ClearVPNConnectionLogs(db *gorm.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := db.Where("1 = 1").Delete(&model.VPNConnectionLog{}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		// 截断 openvpn.log（若路径存在），仅影响连接记录的数据来源，不影响当前运行的进程。
		if cfg != nil && cfg.LogFilePath != "" {
			_ = os.Truncate(cfg.LogFilePath, 0)
		}
		c.JSON(http.StatusOK, gin.H{"message": "已清空"})
	}
}

// GetVPNLogFile 返回 OpenVPN 日志文件内容（最后 N 行），用于面板实时查看
// name: openvpn.log | openvpn-status.log；lines: 默认 500
func GetVPNLogFile(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Query("name")
		if name != "openvpn.log" && name != "openvpn-status.log" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "仅支持 name=openvpn.log 或 openvpn-status.log"})
			return
		}
		path := cfg.LogFilePath
		if name == "openvpn-status.log" {
			path = cfg.StatusFilePath
		}
		lines := 500
		if l := c.Query("lines"); l != "" {
			if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 5000 {
				lines = n
			}
		}
		content, err := readLastLines(path, lines)
		if err != nil {
			if os.IsNotExist(err) {
				c.JSON(http.StatusOK, gin.H{"content": "", "path": path, "message": "文件不存在或尚未生成"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "读取失败: " + err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"content": content, "path": path})
	}
}

func readLastLines(path string, n int) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	var queue []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		queue = append(queue, sc.Text())
		if len(queue) > n {
			queue = queue[1:]
		}
	}
	if err := sc.Err(); err != nil {
		return "", err
	}
	return strings.Join(queue, "\n"), nil
}

// SyncVPNLogsFromFile 解析 openvpn.log 中最近的认证/握手/成功记录，同步到 vpn_connection_logs 中。
// 仅作为补充信息：status 文件只包含已建立的数据通道连接，账号/密码/证书等失败只会出现在主日志里。
func SyncVPNLogsFromFile(db *gorm.DB, logPath string) {
	content, err := readLastLines(logPath, 1000)
	if err != nil || content == "" {
		return
	}
	lines := strings.Split(content, "\n")
	// 行内未解析到时间时用当前东八区时刻，与 ParseOpenVPNLogTime 一致
	now := time.Now().In(model.Shanghai)
	for _, line := range lines {
		// 1) 账号/密码错误：AUTH_FAILED
		if strings.Contains(line, "AUTH_FAILED") {
			// 行首时间为 openvpn.log 本地墙钟，必须用 ParseInLocation(Asia/Shanghai)，避免与 created_at 差 8 小时
			ts := now
			if t, ok := model.ParseOpenVPNLogTime(line); ok {
				ts = t
			}
			username := "UNDEF"
			// 行中一般包含类似 "[username] AUTH_FAILED" 的片段，优先从中取用户名。
			if i := strings.Index(line, "["); i >= 0 {
				if j := strings.Index(line[i+1:], "]"); j > 0 {
					username = strings.TrimSpace(line[i+1 : i+1+j])
					if username == "" {
						username = "UNDEF"
					}
				}
			}
			// 若日志中仅标记为 UNDEF，则这类记录信息价值较低且容易与后续同一连接的有效用户名重复，这里跳过，
			// 只保留带具体用户名的 AUTH_FAILED 记录。
			if strings.EqualFold(username, "UNDEF") {
				continue
			}
			// 尝试从 "x.x.x.x:port" 里提取来源 IP（若失败则留空）。
			realIP := ""
			parts := strings.Fields(line)
			for _, p := range parts {
				if strings.Count(p, ".") == 3 && strings.Contains(p, ":") {
					hostPort := strings.TrimSpace(p)
					if idx := strings.LastIndex(hostPort, ":"); idx > 0 {
						realIP = hostPort[:idx]
					}
					break
				}
			}
			// 避免重复写入：按用户名 + 时间戳 + status=failed 做一次存在性检查。
			var count int64
			_ = db.Model(&model.VPNConnectionLog{}).
				Where("username = ? AND status = ? AND connected_at = ?", username, "failed", model.NT(ts).String()).
				Count(&count)
			if count > 0 {
				continue
			}
			_ = db.Create(&model.VPNConnectionLog{
				Username:    username,
				Status:      "failed",
				RealIP:      realIP,
				ConnectedAt: model.NT(ts),
			}).Error
			continue
		}

		// 2) TLS/证书/握手错误：TLS Error / TLS handshake failed
		if strings.Contains(line, "TLS Error: TLS key negotiation failed") || strings.Contains(line, "TLS Error: TLS handshake failed") {

			ts := now
			if t, ok := model.ParseOpenVPNLogTime(line); ok {
				ts = t
			}

			// 用户名未知，统一标记为“未知”
			username := "未知"

			realIP := ""
			parts := strings.Fields(line)
			for _, p := range parts {
				if strings.Count(p, ".") == 3 && strings.Contains(p, ":") {
					hostPort := strings.TrimSpace(p)
					if idx := strings.LastIndex(hostPort, ":"); idx > 0 {
						realIP = hostPort[:idx]
					}
					break
				}
			}

			var n int64
			_ = db.Model(&model.VPNConnectionLog{}).
				Where("username = ? AND status = ? AND connected_at = ? AND real_ip = ?", username, "tls_error", model.NT(ts).String(), realIP).
				Count(&n)
			if n > 0 {
				continue
			}
			_ = db.Create(&model.VPNConnectionLog{
				Username:    username,
				Status:      "tls_error",
				RealIP:      realIP,
				ConnectedAt: model.NT(ts),
			}).Error
			continue
		}

		// 3) 连接成功：MULTI: Learn（数据通道已建立、已分配虚拟 IP）
		// 示例：2026-03-11 09:59:52 MULTI: Learn: 10.8.8.44 -> 666666/58.56.20.70:56712
		//       虚拟IP -> 用户名/来源IP:端口
		if strings.Contains(line, "MULTI: Learn") {
			ts := now
			if t, ok := model.ParseOpenVPNLogTime(line); ok {
				ts = t
			}

			username := "未知"
			realIP := ""
			virtualIP := ""

			// 取 "MULTI: Learn:" 之后的部分
			learnIdx := strings.Index(line, "MULTI: Learn:")
			if learnIdx >= 0 {
				payload := strings.TrimSpace(line[learnIdx+len("MULTI: Learn:"):])
				// payload 形如: 10.8.8.44 -> 666666/58.56.20.70:56712
				arrow := strings.Index(payload, " -> ")
				if arrow > 0 {
					virtualIP = strings.TrimSpace(payload[:arrow])
					right := strings.TrimSpace(payload[arrow+len(" -> "):])
					// right 形如: 666666/58.56.20.70:56712
					slash := strings.Index(right, "/")
					if slash > 0 {
						username = strings.TrimSpace(right[:slash])
						hostPort := strings.TrimSpace(right[slash+1:])
						if colon := strings.Index(hostPort, ":"); colon > 0 {
							realIP = hostPort[:colon]
						} else if hostPort != "" {
							realIP = hostPort
						}
					}
				}
			}
			if username == "" {
				username = "未知"
			}

			var m int64
			q := db.Model(&model.VPNConnectionLog{}).
				Where("username = ? AND status = ? AND connected_at = ?", username, "success", model.NT(ts).String())
			if realIP != "" {
				q = q.Where("real_ip = ?", realIP)
			}
			_ = q.Count(&m)
			if m > 0 {
				continue
			}
			_ = db.Create(&model.VPNConnectionLog{
				Username:    username,
				Status:      "success",
				RealIP:      realIP,
				VirtualIP:   virtualIP,
				ConnectedAt: model.NT(ts),
			}).Error
		}
	}
}

// SyncVPNStatusList 根据当前 status 连接列表：
// 1) 不在列表里的已打开记录补 disconnected_at（下线时间）
// 2) 在列表里的仅更新已有「未断开」记录的流量/IP 等，不新建连接记录（新建由 openvpn.log 等逻辑负责）
func SyncVPNStatusList(db *gorm.DB, list []vpn.ClientUsage) {
	currentByCN := make(map[string]vpn.ClientUsage)
	for _, u := range list {
		name := strings.TrimSpace(u.CommonName)
		if name == "" || strings.EqualFold(name, "UNDEF") {
			continue
		}
		currentByCN[name] = u
	}

	var open []model.VPNConnectionLog
	if err := db.Where("disconnected_at IS NULL").Find(&open).Error; err != nil {
		return
	}
	for i := range open {
		if _, in := currentByCN[open[i].Username]; !in {
			// UpdateColumn 写入纯字符串，避免驱动按 UTC 写成 15:xx 与 connected_at 23:xx 颠倒
			_ = db.Model(&open[i]).UpdateColumn("disconnected_at", model.NowNaive().String())
		}
	}

	for _, u := range list {
		name := strings.TrimSpace(u.CommonName)
		if name == "" || strings.EqualFold(name, "UNDEF") {
			continue
		}
		var existing model.VPNConnectionLog
		err := db.Where("username = ? AND disconnected_at IS NULL", name).First(&existing).Error
		if err != nil {
			// 没有已存在的未断开记录则跳过，不新建（避免 status 短暂残留误插 success）
			continue
		}
		_ = db.Model(&existing).Updates(map[string]interface{}{
			"bytes_recv": u.BytesRecv,
			"bytes_sent": u.BytesSent,
			"real_ip":    u.RealIP,
			"virtual_ip": u.VirtualIP,
			"status":     "success",
		}).Error
	}
}
