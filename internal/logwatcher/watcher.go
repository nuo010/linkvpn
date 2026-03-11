package logwatcher

import (
	"log"
	"singleOpenVpn/internal/vpn"
	"time"

	"gorm.io/gorm"

	"singleOpenVpn/internal/config"
	"singleOpenVpn/internal/handler"
)

// StartOpenVPNLogWatcher 后台轮询 openvpn.log，将新产生的连接事件持续写入数据库。
// 简化实现：每隔 interval 调用一次 SyncVPNLogsFromFile，依靠去重逻辑避免重复记录。
func StartOpenVPNLogWatcher(db *gorm.DB, cfg *config.Config, interval time.Duration) {
	if cfg == nil || cfg.LogFilePath == "" {
		return
	}
	if interval <= 0 {
		interval = 3 * time.Second
	}

	log.Printf("logwatcher: 启动 openvpn.log 后台同步 (interval=%s, path=%s)", interval, cfg.LogFilePath)

	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			func() {
				defer func() {
					if r := recover(); r != nil {
						log.Printf("logwatcher: panic recovered: %v", r)
					}
				}()
				// 从 openvpn.log 中同步成功/失败/TLS 错误等记录
				handler.SyncVPNLogsFromFile(db, cfg.LogFilePath)

				// 可选：从 status 文件中同步成功在线的连接与下线时间
				if cfg.StatusFilePath != "" {
					if list, err := vpn.ParseStatusFile(cfg.StatusFilePath); err == nil {
						handler.SyncVPNStatusList(db, list)
					}
				}
			}()
		}
	}()
}
