package model

import (
	"strings"
	"time"
)

// 奖励规则类型。
const (
	RewardStreakBonus = "streak_bonus" // 连续签到满 N 天额外奖励，Threshold 为天数，Points 为奖励积分
	RewardDailyBase   = "daily_base"   // 每日基础积分，Points 为基础积分
)

// 奖励规则状态。
const (
	RewardRuleActive   = "active"
	RewardRuleInactive = "inactive"
)

// RewardRule 签到奖励规则，可作用于全局（ActivityID 为空）或指定活动。
type RewardRule struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Type       string    `json:"type"`
	ActivityID string    `json:"activity_id"` // 为空表示全局规则
	Threshold  int       `json:"threshold"`   // streak_bonus 的天数门槛
	Points     int64     `json:"points"`      // 奖励积分
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Validate 校验奖励规则字段。
func (r *RewardRule) Validate() error {
	r.Name = strings.TrimSpace(r.Name)
	r.ActivityID = strings.TrimSpace(r.ActivityID)
	if r.Name == "" {
		return NewValidationError("name", "规则名称不能为空")
	}
	switch r.Type {
	case RewardStreakBonus:
		if r.Threshold < 1 {
			return NewValidationError("threshold", "连续天数门槛必须大于 0")
		}
		if r.Points <= 0 {
			return NewValidationError("points", "奖励积分必须大于 0")
		}
	case RewardDailyBase:
		if r.Points <= 0 {
			return NewValidationError("points", "基础积分必须大于 0")
		}
	default:
		return NewValidationError("type", "奖励规则类型不合法")
	}
	if r.Status == "" {
		r.Status = RewardRuleActive
	}
	if r.Status != RewardRuleActive && r.Status != RewardRuleInactive {
		return NewValidationError("status", "规则状态不合法")
	}
	return nil
}

// AppliesTo 判断规则是否作用于指定活动（全局规则作用于所有活动）。
func (r *RewardRule) AppliesTo(activityID string) bool {
	if r.Status != RewardRuleActive {
		return false
	}
	return r.ActivityID == "" || r.ActivityID == activityID
}

// RewardRuleFilter 奖励规则列表筛选条件。
type RewardRuleFilter struct {
	Type       string
	Status     string
	ActivityID string
}

// Match 判断奖励规则是否满足筛选条件。
func (f RewardRuleFilter) Match(r *RewardRule) bool {
	if f.Type != "" && r.Type != f.Type {
		return false
	}
	if f.Status != "" && r.Status != f.Status {
		return false
	}
	if f.ActivityID != "" && r.ActivityID != f.ActivityID {
		return false
	}
	return true
}
