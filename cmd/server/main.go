package main

import (
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"singleOpenVpn/internal/config"
	"singleOpenVpn/internal/logwatcher"
	"singleOpenVpn/internal/router"
	"singleOpenVpn/internal/startup"
	"singleOpenVpn/internal/store"
	"singleOpenVpn/internal/vpn"
)

func main() {
	startup.StartTime = time.Now()
	cfg := config.Load()
	if err := ensureVPNBaseDir(cfg.VPNBasePath); err != nil {
		log.Fatalf("创建 VPN 目录失败: %v", err)
	}
	if err := vpn.RepairCCDFilesTrailingNewline(cfg.VPNBasePath); err != nil {
		log.Printf("修复 CCD 文件末尾换行时出错: %v", err)
	}

	// 挂载目录为空（无 CA）时自动初始化 PKI 并启动 OpenVPN，无需手动点击「初始化 PKI」
	if vpn.IsPKIEmpty(cfg.VPNBasePath) {
		log.Print("检测到挂载目录未初始化 PKI，正在自动创建 CA 与服务端证书…")
		if err := vpn.EnsurePKI(cfg.VPNBasePath, cfg.ServerName); err != nil {
			log.Printf("自动初始化 PKI 失败: %v（可稍后在面板手动点击「初始化 PKI」）", err)
		} else {
			log.Print("PKI 已自动初始化完成")
			if err := vpn.StartOpenVPN(cfg.VPNBasePath); err != nil {
				log.Printf("自动启动 OpenVPN 失败: %v", err)
			} else {
				log.Print("OpenVPN 已在后台启动")
			}
		}
	}

	db, err := store.NewDB(cfg)
	if err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}

	// 后台持续同步 openvpn.log 到连接记录表
	logwatcher.StartOpenVPNLogWatcher(db, cfg, 3*time.Second)

	r := router.Setup(db, cfg)
	addr := ":" + strconv.Itoa(cfg.HTTPPort)
	log.Printf("OpenVPN 管理面板启动: http://0.0.0.0%s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}

func ensureVPNBaseDir(base string) error {
	dirs := []string{
		base,
		filepath.Join(base, "data"),
		filepath.Join(base, "client-configs"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return err
		}
	}
	return nil
}
