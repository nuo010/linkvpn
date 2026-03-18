package handler

import (
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"singleOpenVpn/internal/config"
	"singleOpenVpn/internal/model"
	"singleOpenVpn/internal/startup"
	"singleOpenVpn/internal/vpn"
)

// HomeStats 首页仪表盘数据
type HomeStats struct {
	UptimeSeconds  int64             `json:"uptime_seconds"`
	UserCount      int64             `json:"user_count"`
	OnlineCount    int               `json:"online_count"`
	OpenVPNRunning bool              `json:"openvpn_running"`
	OpenVPNPID     int               `json:"openvpn_pid"`
	OpenVPNVersion string            `json:"openvpn_version"`
	TotalBytesRecv int64             `json:"total_bytes_recv"`
	TotalBytesSent int64             `json:"total_bytes_sent"`
	Top10Upload    []UserTrafficItem `json:"top10_upload"`
	Top10Download  []UserTrafficItem `json:"top10_download"`
	PKIInitialized bool              `json:"pki_initialized"`
	BasePath       string            `json:"base_path"`
}

// UserTrafficItem 用户流量项（Top10）
type UserTrafficItem struct {
	Username  string `json:"username"`
	BytesRecv int64  `json:"bytes_recv"`
	BytesSent int64  `json:"bytes_sent"`
}

// GetHome 首页仪表盘：运行时长、用户/在线统计、总流量、Top10 流量等
func GetHome(db *gorm.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		stats := HomeStats{}
		if !startup.StartTime.IsZero() {
			stats.UptimeSeconds = int64(time.Since(startup.StartTime).Seconds())
			if stats.UptimeSeconds < 0 {
				stats.UptimeSeconds = 0
			}
		}

		var userCount int64
		_ = db.Model(&model.VPNUser{}).Count(&userCount).Error
		stats.UserCount = userCount

		var usageList []vpn.ClientUsage
		if path := cfg.StatusFilePath; path != "" {
			if list, err := vpn.ParseStatusFile(path); err == nil {
				// 仅统计“登录成功”的连接：过滤掉 CommonName 为空或为 UNDEF 的记录
				for _, u := range list {
					name := strings.TrimSpace(u.CommonName)
					if name == "" || strings.EqualFold(name, "UNDEF") {
						continue
					}
					usageList = append(usageList, u)
				}
				stats.OnlineCount = len(usageList)
				for _, u := range usageList {
					stats.TotalBytesRecv += u.BytesRecv
					stats.TotalBytesSent += u.BytesSent
				}
				// Top10 上传（按 bytes_recv 降序）
				sort.Slice(usageList, func(i, j int) bool { return usageList[i].BytesRecv > usageList[j].BytesRecv })
				for i := 0; i < len(usageList) && i < 10; i++ {
					stats.Top10Upload = append(stats.Top10Upload, UserTrafficItem{
						Username:  usageList[i].CommonName,
						BytesRecv: usageList[i].BytesRecv,
						BytesSent: usageList[i].BytesSent,
					})
				}
				// Top10 下载（按 bytes_sent 降序）
				sort.Slice(usageList, func(i, j int) bool { return usageList[i].BytesSent > usageList[j].BytesSent })
				for i := 0; i < len(usageList) && i < 10; i++ {
					stats.Top10Download = append(stats.Top10Download, UserTrafficItem{
						Username:  usageList[i].CommonName,
						BytesRecv: usageList[i].BytesRecv,
						BytesSent: usageList[i].BytesSent,
					})
				}
			}
		}

		// PKI 与工作目录
		if _, err := os.Stat(filepath.Join(cfg.VPNBasePath, "easy-rsa", "pki", "ca.crt")); err == nil {
			stats.PKIInitialized = true
		}
		stats.BasePath = cfg.VPNBasePath

		// OpenVPN 进程状态与 PID（仅在 Linux 环境有效）
		if pid, ok := vpn.FindOpenVPNPID(); ok {
			stats.OpenVPNRunning = true
			stats.OpenVPNPID = pid
		}
		stats.OpenVPNVersion = vpn.GetOpenVPNVersion()

		c.JSON(http.StatusOK, stats)
	}
}
