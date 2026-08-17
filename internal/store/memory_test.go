package store

import (
	"testing"
	"time"

	"checkinkeeper/internal/model"
)

func newUser(id, username string) *model.User {
	return &model.User{
		ID:        id,
		Username:  username,
		Nickname:  username,
		Status:    model.UserActive,
		CreatedAt: time.Now(),
	}
}

func TestUserCRUDAndConflict(t *testing.T) {
	s := NewMemoryStore()
	u := newUser("u1", "alice")
	if err := s.CreateUser(u); err != nil {
		t.Fatal(err)
	}
	if err := s.CreateUser(newUser("u2", "alice")); err != ErrConflict {
		t.Fatalf("重复用户名期望 ErrConflict，得到 %v", err)
	}
	if got, err := s.GetUserByUsername("alice"); err != nil || got.ID != "u1" {
		t.Fatalf("按用户名查询失败: %v", err)
	}
	if _, err := s.GetUserByUsername("nobody"); err != ErrNotFound {
		t.Fatalf("期望 ErrNotFound，得到 %v", err)
	}
	if list := s.ListUsers(); len(list) != 1 {
		t.Fatalf("期望 1 个用户，得到 %d", len(list))
	}
	u.Nickname = "爱丽丝"
	if err := s.UpdateUser(u); err != nil {
		t.Fatal(err)
	}
	if err := s.DeleteUser("u1"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.GetUser("u1"); err != ErrNotFound {
		t.Fatalf("删除后期望 ErrNotFound，得到 %v", err)
	}
	if err := s.UpdateUser(newUser("missing", "x")); err != ErrNotFound {
		t.Fatalf("期望 ErrNotFound，得到 %v", err)
	}
}

func newActivity(id, code string) *model.Activity {
	return &model.Activity{
		ID:        id,
		Code:      code,
		Name:      "测试活动",
		StartTime: time.Now().Add(-time.Hour),
		EndTime:   time.Now().Add(24 * time.Hour),
		Status:    model.ActivityActive,
		CreatedAt: time.Now(),
	}
}

func TestActivityCRUDAndConflict(t *testing.T) {
	s := NewMemoryStore()
	a := newActivity("a1", "DAILY")
	if err := s.CreateActivity(a); err != nil {
		t.Fatal(err)
	}
	if err := s.CreateActivity(newActivity("a2", "DAILY")); err != ErrConflict {
		t.Fatalf("重复编码期望 ErrConflict，得到 %v", err)
	}
	if got, err := s.GetActivityByCode("DAILY"); err != nil || got.ID != "a1" {
		t.Fatalf("按编码查询失败: %v", err)
	}
	if _, err := s.GetActivityByCode("NOPE"); err != ErrNotFound {
		t.Fatalf("期望 ErrNotFound，得到 %v", err)
	}
	if list := s.ListActivities(); len(list) != 1 {
		t.Fatalf("期望 1 个活动，得到 %d", len(list))
	}
	a.Name = "新名字"
	if err := s.UpdateActivity(a); err != nil {
		t.Fatal(err)
	}
	if err := s.DeleteActivity("a1"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.GetActivity("a1"); err != ErrNotFound {
		t.Fatalf("删除后期望 ErrNotFound，得到 %v", err)
	}
	if err := s.DeleteActivity("a1"); err != ErrNotFound {
		t.Fatalf("重复删除期望 ErrNotFound，得到 %v", err)
	}
}

func newCheckin(id, userID, activityID, date string) *model.CheckinRecord {
	return &model.CheckinRecord{
		ID:           id,
		ActivityID:   activityID,
		UserID:       userID,
		CheckinDate:  date,
		Streak:       1,
		PointsEarned: 10,
		CreatedAt:    time.Now(),
	}
}

func TestCheckinRecordCreateAndConflict(t *testing.T) {
	s := NewMemoryStore()
	c := newCheckin("c1", "u1", "a1", "2026-08-16")
	if err := s.CreateCheckinRecord(c); err != nil {
		t.Fatal(err)
	}
	// 同用户同活动同天重复签到冲突
	if err := s.CreateCheckinRecord(newCheckin("c2", "u1", "a1", "2026-08-16")); err != ErrConflict {
		t.Fatalf("重复签到期望 ErrConflict，得到 %v", err)
	}
	// 不同天不冲突
	if err := s.CreateCheckinRecord(newCheckin("c3", "u1", "a1", "2026-08-17")); err != nil {
		t.Fatalf("不同天签到不应冲突: %v", err)
	}
	if got, err := s.GetCheckinByUserActivityDate("u1", "a1", "2026-08-16"); err != nil || got.ID != "c1" {
		t.Fatalf("按日期查询失败: %v", err)
	}
	if _, err := s.GetCheckinByUserActivityDate("u1", "a1", "2026-01-01"); err != ErrNotFound {
		t.Fatalf("期望 ErrNotFound，得到 %v", err)
	}
	if list := s.ListCheckinRecords(); len(list) != 2 {
		t.Fatalf("期望 2 条记录，得到 %d", len(list))
	}
}

func TestLatestCheckinByUserActivity(t *testing.T) {
	s := NewMemoryStore()
	_ = s.CreateCheckinRecord(newCheckin("c1", "u1", "a1", "2026-08-14"))
	_ = s.CreateCheckinRecord(newCheckin("c2", "u1", "a1", "2026-08-16"))
	_ = s.CreateCheckinRecord(newCheckin("c3", "u1", "a1", "2026-08-15"))
	latest, err := s.LatestCheckinByUserActivity("u1", "a1")
	if err != nil || latest.CheckinDate != "2026-08-16" {
		t.Fatalf("期望最近签到为 2026-08-16，得到 %v err=%v", latest, err)
	}
	if _, err := s.LatestCheckinByUserActivity("u9", "a1"); err != ErrNotFound {
		t.Fatalf("无记录期望 ErrNotFound，得到 %v", err)
	}
}

func newRewardRule(id, name, ruleType string) *model.RewardRule {
	r := &model.RewardRule{
		ID:        id,
		Name:      name,
		Type:      ruleType,
		Points:    10,
		Status:    model.RewardRuleActive,
		CreatedAt: time.Now(),
	}
	if ruleType == model.RewardStreakBonus {
		r.Threshold = 7
	}
	return r
}

func TestRewardRuleCRUD(t *testing.T) {
	s := NewMemoryStore()
	r := newRewardRule("r1", "每日基础", model.RewardDailyBase)
	if err := s.CreateRewardRule(r); err != nil {
		t.Fatal(err)
	}
	if got, err := s.GetRewardRule("r1"); err != nil || got.Name != "每日基础" {
		t.Fatalf("查询规则失败: %v", err)
	}
	if list := s.ListRewardRules(); len(list) != 1 {
		t.Fatalf("期望 1 条规则，得到 %d", len(list))
	}
	r.Status = model.RewardRuleInactive
	if err := s.UpdateRewardRule(r); err != nil {
		t.Fatal(err)
	}
	if err := s.DeleteRewardRule("r1"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.GetRewardRule("r1"); err != ErrNotFound {
		t.Fatalf("删除后期望 ErrNotFound，得到 %v", err)
	}
	if err := s.UpdateRewardRule(newRewardRule("missing", "x", model.RewardDailyBase)); err != ErrNotFound {
		t.Fatalf("期望 ErrNotFound，得到 %v", err)
	}
	if err := s.DeleteRewardRule("missing"); err != ErrNotFound {
		t.Fatalf("期望 ErrNotFound，得到 %v", err)
	}
}

func newGrant(id, userID string) *model.RewardGrant {
	return &model.RewardGrant{
		ID:         id,
		UserID:     userID,
		ActivityID: "a1",
		Type:       model.GrantBase,
		Points:     10,
		Streak:     1,
		CreatedAt:  time.Now(),
	}
}

func TestRewardGrantCreateAndList(t *testing.T) {
	s := NewMemoryStore()
	if err := s.CreateRewardGrant(newGrant("g1", "u1")); err != nil {
		t.Fatal(err)
	}
	if err := s.CreateRewardGrant(newGrant("g2", "u1")); err != nil {
		t.Fatal(err)
	}
	if got, err := s.GetRewardGrant("g1"); err != nil || got.UserID != "u1" {
		t.Fatalf("查询发放流水失败: %v", err)
	}
	if _, err := s.GetRewardGrant("missing"); err != ErrNotFound {
		t.Fatalf("期望 ErrNotFound，得到 %v", err)
	}
	if list := s.ListRewardGrants(); len(list) != 2 {
		t.Fatalf("期望 2 条流水，得到 %d", len(list))
	}
}
