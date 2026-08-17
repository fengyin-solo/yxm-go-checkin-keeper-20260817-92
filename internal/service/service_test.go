package service

import (
	"testing"
	"time"

	"checkinkeeper/internal/config"
	"checkinkeeper/internal/model"
	"checkinkeeper/internal/store"
	"checkinkeeper/pkg/logger"
)

func newTestService() *Service {
	cfg := &config.Config{MaxPageSize: 100}
	log := logger.NewLevel(logger.LevelError)
	return New(store.NewMemoryStore(), log, cfg)
}

// baseTime 固定基准时间，便于控制签到日期。
var baseTime = time.Date(2026, 8, 10, 10, 0, 0, 0, time.UTC)

// setupActiveActivity 创建用户与进行中的活动，返回 (活动, 用户)。
func setupActiveActivity(t *testing.T, svc *Service) (*model.Activity, *model.User) {
	t.Helper()
	user, err := svc.CreateUser(model.User{Username: "alice"})
	if err != nil {
		t.Fatal(err)
	}
	activity, err := svc.CreateActivity(model.Activity{
		Code:      "DAILY",
		Name:      "每日打卡",
		StartTime: baseTime.AddDate(0, 0, -1),
		EndTime:   baseTime.AddDate(0, 0, 30),
	})
	if err != nil {
		t.Fatal(err)
	}
	activity, err = svc.TransitionActivity(activity.ID, model.ActivityActive)
	if err != nil {
		t.Fatal(err)
	}
	return activity, user
}

func TestCreateUserValidation(t *testing.T) {
	svc := newTestService()
	cases := []struct {
		name  string
		input model.User
	}{
		{"空用户名", model.User{}},
		{"用户名过长", model.User{Username: "012345678901234567890123456789012345"}},
		{"非法状态", model.User{Username: "x", Status: "weird"}},
	}
	for _, c := range cases {
		if _, err := svc.CreateUser(c.input); err == nil {
			t.Errorf("%s: 期望校验失败但成功了", c.name)
		} else if !model.IsValidationError(err) {
			t.Errorf("%s: 期望 ValidationError，得到 %v", c.name, err)
		}
	}
	// 昵称缺省回落用户名
	u, err := svc.CreateUser(model.User{Username: "bob"})
	if err != nil {
		t.Fatal(err)
	}
	if u.Nickname != "bob" {
		t.Errorf("空昵称应回落用户名，得到 %s", u.Nickname)
	}
	// 重复用户名
	if _, err := svc.CreateUser(model.User{Username: "bob"}); err == nil {
		t.Fatal("重复用户名应冲突")
	}
}

func TestActivityLifecycle(t *testing.T) {
	svc := newTestService()
	a, err := svc.CreateActivity(model.Activity{
		Code: "LIFE", Name: "生命周期",
		StartTime: baseTime, EndTime: baseTime.AddDate(0, 0, 10),
	})
	if err != nil {
		t.Fatal(err)
	}
	if a.Status != model.ActivityDraft {
		t.Fatalf("新活动应为 draft，得到 %s", a.Status)
	}
	// 非法流转 draft -> finished
	if _, err := svc.TransitionActivity(a.ID, model.ActivityFinished); err == nil {
		t.Fatal("draft->finished 应被拒绝")
	}
	// draft -> active -> finished
	a, err = svc.TransitionActivity(a.ID, model.ActivityActive)
	if err != nil || a.Status != model.ActivityActive {
		t.Fatalf("激活失败: %v", err)
	}
	a, err = svc.TransitionActivity(a.ID, model.ActivityFinished)
	if err != nil || a.Status != model.ActivityFinished {
		t.Fatalf("结束失败: %v", err)
	}
	// 已结束不可编辑
	if _, err := svc.UpdateActivity(a.ID, "新名字", ""); err == nil {
		t.Fatal("已结束活动编辑应被拒绝")
	}
}

