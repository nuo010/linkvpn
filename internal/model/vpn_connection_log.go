package model

import (
	"gorm.io/gorm"
)

// VPNConnectionLog VPN 用户连接/下线记录（由 status 文件同步得到）
type VPNConnectionLog struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	Username       string         `gorm:"size:64;not null;index" json:"username"`       // 证书 CN / 用户名
	Status         string         `gorm:"size:16;index" json:"status"`                  // 连接状态：success | failed 等
	RealIP         string         `gorm:"size:64" json:"real_ip"`                       // 用户登录 IP（公网）
	VirtualIP      string         `gorm:"size:32" json:"virtual_ip"`                    // 分配的 VPN IP
	ConnectedAt    NaiveTime      `gorm:"not null;index;type:text" json:"connected_at"` // 入库格式 2006-01-02 15:04:05
	DisconnectedAt *NaiveTime     `gorm:"type:text" json:"disconnected_at"`             // 下线时间，NULL 表示当前仍在线
	BytesRecv      int64          `json:"bytes_recv"`                                   // 上传流量（服务器收到）
	BytesSent      int64          `json:"bytes_sent"`                                   // 下载流量（服务器发出）
	CreatedAt      NaiveTime      `gorm:"type:text" json:"created_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

func (VPNConnectionLog) TableName() string { return "vpn_connection_logs" }

func (v *VPNConnectionLog) BeforeCreate(tx *gorm.DB) error {
	if v.CreatedAt.IsZero() {
		v.CreatedAt = NowNaive()
	}
	return nil
}

// AfterCreate 部分 sqlite 驱动未走 NaiveTime.Valuer，created_at 会变成 UTC 墙钟；插入后强制写成统一字符串
func (v *VPNConnectionLog) AfterCreate(tx *gorm.DB) error {
	if v.ID == 0 {
		return nil
	}
	s := NowNaive().String()
	if s == "" {
		return nil
	}
	return tx.Model(v).UpdateColumn("created_at", s).Error
}
