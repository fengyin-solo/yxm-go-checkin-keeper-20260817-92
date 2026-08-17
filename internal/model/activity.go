package model

import (
	"strings"
	"time"
)

// 签到活动状态。
const (
	ActivityDraft    = "draft"    // 草稿
	ActivityActive   = "active"   // 进行中
	ActivityFinished = "finished" // 已结束，终态
)

// activityTransitions 活动状态机：draft -> active -> finished。
var activityTransitions = map[string]map[string]bool{
	ActivityDraft:  {ActivityActive: true},
	ActivityActive: {ActivityFinished: true},
}

// CanActivityTransition 判断活动状态流转是否合法。
func CanActivityTransition(from, to string) bool {
	if m, ok := activityTransitions[from]; ok {
		return m[to]
	}
	return false
}

// Activity 签到打卡活动。
type Activity struct {
	ID          string    `json:"id"`
	Code        string    `json:"code"` // 活动编码，全局唯一
	Name        string    `json:"name"`
	Description string    `json:"description"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Validate 校验活动字段。
func (a *Activity) Validate() error {
	a.Code = strings.TrimSpace(a.Code)
	a.Name = strings.TrimSpace(a.Name)
	a.Description = strings.TrimSpace(a.Description)
	if a.Code == "" {
		return NewValidationError("code", "活动编码不能为空")
	}
	if len(a.Code) > 32 {
		return NewValidationError("code", "活动编码不能超过 32 个字符")
	}
	if a.Name == "" {
		return NewValidationError("name", "活动名称不能为空")
	}
	if a.StartTime.IsZero() {
		return NewValidationError("start_time", "开始时间不能为空")
	}
	if a.EndTime.IsZero() {
		return NewValidationError("end_time", "结束时间不能为空")
	}
	if a.EndTime.Before(a.StartTime) {
		return NewValidationError("end_time", "结束时间不能早于开始时间")
	}
	if a.Status == "" {
		a.Status = ActivityDraft
	}
	switch a.Status {
	case ActivityDraft, ActivityActive, ActivityFinished:
	default:
		return NewValidationError("status", "活动状态不合法")
	}
	return nil
}

// IsRunning 判断活动在给定时间点是否处于进行中的时间窗内。
func (a *Activity) IsRunning(now time.Time) bool {
	return a.Status == ActivityActive && !now.Before(a.StartTime) && !now.After(a.EndTime)
}

// ActivityFilter 活动列表筛选条件。
type ActivityFilter struct {
	Status  string
	Keyword string
}

// Match 判断活动是否满足筛选条件。
func (f ActivityFilter) Match(a *Activity) bool {
	if f.Status != "" && a.Status != f.Status {
		return false
	}
	if f.Keyword != "" {
		k := strings.ToLower(strings.TrimSpace(f.Keyword))
		if k != "" && !strings.Contains(strings.ToLower(a.Name), k) &&
			!strings.Contains(strings.ToLower(a.Code), k) {
			return false
		}
	}
	return true
}
