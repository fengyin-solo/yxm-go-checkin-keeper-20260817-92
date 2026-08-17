package service

import (
	"sort"
	"time"

	"checkinkeeper/internal/model"
	"checkinkeeper/pkg/idgen"
)

// Checkin 用户签到。
// 流程：校验用户与活动 -> 活动进行中 -> 今日未签到 -> 计算连续天数 -> 发放奖励。
func (s *Service) Checkin(activityID, userID string, now time.Time) (*model.CheckinRecord, error) {
	activity, err := s.store.GetActivity(activityID)
	if err != nil {
		return nil, err
	}
	if !activity.IsRunning(now) {
		return nil, model.NewValidationError("activity", "活动未在进行中，不可签到")
	}
	user, err := s.store.GetUser(userID)
	if err != nil {
		return nil, err
	}
	if user.Status != model.UserActive {
		return nil, model.NewValidationError("user", "用户已停用，不可签到")
	}
	date := now.Format("2006-01-02")
	if _, err := s.store.GetCheckinByUserActivityDate(userID, activityID, date); err == nil {
		return nil, model.NewValidationError("checkin", "今日已签到，不可重复签到")
	}

	streak := s.calcStreak(userID, activityID, now)
	basePoints := s.dailyBasePoints(activityID)
	totalPoints := basePoints

	record := &model.CheckinRecord{
		ID:           idgen.Hex(),
		ActivityID:   activityID,
		UserID:       userID,
		CheckinDate:  date,
		Streak:       streak,
		PointsEarned: basePoints,
		CreatedAt:    now,
	}
	if err := record.Validate(); err != nil {
		return nil, err
	}
	if err := s.store.CreateCheckinRecord(record); err != nil {
		return nil, err
	}

	// 发放基础积分
	if basePoints > 0 {
		s.grantReward(userID, activityID, "", model.GrantBase, basePoints, streak, now)
	}
	// 连续签到奖励
	bonusPoints := s.streakBonusPoints(activityID, streak)
	if bonusPoints > 0 {
		totalPoints += bonusPoints
		s.grantReward(userID, activityID, s.streakBonusRuleID(activityID, streak), model.GrantBonus, bonusPoints, streak, now)
	}

	user.Points += totalPoints
	user.UpdatedAt = now
	if err := s.store.UpdateUser(user); err != nil {
		return nil, err
	}
	s.log.Infof("用户 %s 在活动 %s 签到成功，连续 %d 天，获得 %d 积分",
		user.Username, activity.Code, streak, totalPoints)
	return record, nil
}

// calcStreak 计算本次签到后的连续天数。
// 若昨天已签到则连续天数 +1，否则重置为 1。
func (s *Service) calcStreak(userID, activityID string, now time.Time) int {
	yesterday := now.AddDate(0, 0, -1).Format("2006-01-02")
	if prev, err := s.store.GetCheckinByUserActivityDate(userID, activityID, yesterday); err == nil {
		return prev.Streak + 1
	}
	return 1
}

// dailyBasePoints 取作用于该活动的每日基础积分（取第一个生效的 daily_base 规则）。
func (s *Service) dailyBasePoints(activityID string) int64 {
	for _, r := range s.store.ListRewardRules() {
		if r.Type == model.RewardDailyBase && r.AppliesTo(activityID) {
			return r.Points
		}
	}
	return 0
}

// streakBonusPoints 取连续签到满 streak 天的额外奖励积分。
func (s *Service) streakBonusPoints(activityID string, streak int) int64 {
	var bonus int64
	for _, r := range s.store.ListRewardRules() {
		if r.Type == model.RewardStreakBonus && r.AppliesTo(activityID) &&
			r.Threshold > 0 && streak%r.Threshold == 0 {
			bonus += r.Points
		}
	}
	return bonus
}

// streakBonusRuleID 取触发连续奖励的规则 ID（用于流水留痕）。
func (s *Service) streakBonusRuleID(activityID string, streak int) string {
	for _, r := range s.store.ListRewardRules() {
		if r.Type == model.RewardStreakBonus && r.AppliesTo(activityID) &&
			r.Threshold > 0 && streak%r.Threshold == 0 {
			return r.ID
		}
	}
	return ""
}

