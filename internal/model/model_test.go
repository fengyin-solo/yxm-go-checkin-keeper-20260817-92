package model

import (
	"testing"
	"time"
)

func TestUserValidate(t *testing.T) {
	base := func() *User {
		return &User{Username: "alice"}
	}
	cases := []struct {
		name    string
		mutate  func(*User)
		wantErr bool
	}{
		{"合法", func(u *User) {}, false},
		{"空用户名", func(u *User) { u.Username = "" }, true},
		{"用户名过长", func(u *User) { u.Username = "012345678901234567890123456789012345" }, true},
		{"负积分", func(u *User) { u.Points = -1 }, true},
		{"非法状态", func(u *User) { u.Status = "weird" }, true},
	}
	for _, c := range cases {
		u := base()
		c.mutate(u)
		err := u.Validate()
		if c.wantErr && err == nil {
			t.Errorf("%s: 期望校验失败", c.name)
		}
		if !c.wantErr && err != nil {
			t.Errorf("%s: 期望校验通过，得到 %v", c.name, err)
		}
	}
}

func TestUserFilterMatch(t *testing.T) {
	u := &User{Username: "alice", Nickname: "爱丽丝", Status: UserActive}
	if !(UserFilter{}).Match(u) {
		t.Error("空筛选应匹配所有")
	}
	if !(UserFilter{Status: UserActive, Keyword: "ali"}).Match(u) {
		t.Error("状态+关键词匹配应通过")
	}
	if (UserFilter{Status: UserDisabled}).Match(u) {
		t.Error("状态不匹配应失败")
	}
	if !(UserFilter{Keyword: "爱丽丝"}).Match(u) {
		t.Error("关键词匹配昵称应通过")
	}
	if (UserFilter{Keyword: "bob"}).Match(u) {
		t.Error("关键词不匹配应失败")
	}
}

func TestActivityValidateAndTransition(t *testing.T) {
	now := time.Now()
	base := func() *Activity {
		return &Activity{Code: "A1", Name: "活动", StartTime: now, EndTime: now.Add(time.Hour)}
	}
	cases := []struct {
		name    string
		mutate  func(*Activity)
		wantErr bool
	}{
		{"合法", func(a *Activity) {}, false},
		{"空编码", func(a *Activity) { a.Code = "" }, true},
		{"编码过长", func(a *Activity) { a.Code = "012345678901234567890123456789012345" }, true},
		{"空名称", func(a *Activity) { a.Name = "" }, true},
		{"缺开始时间", func(a *Activity) { a.StartTime = time.Time{} }, true},
		{"缺结束时间", func(a *Activity) { a.EndTime = time.Time{} }, true},
		{"结束早于开始", func(a *Activity) { a.EndTime = now.Add(-time.Hour) }, true},
		{"非法状态", func(a *Activity) { a.Status = "weird" }, true},
	}
	for _, c := range cases {
		a := base()
		c.mutate(a)
		err := a.Validate()
		if c.wantErr && err == nil {
			t.Errorf("%s: 期望校验失败", c.name)
		}
		if !c.wantErr && err != nil {
			t.Errorf("%s: 期望校验通过，得到 %v", c.name, err)
		}
	}
	transitions := []struct {
		from, to string
		want     bool
	}{
		{ActivityDraft, ActivityActive, true},
		{ActivityDraft, ActivityFinished, false},
		{ActivityActive, ActivityFinished, true},
		{ActivityActive, ActivityDraft, false},
		{ActivityFinished, ActivityActive, false},
	}
	for _, c := range transitions {
		if got := CanActivityTransition(c.from, c.to); got != c.want {
			t.Errorf("CanActivityTransition(%s,%s)=%v，期望 %v", c.from, c.to, got, c.want)
		}
	}
}

func TestActivityIsRunning(t *testing.T) {
	now := time.Now()
	a := &Activity{Status: ActivityActive, StartTime: now.Add(-time.Hour), EndTime: now.Add(time.Hour)}
	if !a.IsRunning(now) {
		t.Error("进行中的活动应 IsRunning")
	}
	if a.IsRunning(now.Add(2 * time.Hour)) {
		t.Error("超出结束时间不应 IsRunning")
	}
	a.Status = ActivityDraft
	if a.IsRunning(now) {
		t.Error("草稿活动不应 IsRunning")
	}
}

func TestActivityFilterMatch(t *testing.T) {
	a := &Activity{Code: "DAILY", Name: "每日打卡", Status: ActivityActive}
	if !(ActivityFilter{}).Match(a) {
		t.Error("空筛选应匹配所有")
	}
	if !(ActivityFilter{Status: ActivityActive, Keyword: "打卡"}).Match(a) {
		t.Error("状态+关键词匹配应通过")
	}
	if (ActivityFilter{Status: ActivityDraft}).Match(a) {
		t.Error("状态不匹配应失败")
	}
	if !(ActivityFilter{Keyword: "daily"}).Match(a) {
		t.Error("关键词匹配编码（忽略大小写）应通过")
	}
}

