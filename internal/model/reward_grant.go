package model

import (
	"strings"
	"time"
)

// 奖励发放类型。
const (
	GrantBase  = "base"  // 每日基础积分
	GrantBonus = "bonus" // 连续签到奖励
)

// RewardGrant 一次积分奖励发放的不可变流水。
type RewardGrant struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	ActivityID string    `json:"activity_id"`
	RuleID     string    `json:"rule_id"` // 触发的规则；base 类型可为空
	Type       string    `json:"type"`
	Points     int64     `json:"points"`
	Streak     int       `json:"streak"` // 发放时的连续天数
	CreatedAt  time.Time `json:"created_at"`
}

// Validate 校验奖励发放字段。
func (g *RewardGrant) Validate() error {
	g.UserID = strings.TrimSpace(g.UserID)
	g.ActivityID = strings.TrimSpace(g.ActivityID)
	g.RuleID = strings.TrimSpace(g.RuleID)
	if g.UserID == "" {
		return NewValidationError("user_id", "用户 ID 不能为空")
	}
	if g.ActivityID == "" {
		return NewValidationError("activity_id", "活动 ID 不能为空")
	}
	if g.Type != GrantBase && g.Type != GrantBonus {
		return NewValidationError("type", "发放类型不合法")
	}
	if g.Points <= 0 {
		return NewValidationError("points", "发放积分必须大于 0")
	}
	if g.Streak < 1 {
		return NewValidationError("streak", "连续天数必须大于 0")
	}
	if g.CreatedAt.IsZero() {
		return NewValidationError("created_at", "发放时间不能为空")
	}
	return nil
}

// GrantFilter 奖励发放筛选条件。
type GrantFilter struct {
	UserID     string
	ActivityID string
	Type       string
}

// Match 判断奖励发放是否满足筛选条件。
func (f GrantFilter) Match(g *RewardGrant) bool {
	if f.UserID != "" && g.UserID != f.UserID {
		return false
	}
	if f.ActivityID != "" && g.ActivityID != f.ActivityID {
		return false
	}
	if f.Type != "" && g.Type != f.Type {
		return false
	}
	return true
}
