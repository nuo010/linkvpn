package store

import (
	"fmt"
	"os"
	"path/filepath"

	"singleOpenVpn/internal/config"
	"singleOpenVpn/internal/model"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func NewDB(cfg *config.Config) (*gorm.DB, error) {
	dir := filepath.Dir(cfg.DatabasePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("创建数据库目录失败 %s: %w", dir, err)
	}
	// SQLite 必须在库文件所在目录可写：迁移会建临时表，运行时会写 -wal/-journal
	f, err := os.OpenFile(cfg.DatabasePath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, fmt.Errorf("数据库不可写（readonly database）: %w\n路径: %s\n请对该文件及目录 chmod/chown，或设置可写目录如 DATABASE_PATH=/tmp/panel.db 测试", err, cfg.DatabasePath)
	}
	_ = f.Close()

	db, err := gorm.Open(sqlite.Open(cfg.DatabasePath), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&model.VPNUser{}, &model.ServerConfig{}, &model.LoginLog{}, &model.VPNConnectionLog{}); err != nil {
		return nil, err
	}
	return db, nil
}

// OpenAuthDB 供 authcheck 使用：与 NewDB 用同一种方式打开同一文件，但不做 AutoMigrate。
// 之前用 file:...?mode=ro 会导致 glebarez 行为异常（查不到用户）；readonly 只发生在 Migrate，Open(path) 只读查询一般可工作。
func OpenAuthDB(cfg *config.Config) (*gorm.DB, error) {
	if cfg.DatabasePath == "" {
		return nil, fmt.Errorf("DATABASE_PATH 未配置")
	}
	// 必须与 NewDB 一致用 sqlite.Open(路径)，否则可能连到空库
	db, err := gorm.Open(sqlite.Open(cfg.DatabasePath), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("打开认证数据库失败: %w (path=%s)", err, cfg.DatabasePath)
	}
	return db, nil
}