func TestActivityValidation(t *testing.T) {
	svc := newTestService()
	cases := []struct {
		name  string
		input model.Activity
	}{
		{"空编码", model.Activity{Name: "x", StartTime: baseTime, EndTime: baseTime.Add(time.Hour)}},
		{"空名称", model.Activity{Code: "C", StartTime: baseTime, EndTime: baseTime.Add(time.Hour)}},
		{"缺开始时间", model.Activity{Code: "C", Name: "x", EndTime: baseTime}},
		{"结束早于开始", model.Activity{Code: "C", Name: "x", StartTime: baseTime, EndTime: baseTime.Add(-time.Hour)}},
	}
	for _, c := range cases {
		if _, err := svc.CreateActivity(c.input); err == nil {
			t.Errorf("%s: 期望校验失败但成功了", c.name)
		}
	}
}

func TestCheckinHappyPath(t *testing.T) {
	svc := newTestService()
	activity, user := setupActiveActivity(t, svc)
	record, err := svc.Checkin(activity.ID, user.ID, baseTime)
	if err != nil {
		t.Fatal(err)
	}
	if record.Streak != 1 {
		t.Fatalf("首次签到连续天数应为 1，得到 %d", record.Streak)
	}
	if record.CheckinDate != "2026-08-10" {
		t.Fatalf("签到日期不对: %s", record.CheckinDate)
	}
	// 重复签到应被拒绝
	if _, err := svc.Checkin(activity.ID, user.ID, baseTime.Add(time.Hour)); err == nil {
		t.Fatal("同日重复签到应被拒绝")
	}
}

func TestCheckinStreakAccumulation(t *testing.T) {
	svc := newTestService()
	activity, user := setupActiveActivity(t, svc)
	// 连续三天签到
	for i := 0; i < 3; i++ {
		record, err := svc.Checkin(activity.ID, user.ID, baseTime.AddDate(0, 0, i))
		if err != nil {
			t.Fatalf("第 %d 天签到失败: %v", i+1, err)
		}
		if record.Streak != i+1 {
			t.Fatalf("第 %d 天连续天数应为 %d，得到 %d", i+1, i+1, record.Streak)
		}
	}
	// 断签一天后重置
	record, err := svc.Checkin(activity.ID, user.ID, baseTime.AddDate(0, 0, 4))
	if err != nil {
		t.Fatal(err)
	}
	if record.Streak != 1 {
		t.Fatalf("断签后连续天数应重置为 1，得到 %d", record.Streak)
	}
}

func TestCheckinGuards(t *testing.T) {
	svc := newTestService()
	activity, user := setupActiveActivity(t, svc)
	// 活动不存在
	if _, err := svc.Checkin("missing", user.ID, baseTime); err == nil {
		t.Fatal("活动不存在应报错")
	}
	// 用户不存在
	if _, err := svc.Checkin(activity.ID, "missing", baseTime); err == nil {
		t.Fatal("用户不存在应报错")
	}
	// 草稿活动不可签到
	draft, _ := svc.CreateActivity(model.Activity{
		Code: "DRAFT", Name: "草稿", StartTime: baseTime, EndTime: baseTime.AddDate(0, 0, 5),
	})
	if _, err := svc.Checkin(draft.ID, user.ID, baseTime); err == nil {
		t.Fatal("草稿活动签到应被拒绝")
	}
	// 活动未开始
	future, _ := svc.CreateActivity(model.Activity{
		Code: "FUTURE", Name: "未来", StartTime: baseTime.AddDate(0, 0, 5), EndTime: baseTime.AddDate(0, 0, 10),
	})
	_, _ = svc.TransitionActivity(future.ID, model.ActivityActive)
	if _, err := svc.Checkin(future.ID, user.ID, baseTime); err == nil {
		t.Fatal("活动未开始签到应被拒绝")
	}
	// 停用用户不可签到
	disabled, _ := svc.CreateUser(model.User{Username: "disabled", Status: model.UserDisabled})
	if _, err := svc.Checkin(activity.ID, disabled.ID, baseTime); err == nil {
		t.Fatal("停用用户签到应被拒绝")
	}
}

