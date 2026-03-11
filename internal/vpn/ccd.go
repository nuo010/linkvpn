package vpn

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	CCDDir        = "ccd"
	ServerVPNIP   = "10.8.8.1"      // 服务端在 10.8.8.0/24 网段中的 IP
	ServerVPNCIDR = "10.8.8.0"      // 网段
	SubnetMask    = "255.255.255.0" // topology subnet 时 ifconfig-push 的第二参数必须是子网掩码
)

// ccdFilePath 返回该用户 CCD 文件路径（无扩展名，文件名=证书 CN）
func ccdFilePath(basePath, clientName string) string {
	return filepath.Join(basePath, CCDDir, sanitizeName(clientName))
}

// WriteCCD 为该客户端写入 CCD 配置，实现静态 IP（仅一行 ifconfig-push）
// clientName 为证书 CN（即用户名），staticIP 如 "10.8.8.2"
func WriteCCD(basePath, clientName, staticIP string) error {
	clientName = sanitizeName(clientName)
	if clientName == "" || staticIP == "" {
		return nil
	}
	dir := filepath.Join(basePath, CCDDir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	fpath := ccdFilePath(basePath, clientName)
	// topology subnet 下格式为: ifconfig-push <客户端IP> <子网掩码>
	body := "ifconfig-push " + strings.TrimSpace(staticIP) + " " + SubnetMask + "\n"
	return os.WriteFile(fpath, []byte(body), 0644)
}

// ReadCCD 读取该客户端的 CCD 文件内容；不存在时返回空字符串和 nil
func ReadCCD(basePath, clientName string) (string, error) {
	clientName = sanitizeName(clientName)
	if clientName == "" {
		return "", nil
	}
	fpath := ccdFilePath(basePath, clientName)
	b, err := os.ReadFile(fpath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return string(b), nil
}

// CCDTemplate 返回无文件时的默认 CCD 模板（可含一条 ifconfig-push）
func CCDTemplate(staticIP string) string {
	if staticIP == "" {
		staticIP = "10.8.8.x"
	}
	return fmt.Sprintf(`# 客户端 CCD 配置（client-config-dir）
# 修改后需客户端重连 VPN 后生效
# 每用户仅允许一条 ifconfig-push（topology subnet 下第二参数为子网掩码）
# 格式: ifconfig-push <客户端IP> %s
ifconfig-push %s %s
`, SubnetMask, strings.TrimSpace(staticIP), SubnetMask)
}

// WriteCCDContent 将原始内容写入该客户端的 CCD 文件（高级配置，可含 iroute、push 等）
// 写入时确保文件以换行结尾，否则 OpenVPN 可能不解析最后一行。
func WriteCCDContent(basePath, clientName, content string) error {
	clientName = sanitizeName(clientName)
	if clientName == "" {
		return fmt.Errorf("用户名为空")
	}
	dir := filepath.Join(basePath, CCDDir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	if content != "" && !strings.HasSuffix(content, "\n") {
		content = content + "\n"
	}
	fpath := ccdFilePath(basePath, clientName)
	return os.WriteFile(fpath, []byte(content), 0644)
}

// RemoveCCD 删除该客户端的 CCD 文件
func RemoveCCD(basePath, clientName string) error {
	clientName = sanitizeName(clientName)
	if clientName == "" {
		return nil
	}
	fpath := ccdFilePath(basePath, clientName)
	_ = os.Remove(fpath)
	return nil
}

// RepairCCDFilesTrailingNewline 遍历 ccd 目录下所有文件，若末尾缺少换行则补上（修复历史写入的配置）
func RepairCCDFilesTrailingNewline(basePath string) error {
	dir := filepath.Join(basePath, CCDDir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		fpath := filepath.Join(dir, e.Name())
		b, err := os.ReadFile(fpath)
		if err != nil || len(b) == 0 {
			continue
		}
		if b[len(b)-1] != '\n' {
			if err := os.WriteFile(fpath, append(b, '\n'), 0644); err != nil {
				return err
			}
		}
	}
	return nil
}

// IPPPoolPath 返回 ifconfig-pool-persist 文件路径（与 server.conf 中一致）
func IPPPoolPath(basePath string) string {
	return filepath.Join(basePath, "easy-rsa", "pki", "ipp.txt")
}

// RemoveClientFromIPPPool 从 ipp.txt 中移除该客户端的持久化 IP 记录。
// 修改用户固定 IP 或 CCD 后调用，使客户端下次连接时由 CCD 重新分配新 IP，而非沿用 ipp.txt 中的旧 IP。
func RemoveClientFromIPPPool(basePath, clientName string) error {
	clientName = sanitizeName(clientName)
	if clientName == "" {
		return nil
	}
	ippPath := IPPPoolPath(basePath)
	b, err := os.ReadFile(ippPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	prefix := clientName + ","
	var kept []string
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSuffix(line, "\r")
		if line == "" || strings.HasPrefix(line, prefix) {
			continue
		}
		kept = append(kept, line)
	}
	return os.WriteFile(ippPath, []byte(strings.Join(kept, "\n")+"\n"), 0644)
}
