package config

import (
	"os"
	"path/filepath"
	"strconv"
)

type Config struct {
	HTTPPort       int
	VPNBasePath    string // OpenVPN 与 easy-rsa 工作目录
	StatusFilePath string // OpenVPN status 文件路径
	LogFilePath    string // OpenVPN 主日志路径，如 /var/log/openvpn.log
	StaticDir      string
	ServerName     string
	JWTSecret      string
	AdminUser      string
	AdminPass      string
	DatabasePath   string
}

func Load() *Config {
	vpnBase := os.Getenv("VPN_BASE_PATH")
	if vpnBase == "" {
		vpnBase = "/etc/openvpn"
	}
	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		dbPath = filepath.Join(vpnBase, "data", "panel.db")
	}
	port := 8789
	if p := os.Getenv("HTTP_PORT"); p != "" {
		if v, err := strconv.Atoi(p); err == nil {
			port = v
		}
	}
	statusFile := os.Getenv("STATUS_FILE")
	if statusFile == "" {
		statusFile = filepath.Join(vpnBase, "openvpn-status.log")
	}
	logFile := os.Getenv("LOG_FILE")
	if logFile == "" {
		logFile = filepath.Join(vpnBase, "openvpn.log")
	}
	return &Config{
		HTTPPort:       port,
		VPNBasePath:    vpnBase,
		StatusFilePath: statusFile,
		LogFilePath:    logFile,
		StaticDir:      os.Getenv("STATIC_DIR"),
		ServerName:     getEnv("SERVER_NAME", "OpenVPN-Server"),
		JWTSecret:      getEnv("JWT_SECRET", "change-me-in-production"),
		AdminUser:      getEnv("ADMIN_USER", "admin"),
		AdminPass:      getEnv("ADMIN_PASS", "admin"),
		DatabasePath:   dbPath,
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