func TestRewardRulesGrantPoints(t *testing.T) {
	svc := newTestService()
	activity, user := setupActiveActivity(t, svc)
	// 每日基础 10 分
	if _, err := svc.CreateRewardRule(model.RewardRule{
		Name: "每日基础", Type: model.RewardDailyBase, Points: 10,
	}); err != nil {
		t.Fatal(err)
	}
	// 连续 3 天奖励 50 分
	if _, err := svc.CreateRewardRule(model.RewardRule{
		Name: "连续3天", Type: model.RewardStreakBonus, Threshold: 3, Points: 50,
	}); err != nil {
		t.Fatal(err)
	}
	// 第 1、2 天：各得 10 分
	for i := 0; i < 2; i++ {
		if _, err := svc.Checkin(activity.ID, user.ID, baseTime.AddDate(0, 0, i)); err != nil {
			t.Fatal(err)
		}
	}
	got, _ := svc.GetUser(user.ID)
	if got.Points != 20 {
		t.Fatalf("前两天应累计 20 分，得到 %d", got.Points)
	}
	// 第 3 天：基础 10 + 连续奖励 50 = 60
	record, err := svc.Checkin(activity.ID, user.ID, baseTime.AddDate(0, 0, 2))
	if err != nil {
		t.Fatal(err)
	}
	if record.Streak != 3 {
		t.Fatalf("第 3 天连续天数应为 3，得到 %d", record.Streak)
	}
	got, _ = svc.GetUser(user.ID)
	if got.Points != 80 {
		t.Fatalf("第 3 天后应累计 80 分，得到 %d", got.Points)
	}
	// 发放流水：3 条 base + 1 条 bonus
	grants, total, err := svc.ListRewardGrants(model.GrantFilter{UserID: user.ID}, 1, 100)
	if err != nil || total != 4 || len(grants) != 4 {
		t.Fatalf("期望 4 条发放流水，得到 total=%d err=%v", total, err)
	}
	bonusCount := 0
	for _, g := range grants {
		if g.Type == model.GrantBonus {
			bonusCount++
			if g.Points != 50 {
				t.Fatalf("连续奖励应为 50 分，得到 %d", g.Points)
			}
		}
	}
	if bonusCount != 1 {
		t.Fatalf("期望 1 条 bonus 流水，得到 %d", bonusCount)
	}
}

