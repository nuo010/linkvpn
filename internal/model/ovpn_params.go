package model

// OpenVPNParams 管理面板中可配置的 OpenVPN 参数（与参考界面一致）
type OpenVPNParams struct {
	Port              string `json:"port"`                 // 端口，如 1194
	Protocol          string `json:"protocol"`             // udp / tcp
	MaxConnections    string `json:"max_connections"`      // 最大连接数，如 200
	VPNGateway        bool   `json:"vpn_gateway"`          // 是否启用 VPN 网关（推送默认路由）
	Subnet            string `json:"subnet"`               // 子网，如 10.8.8.0/24 或 10.8.8.0
	Management        string `json:"management"`           // 管理接口地址，如 127.0.0.1:7505
	IPv6              bool   `json:"ipv6"`                 // 是否启用 IPv6
	IPv6Subnet        string `json:"ipv6_subnet"`          // IPv6 子网
	PushDNS1          string `json:"push_dns1"`            // 推送 DNS1，如 8.8.8.8
	PushDNS2          string `json:"push_dns2"`            // 推送 DNS2，如 2001:4860:4860::8888
	AutoApplyToConfig bool   `json:"auto_apply_to_config"` // 是否自动将参数应用到 server.conf（保存参数时一并写入）
}

func DefaultOpenVPNParams() OpenVPNParams {
	return OpenVPNParams{
		Port:           "1194",
		Protocol:       "udp",
		MaxConnections: "200",
		VPNGateway:     false,
		Subnet:         "10.8.8.0/24",
		Management:     "127.0.0.1:7505",
		IPv6:           false,
		IPv6Subnet:     "fd00:8::/64",
		PushDNS1:       "8.8.8.8",
		PushDNS2:       "2001:4860:4860::8888",
	}
}

const OVPNParamsKey = "openvpn_params" // 存 ServerConfig 时的 key，值为 JSON
