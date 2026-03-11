package model

import (
	"gorm.io/gorm"
)

// Kind 用户类型：user=用户，client=客户端
const KindUser = "user"
const KindClient = "client"

type VPNUser struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Name         string         `gorm:"size:64;uniqueIndex;not null" json:"name"`
	Kind         string         `gorm:"size:16;default:user" json:"kind"` // user | client
	Email        string         `gorm:"size:128" json:"email"`
	Remark       string         `gorm:"size:256" json:"remark"`            // 备注
	PasswordHash string         `gorm:"size:128" json:"-"`                 // 不明文返回
	StaticIP     string         `gorm:"size:32" json:"static_ip"`          // 静态 IP，如 10.8.8.2，对应 CCD
	RouteNopull  bool           `gorm:"default:false" json:"route_nopull"` // 下载的配置中是否添加 route-nopull，忽略服务端推送的路由
	ExpiresAt    *NaiveTime     `gorm:"type:text" json:"expires_at"`       // 账号到期时间，nil 表示永不过期
	Enabled      bool           `gorm:"default:true" json:"enabled"`
	CreatedAt    NaiveTime      `gorm:"type:text" json:"created_at"`
	UpdatedAt    NaiveTime      `gorm:"type:text" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (VPNUser) TableName() string { return "vpn_users" }

func (u *VPNUser) BeforeCreate(tx *gorm.DB) error {
	if u.CreatedAt.IsZero() {
		u.CreatedAt = NowNaive()
	}
	if u.UpdatedAt.IsZero() {
		u.UpdatedAt = NowNaive()
	}
	return nil
}

func (u *VPNUser) BeforeUpdate(tx *gorm.DB) error {
	u.UpdatedAt = NowNaive()
	return nil
}

// AfterCreate/AfterUpdate：glebarez/sqlite 对 NaiveTime 有时仍写入 RFC3339Nano，强制改成统一 naive 字符串，避免下次 Find 扫描失败
func (u *VPNUser) AfterCreate(tx *gorm.DB) error {
	if u.ID == 0 {
		return nil
	}
	s := NowNaive().String()
	if s == "" {
		return nil
	}
	_ = tx.Model(u).UpdateColumn("created_at", s)
	return tx.Model(u).UpdateColumn("updated_at", s).Error
}

func (u *VPNUser) AfterUpdate(tx *gorm.DB) error {
	if u.ID == 0 {
		return nil
	}
	s := NowNaive().String()
	if s == "" {
		return nil
	}
	return tx.Model(u).UpdateColumn("updated_at", s).Error
}
