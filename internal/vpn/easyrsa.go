package vpn

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	EasyRSADir = "easy-rsa"
	PKIDir     = "pki"
)

// IsPKIEmpty 判断挂载目录是否为「新建」：无 CA 证书即视为未初始化，需自动创建
func IsPKIEmpty(basePath string) bool {
	caPath := filepath.Join(basePath, EasyRSADir, PKIDir, "ca.crt")
	_, err := os.Stat(caPath)
	return os.IsNotExist(err)
}

// EnsurePKI 若 PKI 为空则自动初始化：init-pki、build-ca、gen-dh、build-server-full，并写入默认 server.conf（若不存在）
func EnsurePKI(basePath, serverName string) error {
	if !IsPKIEmpty(basePath) {
		return nil
	}
	if err := InitPKI(basePath); err != nil {
		return err
	}
	if err := BuildCA(basePath, serverName); err != nil {
		return err
	}
	if err := GenDH(basePath); err != nil {
		return err
	}
	if err := GenServerCert(basePath, "server"); err != nil {
		return err
	}
	serverConfPath := filepath.Join(basePath, "server.conf")
	if _, err := os.Stat(serverConfPath); os.IsNotExist(err) {
		content := DefaultServerConfig(1194, ServerVPNCIDR, basePath)
		if err := WriteServerConfig(basePath, content); err != nil {
			return err
		}
	}
	return nil
}

// StartOpenVPN 在后台启动 OpenVPN，标准输出/错误追加到 openvpn.log（用于自动初始化 PKI 后启动）
func StartOpenVPN(basePath string) error {
	configPath := filepath.Join(basePath, "server.conf")
	if _, err := os.Stat(configPath); err != nil {
		return err
	}
	logPath := filepath.Join(basePath, "openvpn.log")
	logFile, err := os.OpenFile(logPath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	// 使用绝对路径配置，避免依赖进程 CWD
	cmd := exec.Command("openvpn", "--config", configPath)
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	err = cmd.Start()
	logFile.Close() // 子进程已持有 fd，关闭父进程句柄不影响子进程写入
	if err != nil {
		return err
	}
	return nil
}

func envWithBatch(env []string) []string {
	base := os.Environ()
	if env != nil {
		base = env
	}
	return append(base, "EASYRSA_BATCH=1")
}

// InitPKI 初始化 PKI（仅首次），调用 easyrsa init-pki
func InitPKI(basePath string) error {
	easyRSAPath := filepath.Join(basePath, EasyRSADir)
	if _, err := os.Stat(easyRSAPath); os.IsNotExist(err) {
		return err
	}
	cmd := exec.Command("bash", "-c", "cd "+easyRSAPath+" && ./easyrsa init-pki")
	cmd.Env = envWithBatch(cmd.Env)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// BuildCA 构建 CA，无交互
func BuildCA(basePath, cn string) error {
	easyRSAPath := filepath.Join(basePath, EasyRSADir)
	cmd := exec.Command("bash", "-c", "cd "+easyRSAPath+" && ./easyrsa --batch build-ca nopass")
	cmd.Env = envWithBatch(cmd.Env)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// GenDH 生成 DH 参数（build-server-full 前需要）
func GenDH(basePath string) error {
	easyRSAPath := filepath.Join(basePath, EasyRSADir)
	cmd := exec.Command("bash", "-c", "cd "+easyRSAPath+" && ./easyrsa --batch gen-dh")
	cmd.Env = envWithBatch(cmd.Env)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// GenServerCert 生成服务端证书
func GenServerCert(basePath, serverName string) error {
	easyRSAPath := filepath.Join(basePath, EasyRSADir)
	cmd := exec.Command("bash", "-c", "cd "+easyRSAPath+" && ./easyrsa --batch build-server-full "+serverName+" nopass")
	cmd.Env = envWithBatch(cmd.Env)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// GenClientCert 生成客户端证书
func GenClientCert(basePath, clientName string) error {
	easyRSAPath := filepath.Join(basePath, EasyRSADir)
	cmd := exec.Command("bash", "-c", "cd "+easyRSAPath+" && ./easyrsa --batch build-client-full "+sanitizeName(clientName)+" nopass")
	cmd.Env = envWithBatch(cmd.Env)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// RevokeClient 吊销客户端证书
func RevokeClient(basePath, clientName string) error {
	easyRSAPath := filepath.Join(basePath, EasyRSADir)
	cmd := exec.Command("bash", "-c", "cd "+easyRSAPath+" && ./easyrsa --batch revoke "+sanitizeName(clientName))
	cmd.Env = envWithBatch(cmd.Env)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func sanitizeName(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, " ", "_")
	s = strings.ReplaceAll(s, "/", "_")
	if s == "" {
		s = "client"
	}
	return s
}

// GetClientOVPNPath 返回客户端 .ovpn 文件路径（由 GenClientConfig 生成）
func GetClientOVPNPath(basePath, clientName string) string {
	return filepath.Join(basePath, "client-configs", sanitizeName(clientName)+".ovpn")
}

// GetPKIPath 返回 pki 目录
func GetPKIPath(basePath string) string {
	return filepath.Join(basePath, EasyRSADir, PKIDir)
}