func TestInactiveRuleNotApplied(t *testing.T) {
	svc := newTestService()
	activity, user := setupActiveActivity(t, svc)
	rule, err := svc.CreateRewardRule(model.RewardRule{
		Name: "每日基础", Type: model.RewardDailyBase, Points: 10,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.UpdateRewardRule(rule.ID, "", model.RewardRuleInactive, 0, 0); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Checkin(activity.ID, user.ID, baseTime); err != nil {
		t.Fatal(err)
	}
	got, _ := svc.GetUser(user.ID)
	if got.Points != 0 {
		t.Fatalf("停用规则不应发放积分，得到 %d", got.Points)
	}
}

func TestActivityScopedRewardRule(t *testing.T) {
	svc := newTestService()
	a1, user := setupActiveActivity(t, svc)
	a2, err := svc.CreateActivity(model.Activity{
		Code: "OTHER", Name: "另一个", StartTime: baseTime.AddDate(0, 0, -1), EndTime: baseTime.AddDate(0, 0, 30),
	})
	if err != nil {
		t.Fatal(err)
	}
	a2, _ = svc.TransitionActivity(a2.ID, model.ActivityActive)
	// 规则只作用于活动一
	if _, err := svc.CreateRewardRule(model.RewardRule{
		Name: "仅活动一", Type: model.RewardDailyBase, ActivityID: a1.ID, Points: 10,
	}); err != nil {
		t.Fatal(err)
	}
	_, _ = svc.Checkin(a1.ID, user.ID, baseTime)
	_, _ = svc.Checkin(a2.ID, user.ID, baseTime)
	got, _ := svc.GetUser(user.ID)
	if got.Points != 10 {
		t.Fatalf("仅活动一应发放 10 分，得到 %d", got.Points)
	}
	// 关联不存在活动的规则应被拒绝
	if _, err := svc.CreateRewardRule(model.RewardRule{
		Name: "x", Type: model.RewardDailyBase, ActivityID: "missing", Points: 1,
	}); err == nil {
		t.Fatal("关联不存在活动的规则应被拒绝")
	}
}

func TestGetUserStreak(t *testing.T) {
	svc := newTestService()
	activity, user := setupActiveActivity(t, svc)
	for i := 0; i < 3; i++ {
		if _, err := svc.Checkin(activity.ID, user.ID, baseTime.AddDate(0, 0, i)); err != nil {
			t.Fatal(err)
		}
	}
	streak, err := svc.GetUserStreak(user.ID, activity.ID)
	if err != nil {
		t.Fatal(err)
	}
	if streak.TotalDays != 3 || streak.CurrentStreak != 3 || streak.MaxStreak != 3 {
		t.Fatalf("签到统计不对: %+v", streak)
	}
	// 不存在的用户/活动
	if _, err := svc.GetUserStreak("missing", activity.ID); err == nil {
		t.Fatal("用户不存在应报错")
	}
	if _, err := svc.GetUserStreak(user.ID, "missing"); err == nil {
		t.Fatal("活动不存在应报错")
	}
}

func TestDeleteGuards(t *testing.T) {
	svc := newTestService()
	activity, user := setupActiveActivity(t, svc)
	if _, err := svc.Checkin(activity.ID, user.ID, baseTime); err != nil {
		t.Fatal(err)
	}
	// 有签到记录的用户/活动不可删除
	if err := svc.DeleteUser(user.ID); err == nil {
		t.Fatal("有签到记录的用户删除应被拒绝")
	}
	if err := svc.DeleteActivity(activity.ID); err == nil {
		t.Fatal("有签到记录的活动删除应被拒绝")
	}
}

func TestStatsOverviewAndGrouping(t *testing.T) {
	svc := newTestService()
	activity, user := setupActiveActivity(t, svc)
	if _, err := svc.CreateRewardRule(model.RewardRule{
		Name: "每日基础", Type: model.RewardDailyBase, Points: 10,
	}); err != nil {
		t.Fatal(err)
	}
	user2, _ := svc.CreateUser(model.User{Username: "bob"})
	_, _ = svc.Checkin(activity.ID, user.ID, baseTime)
	_, _ = svc.Checkin(activity.ID, user2.ID, baseTime)
	_, _ = svc.Checkin(activity.ID, user.ID, baseTime.AddDate(0, 0, 1))

	overview := svc.Stats()
	if overview.UserCount != 2 || overview.ActivityCount != 1 {
		t.Fatalf("概览数量不对: %+v", overview)
	}
	if overview.CheckinCount != 3 || overview.ActiveActivities != 1 {
		t.Fatalf("签到/活动计数不对: %+v", overview)
	}
	if overview.TotalPoints != 30 || overview.GrantedPoints != 30 {
		t.Fatalf("积分统计不对: %+v", overview)
	}

	byActivity := svc.StatsByActivity()
	if len(byActivity) != 1 || byActivity[0].CheckinCount != 3 || byActivity[0].UserCount != 2 {
		t.Fatalf("按活动统计不对: %+v", byActivity)
	}

	byDay := svc.StatsByDay()
	if len(byDay) != 2 {
		t.Fatalf("期望 2 天统计，得到 %d", len(byDay))
	}
	if byDay[0].Date != "2026-08-10" || byDay[0].CheckinCount != 2 {
		t.Fatalf("第一天统计不对: %+v", byDay[0])
	}

	board := svc.Leaderboard(10)
	if len(board) != 2 || board[0].Points != 20 {
		t.Fatalf("排行榜不对: %+v", board)
	}
}

func TestListFilterAndPagination(t *testing.T) {
	svc := newTestService()
	activity, user := setupActiveActivity(t, svc)
	for i := 0; i < 5; i++ {
		if _, err := svc.Checkin(activity.ID, user.ID, baseTime.AddDate(0, 0, i)); err != nil {
			t.Fatal(err)
		}
	}
	items, total, err := svc.ListCheckinRecords(model.CheckinFilter{UserID: user.ID}, 1, 10)
	if err != nil || total != 5 || len(items) != 5 {
		t.Fatalf("按用户筛选失败: total=%d err=%v", total, err)
	}
	// 日期范围筛选
	_, total, err = svc.ListCheckinRecords(model.CheckinFilter{
		FromDate: "2026-08-11", ToDate: "2026-08-12",
	}, 1, 10)
	if err != nil || total != 2 {
		t.Fatalf("日期范围筛选期望 2 条，得到 %d", total)
	}
	// 分页
	items, total, err = svc.ListCheckinRecords(model.CheckinFilter{}, 2, 3)
	if err != nil || total != 5 || len(items) != 2 {
		t.Fatalf("分页期望 total=5 len=2，得到 total=%d len=%d", total, len(items))
	}
	// 用户列表筛选
	users, total, err := svc.ListUsers(model.UserFilter{Keyword: "ali"}, 1, 10)
	if err != nil || total != 1 || users[0].Username != "alice" {
		t.Fatalf("用户关键词筛选失败: total=%d err=%v", total, err)
	}
}

func TestGetUserCalendar(t *testing.T) {
	svc := newTestService()
	activity, user := setupActiveActivity(t, svc)
	// 8/10、8/11、8/12 连续签到
	for i := 0; i < 3; i++ {
		if _, err := svc.Checkin(activity.ID, user.ID, baseTime.AddDate(0, 0, i)); err != nil {
			t.Fatal(err)
		}
	}
	calendar, err := svc.GetUserCalendar(user.ID, activity.ID, "2026-08")
	if err != nil {
		t.Fatal(err)
	}
	if calendar.Count != 3 || len(calendar.Dates) != 3 {
		t.Fatalf("期望 3 个签到日，得到 %d", calendar.Count)
	}
	// 日期应升序
	for i := 1; i < len(calendar.Dates); i++ {
		if calendar.Dates[i-1] >= calendar.Dates[i] {
			t.Fatalf("日历日期应升序，得到 %v", calendar.Dates)
		}
	}
	if calendar.Dates[0] != "2026-08-10" || calendar.Dates[2] != "2026-08-12" {
		t.Fatalf("日历日期不对: %v", calendar.Dates)
	}
	// 查询无签到的月份应返回空
	empty, err := svc.GetUserCalendar(user.ID, activity.ID, "2026-09")
	if err != nil || empty.Count != 0 || len(empty.Dates) != 0 {
		t.Fatalf("空月份应返回 0 条，得到 %+v err=%v", empty, err)
	}
	// 月份格式非法
	if _, err := svc.GetUserCalendar(user.ID, activity.ID, "2026-8"); err == nil {
		t.Fatal("非法月份格式应被拒绝")
	} else if !model.IsValidationError(err) {
		t.Fatalf("期望 ValidationError，得到 %v", err)
	}
	// 用户/活动不存在
	if _, err := svc.GetUserCalendar("missing", activity.ID, "2026-08"); err == nil {
		t.Fatal("用户不存在应报错")
	}
	if _, err := svc.GetUserCalendar(user.ID, "missing", "2026-08"); err == nil {
		t.Fatal("活动不存在应报错")
	}
}

func TestCalendarMonthBoundary(t *testing.T) {
	svc := newTestService()
	user, err := svc.CreateUser(model.User{Username: "carol"})
	if err != nil {
		t.Fatal(err)
	}
	// 活动窗口覆盖 8/31 与 9/1
	activity, err := svc.CreateActivity(model.Activity{
		Code: "BOUNDARY", Name: "跨月",
		StartTime: time.Date(2026, 8, 30, 0, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2026, 9, 2, 23, 59, 59, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}
	activity, _ = svc.TransitionActivity(activity.ID, model.ActivityActive)
	aug31 := time.Date(2026, 8, 31, 9, 0, 0, 0, time.UTC)
	sep1 := time.Date(2026, 9, 1, 9, 0, 0, 0, time.UTC)
	if _, err := svc.Checkin(activity.ID, user.ID, aug31); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Checkin(activity.ID, user.ID, sep1); err != nil {
		t.Fatal(err)
	}
	// 查 8 月只应含 8/31
	aug, err := svc.GetUserCalendar(user.ID, activity.ID, "2026-08")
	if err != nil {
		t.Fatal(err)
	}
	if aug.Count != 1 || aug.Dates[0] != "2026-08-31" {
		t.Fatalf("8 月日历应只含 8/31，得到 %v", aug.Dates)
	}
	// 查 9 月只应含 9/1
	sep, err := svc.GetUserCalendar(user.ID, activity.ID, "2026-09")
	if err != nil {
		t.Fatal(err)
	}
	if sep.Count != 1 || sep.Dates[0] != "2026-09-01" {
		t.Fatalf("9 月日历应只含 9/1，得到 %v", sep.Dates)
	}
}
