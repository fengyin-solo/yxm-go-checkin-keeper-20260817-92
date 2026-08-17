package model

import (
	"strings"
	"time"
)

// 用户状态。
const (
	UserActive   = "active"
	UserDisabled = "disabled"
)

// User 参与签到打卡的用户。
// Points 为积分余额，单位：积分（int64）。
type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Nickname  string    `json:"nickname"`
	Points    int64     `json:"points"` // 积分余额
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Validate 校验用户字段。
func (u *User) Validate() error {
	u.Username = strings.TrimSpace(u.Username)
	u.Nickname = strings.TrimSpace(u.Nickname)
	if u.Username == "" {
		return NewValidationError("username", "用户名不能为空")
	}
	if len(u.Username) > 32 {
		return NewValidationError("username", "用户名不能超过 32 个字符")
	}
	u.Nickname = strings.TrimSpace(u.Nickname)
	if u.Points < 0 {
		return NewValidationError("points", "积分余额不能为负数")
	}
	u.Status = strings.TrimSpace(u.Status)
	if u.Status == "" {
		u.Status = UserActive
	}
	if u.Status != UserActive && u.Status != UserDisabled {
		return NewValidationError("status", "用户状态不合法")
	}
	return nil
}

// UserFilter 用户列表筛选条件。
type UserFilter struct {
	Status  string
	Keyword string
}

// Match 判断用户是否满足筛选条件。
func (f UserFilter) Match(u *User) bool {
	if f.Status != "" && u.Status != f.Status {
		return false
	}
	if f.Keyword != "" {
		k := strings.ToLower(strings.TrimSpace(f.Keyword))
		if k != "" && !strings.Contains(strings.ToLower(u.Username), k) &&
			!strings.Contains(strings.ToLower(u.Nickname), k) {
			return false
		}
	}
	return true
}
