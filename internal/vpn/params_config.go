package vpn

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

// BuildServerConfigFromParams 根据面板参数生成 server.conf 内容。
// 所有写入的路径均为绝对路径（ca/cert/key/dh、client-config-dir、status、log-append、ifconfig-pool-persist）。
// vpnBase 为 VPN 工作目录（如 /etc/openvpn），空则 PKI/client-config-dir 等使用 /etc/openvpn，status/log 使用 /var/log。
// subnet 可为 "10.8.8.0/24" 或 "10.8.8.0"，将转为 server x.x.x.x 255.255.255.0
func BuildServerConfigFromParams(vpnBase, port, proto, maxClients, subnet, management, pushDNS1, pushDNS2 string, vpnGateway, ipv6 bool, ipv6Subnet string) string {
	portNum := 1194
	if p, _ := strconv.Atoi(port); p > 0 {
		portNum = p
	}
	if proto == "" {
		proto = "udp"
	}
	serverNet := parseSubnetToServer(subnet) // "10.8.8.0"
	if serverNet == "" {
		serverNet = ServerVPNCIDR
	}

	var b strings.Builder
	b.WriteString("# OpenVPN 服务端配置（由参数配置生成）\n")
	b.WriteString("# 修改后需重启 OpenVPN 服务生效\n\n")
	b.WriteString(fmt.Sprintf("port %d\n", portNum))
	b.WriteString(fmt.Sprintf("proto %s\n", proto))
	b.WriteString("dev tun\n\n")
	vpnBasePKI := vpnBase
	if vpnBase == "" {
		vpnBasePKI = "/etc/openvpn"
	}
	b.WriteString("ca " + filepath.Join(vpnBasePKI, "easy-rsa", "pki", "ca.crt") + "\n")
	b.WriteString("cert " + filepath.Join(vpnBasePKI, "easy-rsa", "pki", "issued", "server.crt") + "\n")
	b.WriteString("key " + filepath.Join(vpnBasePKI, "easy-rsa", "pki", "private", "server.key") + "\n")
	b.WriteString("dh " + filepath.Join(vpnBasePKI, "easy-rsa", "pki", "dh.pem") + "\n\n")
	b.WriteString("topology subnet\n")
	b.WriteString(fmt.Sprintf("server %s 255.255.255.0\n", serverNet))
	if maxClients != "" {
		if n, _ := strconv.Atoi(maxClients); n > 0 {
			b.WriteString(fmt.Sprintf("max-clients %d\n", n))
		}
	}
	b.WriteString("ifconfig-pool-persist " + filepath.Join(vpnBasePKI, "easy-rsa", "pki", "ipp.txt") + "\n")
	// client-config-dir 始终使用绝对路径，避免 OpenVPN 工作目录影响
	if vpnBase != "" {
		b.WriteString("client-config-dir " + filepath.Join(vpnBase, CCDDir) + "\n")
	} else {
		b.WriteString("client-config-dir " + filepath.Join("/etc/openvpn", CCDDir) + "\n")
	}
	b.WriteString("client-to-client\n\n")
	if management != "" {
		b.WriteString(fmt.Sprintf("management %s\n", management))
	}
	b.WriteString("keepalive 10 120\n")
	b.WriteString("cipher AES-256-GCM\n")
	b.WriteString("auth SHA256\n")
	b.WriteString("user nobody\n")
	b.WriteString("group nogroup\n")
	b.WriteString("persist-key\n")
	b.WriteString("persist-tun\n")
	statusPath := "/var/log/openvpn-status.log"
	logPath := "/var/log/openvpn.log"
	if vpnBase != "" {
		statusPath = filepath.Join(vpnBase, "openvpn-status.log")
		logPath = filepath.Join(vpnBase, "openvpn.log")
	}
	b.WriteString(fmt.Sprintf("status %s 10\n", statusPath))
	b.WriteString(fmt.Sprintf("log-append %s\n", logPath))
	b.WriteString("verb 3\n")
	b.WriteString("explicit-exit-notify 1\n")
	if pushDNS1 != "" {
		b.WriteString(fmt.Sprintf("push \"dhcp-option DNS %s\"\n", pushDNS1))
	}
	if pushDNS2 != "" {
		b.WriteString(fmt.Sprintf("push \"dhcp-option DNS %s\"\n", pushDNS2))
	}
	if vpnGateway {
		b.WriteString("push \"redirect-gateway def1 bypass-dhcp\"\n")
	}
	if ipv6 && ipv6Subnet != "" {
		b.WriteString(fmt.Sprintf("server-ipv6 %s\n", ipv6Subnet))
	}
	return b.String()
}

// parseSubnetToServer 从 "10.8.8.0/24" 或 "10.8.8.0" 得到 "10.8.8.0"
func parseSubnetToServer(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if idx := strings.Index(s, "/"); idx > 0 {
		return strings.TrimSpace(s[:idx])
	}
	return s
}
