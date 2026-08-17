package service

import (
	"sort"

	"checkinkeeper/internal/model"
)

// OverviewStats 全局概览统计。
type OverviewStats struct {
	UserCount        int   `json:"user_count"`
	ActivityCount    int   `json:"activity_count"`
	CheckinCount     int   `json:"checkin_count"`
	ActiveActivities int   `json:"active_activities"`
	TotalPoints      int64 `json:"total_points"`      // 全部用户积分总和
	GrantedPoints    int64 `json:"granted_points"`    // 累计发放积分
	BonusGrants      int   `json:"bonus_grants"`      // 连续奖励发放次数
}

// Stats 全局概览统计。
func (s *Service) Stats() *OverviewStats {
	users := s.store.ListUsers()
	activities := s.store.ListActivities()
	stats := &OverviewStats{
		UserCount:     len(users),
		ActivityCount: len(activities),
		CheckinCount:  len(s.store.ListCheckinRecords()),
	}
	for _, a := range activities {
		if a.Status == model.ActivityActive {
			stats.ActiveActivities++
		}
	}
	for _, u := range users {
		stats.TotalPoints += u.Points
	}
	for _, g := range s.store.ListRewardGrants() {
		stats.GrantedPoints += g.Points
		if g.Type == model.GrantBonus {
			stats.BonusGrants++
		}
	}
	return stats
}

// ActivityStat 单活动统计。
type ActivityStat struct {
	ActivityID   string `json:"activity_id"`
	Code         string `json:"code"`
	Name         string `json:"name"`
	Status       string `json:"status"`
	CheckinCount int    `json:"checkin_count"`
	UserCount    int    `json:"user_count"` // 参与签到的去重用户数
}

// StatsByActivity 按活动分组统计签到情况，按签到数降序。
func (s *Service) StatsByActivity() []*ActivityStat {
	byActivity := make(map[string]*ActivityStat)
	for _, a := range s.store.ListActivities() {
		byActivity[a.ID] = &ActivityStat{
			ActivityID: a.ID, Code: a.Code, Name: a.Name, Status: a.Status,
		}
	}
	userSets := make(map[string]map[string]bool)
	for _, c := range s.store.ListCheckinRecords() {
		st, ok := byActivity[c.ActivityID]
		if !ok {
			continue
		}
		st.CheckinCount++
		if userSets[c.ActivityID] == nil {
			userSets[c.ActivityID] = make(map[string]bool)
		}
		userSets[c.ActivityID][c.UserID] = true
	}
	for id, set := range userSets {
		if st, ok := byActivity[id]; ok {
			st.UserCount = len(set)
		}
	}
	list := make([]*ActivityStat, 0, len(byActivity))
	for _, st := range byActivity {
		list = append(list, st)
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].CheckinCount != list[j].CheckinCount {
			return list[i].CheckinCount > list[j].CheckinCount
		}
		return list[i].Code < list[j].Code
	})
	return list
}

// DailyStat 单日签到统计。
type DailyStat struct {
	Date         string `json:"date"` // YYYY-MM-DD
	CheckinCount int    `json:"checkin_count"`
	PointsSum    int64  `json:"points_sum"`
}

// StatsByDay 按天分组统计签到，按日期升序。
func (s *Service) StatsByDay() []*DailyStat {
	byDay := make(map[string]*DailyStat)
	for _, c := range s.store.ListCheckinRecords() {
		st, ok := byDay[c.CheckinDate]
		if !ok {
			st = &DailyStat{Date: c.CheckinDate}
			byDay[c.CheckinDate] = st
		}
		st.CheckinCount++
		st.PointsSum += c.PointsEarned
	}
	list := make([]*DailyStat, 0, len(byDay))
	for _, st := range byDay {
		list = append(list, st)
	}
	sort.Slice(list, func(i, j int) bool { return list[i].Date < list[j].Date })
	return list
}

// LeaderboardEntry 积分排行榜条目。
type LeaderboardEntry struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Points   int64  `json:"points"`
}

// Leaderboard 积分排行榜 TOP N。
func (s *Service) Leaderboard(n int) []*LeaderboardEntry {
	if n <= 0 {
		n = 10
	}
	list := make([]*LeaderboardEntry, 0)
	for _, u := range s.store.ListUsers() {
		list = append(list, &LeaderboardEntry{
			UserID:   u.ID,
			Username: u.Username,
			Nickname: u.Nickname,
			Points:   u.Points,
		})
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].Points != list[j].Points {
			return list[i].Points > list[j].Points
		}
		return list[i].Username < list[j].Username
	})
	if len(list) > n {
		list = list[:n]
	}
	return list
}
