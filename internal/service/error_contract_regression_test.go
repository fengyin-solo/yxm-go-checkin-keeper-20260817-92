package service

import (
	"testing"

	"checkinkeeper/internal/model"
)

func TestServiceValidationErrorsStayTyped(t *testing.T) {
	svc := newTestService()
	activity, user := setupActiveActivity(t, svc)

	if _, err := svc.CreateRewardRule(model.RewardRule{
		Name: "missing activity rule", Type: model.RewardDailyBase, ActivityID: "missing", Points: 1,
	}); err == nil || !model.IsValidationError(err) {
		t.Fatalf("missing activity on reward rule should be a validation error, got %v", err)
	}

	draft, err := svc.CreateActivity(model.Activity{
		Code: "DRAFT2", Name: "draft", StartTime: baseTime, EndTime: baseTime.AddDate(0, 0, 2),
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.TransitionActivity(draft.ID, model.ActivityFinished); err == nil || !model.IsValidationError(err) {
		t.Fatalf("illegal activity transition should be a validation error, got %v", err)
	}

	disabled, err := svc.CreateUser(model.User{Username: "disabled2", Status: model.UserDisabled})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Checkin(activity.ID, disabled.ID, baseTime); err == nil || !model.IsValidationError(err) {
		t.Fatalf("disabled user checkin should be a validation error, got %v", err)
	}

	rule, err := svc.CreateRewardRule(model.RewardRule{Name: "daily", Type: model.RewardDailyBase, Points: 3})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.UpdateRewardRule(rule.ID, "", "paused", 0, 0); err == nil || !model.IsValidationError(err) {
		t.Fatalf("invalid rule status should be a validation error, got %v", err)
	}

	if _, err := svc.Checkin("missing", user.ID, baseTime); err == nil {
		t.Fatal("missing activity should still return an error")
	}
}