func TestCheckinRecordValidate(t *testing.T) {
	base := func() *CheckinRecord {
		return &CheckinRecord{ActivityID: "a1", UserID: "u1", CheckinDate: "2026-08-16", Streak: 1, PointsEarned: 10}
	}
	cases := []struct {
		name    string
		mutate  func(*CheckinRecord)
		wantErr bool
	}{
		{"合法", func(c *CheckinRecord) {}, false},
		{"空活动ID", func(c *CheckinRecord) { c.ActivityID = "" }, true},
		{"空用户ID", func(c *CheckinRecord) { c.UserID = "" }, true},
		{"空日期", func(c *CheckinRecord) { c.CheckinDate = "" }, true},
		{"日期格式错误", func(c *CheckinRecord) { c.CheckinDate = "2026/08/16" }, true},
		{"连续天数为0", func(c *CheckinRecord) { c.Streak = 0 }, true},
		{"负积分", func(c *CheckinRecord) { c.PointsEarned = -1 }, true},
	}
	for _, c := range cases {
		rec := base()
		c.mutate(rec)
		err := rec.Validate()
		if c.wantErr && err == nil {
			t.Errorf("%s: 期望校验失败", c.name)
		}
		if !c.wantErr && err != nil {
			t.Errorf("%s: 期望校验通过，得到 %v", c.name, err)
		}
	}
}

func TestCheckinFilterMatch(t *testing.T) {
	c := &CheckinRecord{ActivityID: "a1", UserID: "u1", CheckinDate: "2026-08-15"}
	if !(CheckinFilter{}).Match(c) {
		t.Error("空筛选应匹配所有")
	}
	if !(CheckinFilter{ActivityID: "a1", UserID: "u1"}).Match(c) {
		t.Error("全条件匹配应通过")
	}
	if (CheckinFilter{UserID: "u2"}).Match(c) {
		t.Error("用户不匹配应失败")
	}
	if !(CheckinFilter{FromDate: "2026-08-01", ToDate: "2026-08-31"}).Match(c) {
		t.Error("日期范围内应匹配")
	}
	if (CheckinFilter{FromDate: "2026-08-16"}).Match(c) {
		t.Error("早于 FromDate 不应匹配")
	}
	if (CheckinFilter{ToDate: "2026-08-14"}).Match(c) {
		t.Error("晚于 ToDate 不应匹配")
	}
}

func TestRewardRuleValidateAndHelpers(t *testing.T) {
	cases := []struct {
		name    string
		rule    RewardRule
		wantErr bool
	}{
		{"每日基础合法", RewardRule{Name: "r", Type: RewardDailyBase, Points: 10}, false},
		{"每日基础零积分", RewardRule{Name: "r", Type: RewardDailyBase, Points: 0}, true},
		{"连续奖励合法", RewardRule{Name: "r", Type: RewardStreakBonus, Threshold: 7, Points: 50}, false},
		{"连续奖励零门槛", RewardRule{Name: "r", Type: RewardStreakBonus, Threshold: 0, Points: 50}, true},
		{"连续奖励零积分", RewardRule{Name: "r", Type: RewardStreakBonus, Threshold: 7, Points: 0}, true},
		{"非法类型", RewardRule{Name: "r", Type: "weird"}, true},
		{"空名称", RewardRule{Type: RewardDailyBase, Points: 10}, true},
		{"非法状态", RewardRule{Name: "r", Type: RewardDailyBase, Points: 10, Status: "weird"}, true},
	}
	for _, c := range cases {
		r := c.rule
		err := r.Validate()
		if c.wantErr && err == nil {
			t.Errorf("%s: 期望校验失败", c.name)
		}
		if !c.wantErr && err != nil {
			t.Errorf("%s: 期望校验通过，得到 %v", c.name, err)
		}
	}
	// AppliesTo
	global := &RewardRule{Status: RewardRuleActive, ActivityID: ""}
	if !global.AppliesTo("any") {
		t.Error("全局规则应作用于所有活动")
	}
	scoped := &RewardRule{Status: RewardRuleActive, ActivityID: "a1"}
	if !scoped.AppliesTo("a1") || scoped.AppliesTo("a2") {
		t.Error("定向规则作用域不对")
	}
	inactive := &RewardRule{Status: RewardRuleInactive, ActivityID: ""}
	if inactive.AppliesTo("any") {
		t.Error("停用规则不应生效")
	}
}

func TestRewardGrantValidate(t *testing.T) {
	base := func() *RewardGrant {
		return &RewardGrant{UserID: "u1", ActivityID: "a1", Type: GrantBase, Points: 10, Streak: 1, CreatedAt: time.Now()}
	}
	cases := []struct {
		name    string
		mutate  func(*RewardGrant)
		wantErr bool
	}{
		{"合法", func(g *RewardGrant) {}, false},
		{"空用户ID", func(g *RewardGrant) { g.UserID = "" }, true},
		{"空活动ID", func(g *RewardGrant) { g.ActivityID = "" }, true},
		{"非法类型", func(g *RewardGrant) { g.Type = "weird" }, true},
		{"零积分", func(g *RewardGrant) { g.Points = 0 }, true},
		{"连续天数为0", func(g *RewardGrant) { g.Streak = 0 }, true},
		{"零值时间", func(g *RewardGrant) { g.CreatedAt = time.Time{} }, true},
	}
	for _, c := range cases {
		g := base()
		c.mutate(g)
		err := g.Validate()
		if c.wantErr && err == nil {
			t.Errorf("%s: 期望校验失败", c.name)
		}
		if !c.wantErr && err != nil {
			t.Errorf("%s: 期望校验通过，得到 %v", c.name, err)
		}
	}
}

func TestGrantFilterMatch(t *testing.T) {
	g := &RewardGrant{UserID: "u1", ActivityID: "a1", Type: GrantBase}
	if !(GrantFilter{}).Match(g) {
		t.Error("空筛选应匹配所有")
	}
	if !(GrantFilter{UserID: "u1", ActivityID: "a1", Type: GrantBase}).Match(g) {
		t.Error("全条件匹配应通过")
	}
	if (GrantFilter{Type: GrantBonus}).Match(g) {
		t.Error("类型不匹配应失败")
	}
	if (GrantFilter{UserID: "u2"}).Match(g) {
		t.Error("用户不匹配应失败")
	}
}
