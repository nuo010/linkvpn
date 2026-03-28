package vpn

import (
	"fmt"
	"net"
	"path/filepath"
	"strconv"
	"strings"
)

// BuildServerConfigFromParams 根据面板参数生成 server.conf 内容。
// 所有写入的路径均为绝对路径（ca/cert/key/dh、client-config-dir、status、log-append、ifconfig-pool-persist）。
// vpnBase 为 VPN 工作目录（如 /etc/openvpn），空则 PKI/client-config-dir 等使用 /etc/openvpn，status/log 使用 /var/log。
// subnet 可为 "10.8.8.0/24" 或 "10.8.8.0"，将转为 server x.x.x.x 255.255.255.0
func BuildServerConfigFromParams(
	vpnBase, port, proto, device, topology, maxClients, subnet, pushRoutes, management,
	pushDNS1, pushDNS2, keepalive, cipherName, authName, runUser, runGroup, verb string,
	vpnGateway, clientToClient, ipv6, persistKey, persistTun, explicitExitNotify bool,
	ipv6Subnet string,
) string {
	portNum := 1194
	if p, _ := strconv.Atoi(port); p > 0 {
		portNum = p
	}
	proto = strings.ToLower(strings.TrimSpace(proto))
	if proto != "tcp" && proto != "udp" {
		proto = "udp"
	}
	serverNet := parseSubnetToServer(subnet) // "10.8.8.0"
	if serverNet == "" {
		serverNet = ServerVPNCIDR
	}
	device = strings.ToLower(strings.TrimSpace(device))
	if device != "tap" && device != "tun" {
		device = "tun"
	}
	topology = strings.ToLower(strings.TrimSpace(topology))
	if topology != "net30" && topology != "subnet" {
		topology = "subnet"
	}
	if strings.TrimSpace(keepalive) == "" {
		keepalive = "10 120"
	}
	if strings.TrimSpace(cipherName) == "" {
		cipherName = "AES-256-GCM"
	}
	if strings.TrimSpace(authName) == "" {
		authName = "SHA256"
	}
	if strings.TrimSpace(runUser) == "" {
		runUser = "nobody"
	}
	if strings.TrimSpace(runGroup) == "" {
		runGroup = "nogroup"
	}
	if strings.TrimSpace(verb) == "" {
		verb = "3"
	}

	var b strings.Builder
	b.WriteString("# OpenVPN 服务端配置（由参数配置生成）\n")
	b.WriteString("# 修改后需重启 OpenVPN 服务生效\n\n")
	b.WriteString(fmt.Sprintf("port %d\n", portNum))
	b.WriteString(fmt.Sprintf("proto %s\n", proto))
	b.WriteString(fmt.Sprintf("dev %s\n\n", device))
	vpnBasePKI := vpnBase
	if vpnBase == "" {
		vpnBasePKI = "/etc/openvpn"
	}
	b.WriteString("ca " + filepath.Join(vpnBasePKI, "easy-rsa", "pki", "ca.crt") + "\n")
	b.WriteString("cert " + filepath.Join(vpnBasePKI, "easy-rsa", "pki", "issued", "server.crt") + "\n")
	b.WriteString("key " + filepath.Join(vpnBasePKI, "easy-rsa", "pki", "private", "server.key") + "\n")
	b.WriteString("dh " + filepath.Join(vpnBasePKI, "easy-rsa", "pki", "dh.pem") + "\n\n")
	b.WriteString(fmt.Sprintf("topology %s\n", topology))
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
	if clientToClient {
		b.WriteString("client-to-client\n")
	}
	b.WriteString("\n")
	if mgmtLine := formatManagementDirective(management); mgmtLine != "" {
		b.WriteString(mgmtLine)
		b.WriteString("\n")
	}
	b.WriteString(fmt.Sprintf("keepalive %s\n", strings.TrimSpace(keepalive)))
	b.WriteString(fmt.Sprintf("cipher %s\n", strings.TrimSpace(cipherName)))
	b.WriteString(fmt.Sprintf("auth %s\n", strings.TrimSpace(authName)))
	b.WriteString(fmt.Sprintf("user %s\n", strings.TrimSpace(runUser)))
	b.WriteString(fmt.Sprintf("group %s\n", strings.TrimSpace(runGroup)))
	if persistKey {
		b.WriteString("persist-key\n")
	}
	if persistTun {
		b.WriteString("persist-tun\n")
	}
	statusPath := "/var/log/openvpn-status.log"
	logPath := "/var/log/openvpn.log"
	if vpnBase != "" {
		statusPath = filepath.Join(vpnBase, "openvpn-status.log")
		logPath = filepath.Join(vpnBase, "openvpn.log")
	}
	b.WriteString(fmt.Sprintf("status %s 10\n", statusPath))
	b.WriteString(fmt.Sprintf("log-append %s\n", logPath))
	b.WriteString(fmt.Sprintf("verb %s\n", strings.TrimSpace(verb)))
	if explicitExitNotify {
		b.WriteString("explicit-exit-notify 1\n")
	}
	if pushDNS1 != "" {
		b.WriteString(fmt.Sprintf("push \"dhcp-option DNS %s\"\n", pushDNS1))
	}
	if pushDNS2 != "" {
		b.WriteString(fmt.Sprintf("push \"dhcp-option DNS %s\"\n", pushDNS2))
	}
	if vpnGateway {
		b.WriteString("push \"redirect-gateway def1 bypass-dhcp\"\n")
	}
	for _, route := range buildPushRouteLines(pushRoutes) {
		b.WriteString(route)
		b.WriteString("\n")
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

func buildPushRouteLines(raw string) []string {
	lines := strings.Split(raw, "\n")
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "push ") {
			result = append(result, line)
			continue
		}
		ip, ipNet, err := net.ParseCIDR(line)
		if err != nil || ipNet == nil {
			continue
		}
		mask := net.IP(ipNet.Mask).String()
		result = append(result, fmt.Sprintf("push \"route %s %s\"", ip.Mask(ipNet.Mask).String(), mask))
	}
	return result
}

func formatManagementDirective(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if strings.Contains(raw, " ") {
		fields := strings.Fields(raw)
		if len(fields) >= 2 {
			return fmt.Sprintf("management %s %s", fields[0], fields[1])
		}
	}
	host, port, err := net.SplitHostPort(raw)
	if err == nil && strings.TrimSpace(host) != "" && strings.TrimSpace(port) != "" {
		return fmt.Sprintf("management %s %s", host, port)
	}
	if idx := strings.LastIndex(raw, ":"); idx > 0 && idx < len(raw)-1 && !strings.Contains(raw[idx+1:], ":") {
		host = strings.TrimSpace(raw[:idx])
		port = strings.TrimSpace(raw[idx+1:])
		if host != "" && port != "" {
			return fmt.Sprintf("management %s %s", host, port)
		}
	}
	return "management " + raw
}
