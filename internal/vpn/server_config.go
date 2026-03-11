// Package vpn 中写入 server.conf 的路径约定：ca/cert/key/dh、client-config-dir、status、log-append、ifconfig-pool-persist 等
// 必须使用绝对路径，避免 OpenVPN 进程工作目录非 VPN 目录时找不到文件。
package vpn

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DefaultServerConfig 默认服务端配置内容，默认网段 10.8.8.0/24
// basePath 为 VPN 工作目录，client-config-dir 使用绝对路径，避免依赖 OpenVPN 进程 CWD 导致读不到 CCD
func DefaultServerConfig(port int, serverVPNIP, basePath string) string {
	if serverVPNIP == "" {
		serverVPNIP = ServerVPNCIDR // 10.8.8.0
	}
	ccdDir := filepath.Join(basePath, CCDDir)
	if basePath == "" {
		basePath = "/etc/openvpn"
		ccdDir = filepath.Join(basePath, CCDDir) // 保持绝对路径
	}
	statusPath := filepath.Join(basePath, "openvpn-status.log")
	logPath := filepath.Join(basePath, "openvpn.log")
	return fmt.Sprintf(`# OpenVPN 服务端配置（可由管理面板编辑）
# 修改后需重启 OpenVPN 服务生效
# 本系统一用户一客户端，静态 IP 通过 ccd 目录实现

port %d
proto udp
dev tun

# PKI 使用绝对路径，避免 OpenVPN 进程 CWD 非工作目录时找不到文件
ca %s/easy-rsa/pki/ca.crt
cert %s/easy-rsa/pki/issued/server.crt
key %s/easy-rsa/pki/private/server.key
dh %s/easy-rsa/pki/dh.pem

# 虚拟网段 10.8.8.0/24（服务端占 10.8.8.1，客户端从 10.8.8.2 起或由 ccd 指定静态 IP）
topology subnet
server %s 255.255.255.0
ifconfig-pool-persist %s/easy-rsa/pki/ipp.txt
# 每用户/客户端可对应 ccd 目录下同名文件；使用绝对路径确保能读到 CCD
client-config-dir %s
# 允许客户端之间互访（如 10.8.8.6 ping 10.8.8.10）
client-to-client

keepalive 10 120
cipher AES-256-GCM
auth SHA256
user nobody
group nogroup
persist-key
persist-tun
# status 文件供面板显示在线状态与流量
status %s 10
log-append %s
verb 3
explicit-exit-notify 1

# 账号密码认证（基于 SQLite）：使用 /usr/local/bin/authcheck 校验 VPN 用户名与密码
script-security 3
auth-user-pass-verify /usr/local/bin/authcheck via-file
username-as-common-name
verify-client-cert require
`, port, basePath, basePath, basePath, basePath, strings.TrimSpace(serverVPNIP), basePath, ccdDir, statusPath, logPath)
}

// WriteServerConfig 将服务端配置写入 basePath 下的 server.conf
func WriteServerConfig(basePath, content string) error {
	p := filepath.Join(basePath, "server.conf")
	return os.WriteFile(p, []byte(content), 0644)
}

// EnsureClientConfigDir 确保 client-configs 目录存在
func EnsureClientConfigDir(basePath string) error {
	return os.MkdirAll(filepath.Join(basePath, "client-configs"), 0755)
}
