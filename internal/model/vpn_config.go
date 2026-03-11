package model

import (
	"gorm.io/gorm"
)

// ServerConfig 服务端配置（存数据库，便于面板修改）
type ServerConfig struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Key       string         `gorm:"size:128;uniqueIndex;not null" json:"key"`
	Value     string         `gorm:"type:text" json:"value"`
	Comment   string         `gorm:"size:256" json:"comment"`
	CreatedAt NaiveTime      `gorm:"type:text" json:"created_at"`
	UpdatedAt NaiveTime      `gorm:"type:text" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ServerConfig) TableName() string { return "server_configs" }

func (s *ServerConfig) BeforeCreate(tx *gorm.DB) error {
	if s.CreatedAt.IsZero() {
		s.CreatedAt = NowNaive()
	}
	if s.UpdatedAt.IsZero() {
		s.UpdatedAt = NowNaive()
	}
	return nil
}

func (s *ServerConfig) BeforeUpdate(tx *gorm.DB) error {
	s.UpdatedAt = NowNaive()
	return nil
}

func (s *ServerConfig) AfterCreate(tx *gorm.DB) error {
	if s.ID == 0 {
		return nil
	}
	str := NowNaive().String()
	if str == "" {
		return nil
	}
	_ = tx.Model(s).UpdateColumn("created_at", str)
	return tx.Model(s).UpdateColumn("updated_at", str).Error
}

func (s *ServerConfig) AfterUpdate(tx *gorm.DB) error {
	if s.ID == 0 {
		return nil
	}
	str := NowNaive().String()
	if str == "" {
		return nil
	}
	return tx.Model(s).UpdateColumn("updated_at", str).Error
}
