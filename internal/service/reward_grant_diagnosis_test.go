package service

import (
	"testing"

	"checkinkeeper/internal/model"
)

func TestRewardGrantFilterAndOrderRegression(t *testing.T) {
	svc := newTestService()
	activity, user := setupActiveActivity(t, svc)
	if _, err := svc.CreateRewardRule(model.RewardRule{Name: "daily", Type: model.RewardDailyBase, Points: 10}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CreateRewardRule(model.RewardRule{Name: "streak3", Type: model.RewardStreakBonus, Threshold: 3, Points: 50}); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		if _, err := svc.Checkin(activity.ID, user.ID, baseTime.AddDate(0, 0, i)); err != nil {
			t.Fatal(err)
		}
	}

	bonus, total, err := svc.ListRewardGrants(model.GrantFilter{UserID: user.ID, Type: model.GrantBonus}, 1, 10)
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 || len(bonus) != 1 {
		t.Fatalf("bonus filter should return exactly one grant, total=%d grants=%v", total, bonus)
	}
	if bonus[0].Points != 50 || bonus[0].Streak != 3 {
		t.Fatalf("bonus grant should capture the third-day reward, got %+v", bonus[0])
	}

	all, total, err := svc.ListRewardGrants(model.GrantFilter{UserID: user.ID}, 1, 10)
	if err != nil {
		t.Fatal(err)
	}
	if total != 4 || len(all) != 4 {
		t.Fatalf("all grants should include three base rows and one bonus row, total=%d grants=%v", total, all)
	}
	if all[0].CreatedAt.Before(all[len(all)-1].CreatedAt) {
		t.Fatalf("grant list should be newest first, got first=%s last=%s", all[0].CreatedAt, all[len(all)-1].CreatedAt)
	}
}
