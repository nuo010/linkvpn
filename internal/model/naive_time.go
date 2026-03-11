package model

import (
	"database/sql/driver"
	"encoding/json"
	"strings"
	"time"
)

// NaiveTimeLayout 入库统一格式（无时区后缀）
const NaiveTimeLayout = "2006-01-02 15:04:05"

const naiveTimeLayout = NaiveTimeLayout

// Shanghai 东八区。系统内所有时间（含从字符串解析的无时区时间）一律按此时区理解；不再使用 time.Local/UTC 作为墙钟。
var Shanghai *time.Location

var naiveTimeLoc *time.Location

func init() {
	Shanghai, _ = time.LoadLocation("Asia/Shanghai")
	if Shanghai == nil {
		Shanghai = time.FixedZone("CST", 8*3600)
	}
	naiveTimeLoc = Shanghai
}

// NaiveTime 写入/读出均为 "2006-01-02 15:04:05"。sqlite 驱动对 Create/Update 有时不走 Valuer，需在 AfterCreate/AfterUpdate 里 UpdateColumn 成纯字符串
type NaiveTime struct {
	time.Time
}

// ParseOpenVPNLogTime 解析日志行首 "2006-01-02 15:04:05"（按 Asia/Shanghai 墙钟，与 openvpn.log 一致）
func ParseOpenVPNLogTime(line string) (time.Time, bool) {
	if len(line) < 19 {
		return time.Time{}, false
	}
	t, err := time.ParseInLocation(naiveTimeLayout, line[:19], naiveTimeLoc)
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}

// NowNaive 当前时刻的东八区墙钟写入库
func NowNaive() NaiveTime {
	return NT(time.Now().In(Shanghai))
}

// NT 从 time.Time 构造（用于赋值）
func NT(t time.Time) NaiveTime {
	if t.IsZero() {
		return NaiveTime{}
	}
	return NaiveTime{Time: t}
}

// NTP 指针，nil 表示 NULL
func NTP(t *time.Time) *NaiveTime {
	if t == nil {
		return nil
	}
	nt := NT(*t)
	return &nt
}

// String 返回入库/查询用格式字符串
func (n NaiveTime) String() string {
	if n.Time.IsZero() {
		return ""
	}
	return n.In(naiveTimeLoc).Format(naiveTimeLayout)
}

func (n NaiveTime) Value() (driver.Value, error) {
	if n.Time.IsZero() {
		return nil, nil
	}
	return n.In(naiveTimeLoc).Format(naiveTimeLayout), nil
}

func (n *NaiveTime) Scan(v interface{}) error {
	if v == nil {
		n.Time = time.Time{}
		return nil
	}
	switch x := v.(type) {
	case string:
		return n.parseString(x)
	case []byte:
		return n.parseString(string(x))
	default:
		n.Time = time.Time{}
		return nil
	}
}

func (n *NaiveTime) parseString(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		n.Time = time.Time{}
		return nil
	}
	t, err := time.ParseInLocation(naiveTimeLayout, s, naiveTimeLoc)
	if err != nil {
		return err
	}
	n.Time = t
	return nil
}

// MarshalJSON 输出与库内一致，前端少做时区换算
func (n NaiveTime) MarshalJSON() ([]byte, error) {
	if n.Time.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(n.In(naiveTimeLoc).Format(naiveTimeLayout))
}

func (n *NaiveTime) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		n.Time = time.Time{}
		return nil
	}
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	return n.parseString(s)
}

// Before 供过期判断等使用
func (n *NaiveTime) Before(u time.Time) bool {
	if n == nil {
		return false
	}
	return n.Time.Before(u)
}

// IsZero 未设置或零值
func (n NaiveTime) IsZero() bool {
	return n.Time.IsZero()
}

func (n *NaiveTime) IsNilOrZero() bool {
	if n == nil {
		return true
	}
	return n.Time.IsZero()
}
