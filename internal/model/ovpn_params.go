package model

// OpenVPNParams 管理面板中可配置的 OpenVPN 参数（与参考界面一致）
type OpenVPNParams struct {
	Port               string `json:"port"`                 // 端口，如 1194
	Protocol           string `json:"protocol"`             // udp / tcp
	Device             string `json:"device"`               // tun / tap
	Topology           string `json:"topology"`             // subnet / net30
	MaxConnections     string `json:"max_connections"`      // 最大连接数，如 200
	VPNGateway         bool   `json:"vpn_gateway"`          // 是否启用 VPN 网关（推送默认路由）
	ClientToClient     bool   `json:"client_to_client"`     // 是否允许客户端互访
	Subnet             string `json:"subnet"`               // 子网，如 10.8.8.0/24 或 10.8.8.0
	PushRoutes         string `json:"push_routes"`          // 推送路由，按行填写 CIDR，如 192.168.10.0/24
	Management         string `json:"management"`           // 管理接口地址，如 127.0.0.1:7505
	IPv6               bool   `json:"ipv6"`                 // 是否启用 IPv6
	IPv6Subnet         string `json:"ipv6_subnet"`          // IPv6 子网
	PushDNS1           string `json:"push_dns1"`            // 推送 DNS1，如 8.8.8.8
	PushDNS2           string `json:"push_dns2"`            // 推送 DNS2，如 2001:4860:4860::8888
	Keepalive          string `json:"keepalive"`            // keepalive，例如 10 120
	Cipher             string `json:"cipher"`               // 加密算法
	Auth               string `json:"auth"`                 // 摘要算法
	RunUser            string `json:"run_user"`             // 运行用户
	RunGroup           string `json:"run_group"`            // 运行组
	PersistKey         bool   `json:"persist_key"`          // persist-key
	PersistTun         bool   `json:"persist_tun"`          // persist-tun
	Verb               string `json:"verb"`                 // 日志级别
	ExplicitExitNotify bool   `json:"explicit_exit_notify"` // UDP 场景常用
	AutoApplyToConfig  bool   `json:"auto_apply_to_config"` // 是否自动将参数应用到 server.conf（保存参数时一并写入）
}

func DefaultOpenVPNParams() OpenVPNParams {
	return OpenVPNParams{
		Port:               "1194",
		Protocol:           "udp",
		Device:             "tun",
		Topology:           "subnet",
		MaxConnections:     "200",
		VPNGateway:         false,
		ClientToClient:     true,
		Subnet:             "10.8.8.0/24",
		PushRoutes:         "",
		Management:         "",
		IPv6:               false,
		IPv6Subnet:         "fd00:8::/64",
		PushDNS1:           "8.8.8.8",
		PushDNS2:           "2001:4860:4860::8888",
		Keepalive:          "10 120",
		Cipher:             "AES-256-GCM",
		Auth:               "SHA256",
		RunUser:            "nobody",
		RunGroup:           "nogroup",
		PersistKey:         true,
		PersistTun:         true,
		Verb:               "3",
		ExplicitExitNotify: true,
	}
}

const OVPNParamsKey = "openvpn_params" // 存 ServerConfig 时的 key，值为 JSON
