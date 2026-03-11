package main

import (
	"log"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"singleOpenVpn/internal/config"
	"singleOpenVpn/internal/model"
	"singleOpenVpn/internal/store"
)

// authcheck: 被 OpenVPN 调用的账号密码校验程序。
// 使用 SQLite 数据库中的 vpn_users 表校验：
// - 用户存在
// - 启用状态
// - 未过期
// - 密码与 PasswordHash 匹配
//
// OpenVPN 配置示例：
//
//	script-security 3
//	auth-user-pass-verify /usr/local/bin/authcheck via-file
//	username-as-common-name
//	verify-client-cert require
//
// OpenVPN 会通过临时文件（via-file 参数）传入账号密码：第一行为用户名，第二行为密码。
func main() {
	if len(os.Args) < 2 {
		log.Println("authcheck: missing credentials file argument")
		os.Exit(1)
	}

	credFile := os.Args[1]
	data, err := os.ReadFile(credFile)
	if err != nil {
		log.Println("authcheck: read cred file error:", err)
		os.Exit(1)
	}
	lines := strings.SplitN(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n", 3)
	if len(lines) < 2 {
		log.Println("authcheck: invalid cred file format")
		os.Exit(1)
	}
	user := strings.TrimSpace(lines[0])
	pass := strings.TrimSpace(lines[1])
	if user == "" || pass == "" {
		log.Println("authcheck: empty username or password")
		os.Exit(1)
	}

	cfg := config.Load()
	// 不能用 NewDB：会 AutoMigrate → readonly database。OpenAuthDB 与 NewDB 同路径打开，仅跳过 Migrate
	db, err := store.OpenAuthDB(cfg)
	if err != nil {
		log.Println("authcheck: open db error:", err)
		os.Exit(1)
	}

	var u model.VPNUser
	if err := db.Where("name = ? AND enabled = 1", user).First(&u).Error; err != nil {
		log.Println("authcheck: user not found or disabled:", user, err)
		os.Exit(1)
	}

	if u.ExpiresAt != nil && u.ExpiresAt.Before(time.Now()) {
		log.Println("authcheck: user expired:", user)
		os.Exit(1)
	}

	if u.PasswordHash == "" {
		log.Println("authcheck: user has empty password:", user)
		os.Exit(1)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(pass)); err != nil {
		log.Println("authcheck: password mismatch for user:", user)
		os.Exit(1)
	}

	log.Println("authcheck: user auth OK:", user)
	os.Exit(0)
}
