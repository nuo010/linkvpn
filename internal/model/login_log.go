package model

import (
	"gorm.io/gorm"
)

// LoginLog 面板登录日志（管理员登录成功/失败）
type LoginLog struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Username  string         `gorm:"size:64;not null" json:"username"`
	Success   bool           `json:"success"`
	Message   string         `gorm:"size:256" json:"message"` // 如 "登录成功" / "用户名或密码错误"
	SourceIP  string         `gorm:"size:64" json:"source_ip"`
	CreatedAt NaiveTime      `gorm:"type:text" json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (LoginLog) TableName() string { return "login_logs" }
