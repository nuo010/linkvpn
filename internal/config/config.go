package config

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Config struct {
	HTTPPort       int    `json:"http_port" yaml:"http_port"`
	VPNBasePath    string `json:"vpn_base_path" yaml:"vpn_base_path"`       // OpenVPN 与 easy-rsa 工作目录
	StatusFilePath string `json:"status_file_path" yaml:"status_file_path"` // OpenVPN status 文件路径
	LogFilePath    string `json:"log_file_path" yaml:"log_file_path"`       // OpenVPN 主日志路径，如 /var/log/openvpn.log
	StaticDir      string `json:"static_dir" yaml:"static_dir"`
	ServerName     string `json:"server_name" yaml:"server_name"`
	JWTSecret      string `json:"jwt_secret" yaml:"jwt_secret"`
	AdminUser      string `json:"admin_user" yaml:"admin_user"`
	AdminPass      string `json:"admin_pass" yaml:"admin_pass"`
	DatabasePath   string `json:"database_path" yaml:"database_path"`
}

func Load() *Config {
	cfg := &Config{
		HTTPPort:    8789,
		VPNBasePath: "/etc/openvpn",
		ServerName:  "OpenVPN-Server",
		JWTSecret:   "change-me-in-production",
		AdminUser:   "admin",
		AdminPass:   "admin",
		StaticDir:   "/app/web",
	}

	loadFromDefaultConfigFiles(cfg)
	applyEnvOverrides(cfg)
	fillDerivedDefaults(cfg)

	return cfg
}

func loadFromDefaultConfigFiles(cfg *Config) {
	candidates := []string{}
	if explicit := strings.TrimSpace(os.Getenv("CONFIG_FILE")); explicit != "" {
		candidates = append(candidates, explicit)
	} else {
		candidates = append(candidates,
			"/etc/openvpn/config.yaml",
			"/etc/openvpn/config.yml",
			"/etc/openvpn/panel.yaml",
			"/etc/openvpn/panel.yml",
		)
	}

	for _, path := range candidates {
		if strings.TrimSpace(path) == "" {
			continue
		}
		if err := loadFromFile(path, cfg); err == nil {
			return
		}
	}
}

func applyEnvOverrides(cfg *Config) {
	if v := strings.TrimSpace(os.Getenv("VPN_BASE_PATH")); v != "" {
		cfg.VPNBasePath = v
	}
	if v := strings.TrimSpace(os.Getenv("STATIC_DIR")); v != "" {
		cfg.StaticDir = v
	}
	if v := strings.TrimSpace(os.Getenv("STATUS_FILE_PATH")); v != "" {
		cfg.StatusFilePath = v
	}
	if v := strings.TrimSpace(os.Getenv("LOG_FILE_PATH")); v != "" {
		cfg.LogFilePath = v
	}
	if v := strings.TrimSpace(os.Getenv("SERVER_NAME")); v != "" {
		cfg.ServerName = v
	}
	if v := strings.TrimSpace(os.Getenv("JWT_SECRET")); v != "" {
		cfg.JWTSecret = v
	}
	if v := strings.TrimSpace(os.Getenv("ADMIN_USER")); v != "" {
		cfg.AdminUser = v
	}
	if v := strings.TrimSpace(os.Getenv("ADMIN_PASS")); v != "" {
		cfg.AdminPass = v
	}
	if v := strings.TrimSpace(os.Getenv("DATABASE_PATH")); v != "" {
		cfg.DatabasePath = v
	}
	if v := strings.TrimSpace(os.Getenv("HTTP_PORT")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.HTTPPort = n
		}
	}
}

func fillDerivedDefaults(cfg *Config) {
	base := strings.TrimSpace(cfg.VPNBasePath)
	if base == "" {
		base = "/etc/openvpn"
		cfg.VPNBasePath = base
	}
	if strings.TrimSpace(cfg.StatusFilePath) == "" {
		cfg.StatusFilePath = filepath.Join(base, "openvpn-status.log")
	}
	if strings.TrimSpace(cfg.LogFilePath) == "" {
		cfg.LogFilePath = filepath.Join(base, "openvpn.log")
	}
	if strings.TrimSpace(cfg.DatabasePath) == "" {
		cfg.DatabasePath = filepath.Join(base, "panel.db")
	}
}
