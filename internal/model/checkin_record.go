package model

import (
	"strings"
	"time"
)

// CheckinRecord 一次签到记录。
// 同一用户在同一活动同一天只能签到一次（由 service 层校验）。
type CheckinRecord struct {
	ID             string    `json:"id"`
	ActivityID     string    `json:"activity_id"`
	UserID         string    `json:"user_id"`
	CheckinDate    string    `json:"checkin_date"` // YYYY-MM-DD
	Streak         int       `json:"streak"`       // 本次签到时的连续天数
	PointsEarned   int64     `json:"points_earned"` // 本次签到获得的基础积分
	CreatedAt      time.Time `json:"created_at"`
}

// Validate 校验签到记录字段。
func (c *CheckinRecord) Validate() error {
	c.ActivityID = strings.TrimSpace(c.ActivityID)
	c.UserID = strings.TrimSpace(c.UserID)
	c.CheckinDate = strings.TrimSpace(c.CheckinDate)
	if c.ActivityID == "" {
		return NewValidationError("activity_id", "活动 ID 不能为空")
	}
	if c.UserID == "" {
		return NewValidationError("user_id", "用户 ID 不能为空")
	}
	if c.CheckinDate == "" {
		return NewValidationError("checkin_date", "签到日期不能为空")
	}
	if _, err := time.Parse("2006-01-02", c.CheckinDate); err != nil {
		return NewValidationError("checkin_date", "签到日期格式不合法，需为 YYYY-MM-DD")
	}
	if c.Streak < 1 {
		return NewValidationError("streak", "连续天数必须大于 0")
	}
	if c.PointsEarned < 0 {
		return NewValidationError("points_earned", "获得积分不能为负数")
	}
	return nil
}

// CheckinFilter 签到记录筛选条件。
type CheckinFilter struct {
	ActivityID string
	UserID     string
	FromDate   string
	ToDate     string
}

// Match 判断签到记录是否满足筛选条件。
func (f CheckinFilter) Match(c *CheckinRecord) bool {
	if f.ActivityID != "" && c.ActivityID != f.ActivityID {
		return false
	}
	if f.UserID != "" && c.UserID != f.UserID {
		return false
	}
	// FromDate/ToDate 均为闭区间：保留 [FromDate, ToDate] 内的记录。
	if f.FromDate != "" && c.CheckinDate < f.FromDate {
		return false
	}
	if f.ToDate != "" && c.CheckinDate > f.ToDate {
		return false
	}
	return true
}