// grantReward 写入一条奖励发放流水。
func (s *Service) grantReward(userID, activityID, ruleID, grantType string, points int64, streak int, now time.Time) {
	grant := &model.RewardGrant{
		ID:         idgen.Hex(),
		UserID:     userID,
		ActivityID: activityID,
		RuleID:     ruleID,
		Type:       grantType,
		Points:     points,
		Streak:     streak,
		CreatedAt:  now,
	}
	if err := grant.Validate(); err != nil {
		s.log.Warnf("奖励发放校验失败: %v", err)
		return
	}
	_ = s.store.CreateRewardGrant(grant)
}

// GetCheckinRecord 查询签到记录详情。
func (s *Service) GetCheckinRecord(id string) (*model.CheckinRecord, error) {
	return s.store.GetCheckinRecord(id)
}

// ListCheckinRecords 分页查询签到记录，按签到日期倒序。
func (s *Service) ListCheckinRecords(filter model.CheckinFilter, page, size int) ([]*model.CheckinRecord, int, error) {
	all := s.store.ListCheckinRecords()
	matched := make([]*model.CheckinRecord, 0, len(all))
	for _, c := range all {
		if filter.Match(c) {
			matched = append(matched, c)
		}
	}
	sort.Slice(matched, func(i, j int) bool {
		if matched[i].CheckinDate != matched[j].CheckinDate {
			return matched[i].CheckinDate > matched[j].CheckinDate
		}
		return matched[i].CreatedAt.After(matched[j].CreatedAt)
	})
	total := len(matched)
	start := (page - 1) * size
	if start >= total {
		return []*model.CheckinRecord{}, total, nil
	}
	end := start + size
	if end > total {
		end = total
	}
	return matched[start:end], total, nil
}

// UserStreak 查询某用户在某活动的当前连续签到天数与累计签到次数。
type UserStreak struct {
	UserID      string `json:"user_id"`
	ActivityID  string `json:"activity_id"`
	TotalDays   int    `json:"total_days"`
	CurrentStreak int  `json:"current_streak"`
	MaxStreak   int    `json:"max_streak"`
}

// GetUserStreak 统计用户在某活动的签到情况。
func (s *Service) GetUserStreak(userID, activityID string) (*UserStreak, error) {
	if _, err := s.store.GetUser(userID); err != nil {
		return nil, err
	}
	if _, err := s.store.GetActivity(activityID); err != nil {
		return nil, err
	}
	result := &UserStreak{UserID: userID, ActivityID: activityID}
	for _, c := range s.store.ListCheckinRecords() {
		if c.UserID != userID || c.ActivityID != activityID {
			continue
		}
		result.TotalDays++
		if c.Streak > result.MaxStreak {
			result.MaxStreak = c.Streak
		}
	}
	if latest, err := s.store.LatestCheckinByUserActivity(userID, activityID); err == nil {
		result.CurrentStreak = latest.Streak
	}
	return result, nil
}

// CheckinCalendar 某用户在某活动某月的签到日历。
type CheckinCalendar struct {
	UserID     string   `json:"user_id"`
	ActivityID string   `json:"activity_id"`
	Month      string   `json:"month"` // YYYY-MM
	Dates      []string `json:"dates"` // 已签到日期，升序
	Count      int      `json:"count"`
}

// GetUserCalendar 查询用户在某活动指定月份（YYYY-MM）的签到日期列表。
func (s *Service) GetUserCalendar(userID, activityID, month string) (*CheckinCalendar, error) {
	if _, err := s.store.GetUser(userID); err != nil {
		return nil, err
	}
	if _, err := s.store.GetActivity(activityID); err != nil {
		return nil, err
	}
	if _, err := time.Parse("2006-01", month); err != nil {
		return nil, model.NewValidationError("month", "月份格式不合法，需为 YYYY-MM")
	}
	calendar := &CheckinCalendar{
		UserID:     userID,
		ActivityID: activityID,
		Month:      month,
		Dates:      make([]string, 0),
	}
	for _, c := range s.store.ListCheckinRecords() {
		if c.UserID != userID || c.ActivityID != activityID {
			continue
		}
		if len(c.CheckinDate) >= 7 && c.CheckinDate[:7] == month {
			calendar.Dates = append(calendar.Dates, c.CheckinDate)
		}
	}
	sort.Strings(calendar.Dates)
	calendar.Count = len(calendar.Dates)
	return calendar, nil
}
