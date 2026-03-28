package vpn

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GenClientOVPN 生成客户端 .ovpn 文件内容并写入
// 本系统一用户一客户端，clientName 即用户名（证书 CN）
// serverAddr 为公网地址，如 "vpn.example.com" 或 "1.2.3.4"
// routeNopull 为 true 时在配置末尾添加 route-nopull，使客户端忽略服务端推送的路由
func GenClientOVPN(basePath, clientName, serverAddr string, port int, proto string, routeNopull bool) (string, error) {
	clientName = sanitizeName(clientName)
	if clientName == "" {
		return "", errors.New("客户端名不能为空")
	}
	pkiPath := GetPKIPath(basePath)
	dir := filepath.Join(basePath, "client-configs")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("创建目录失败: %w", err)
	}

	caPath := filepath.Join(pkiPath, "ca.crt")
	certPath := filepath.Join(pkiPath, "issued", clientName+".crt")
	keyPath := filepath.Join(pkiPath, "private", clientName+".key")

	ca, err := os.ReadFile(caPath)
	if err != nil {
		return "", fmt.Errorf("读取 CA 证书失败(%s): %w", caPath, err)
	}
	cert, err := os.ReadFile(certPath)
	if err != nil {
		return "", fmt.Errorf("读取客户端证书失败(%s)，请确认已为该用户生成证书: %w", certPath, err)
	}
	key, err := os.ReadFile(keyPath)
	if err != nil {
		return "", fmt.Errorf("读取客户端私钥失败(%s): %w", keyPath, err)
	}
	caStr := trimPEM(string(ca))
	certStr := trimPEM(string(cert))
	keyStr := trimPEM(string(key))
	if caStr == "" || certStr == "" || keyStr == "" {
		return "", errors.New("CA/证书/私钥内容为空，请先在「系统配置」中初始化 PKI 并为该用户生成证书")
	}

	proto = strings.ToLower(strings.TrimSpace(proto))
	if proto != "tcp" && proto != "udp" {
		proto = "udp"
	}
	content := fmt.Sprintf(`client
dev tun
proto %s
remote %s %d
resolv-retry infinite
nobind
persist-key
persist-tun
remote-cert-tls server
cipher AES-256-GCM
auth SHA256
auth-user-pass
verb 3

<ca>
%s</ca>

<cert>
%s</cert>

<key>
%s</key>
`, proto, serverAddr, port, caStr, certStr, keyStr)

	if routeNopull {
		content += "\n# 忽略服务端推送的路由\nroute-nopull\n"
	}

	outPath := filepath.Join(dir, clientName+".ovpn")
	if err := os.WriteFile(outPath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("写入配置文件失败: %w", err)
	}
	return outPath, nil
}

func trimPEM(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if !strings.HasSuffix(s, "\n") {
		s += "\n"
	}
	return s
}
