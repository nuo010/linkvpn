package config

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
	// 1) 默认值
	cfg := &Config{
		HTTPPort:    8789,
		VPNBasePath: "/etc/openvpn",
		ServerName:  "OpenVPN-Server",
		JWTSecret:   "change-me-in-production",
		AdminUser:   "admin",
		AdminPass:   "admin",
	}
	_ = loadFromFile("/etc/openvpn/config.yaml", cfg)
	return cfg
}
