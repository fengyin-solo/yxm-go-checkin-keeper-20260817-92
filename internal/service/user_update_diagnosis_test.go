package service

import (
	"testing"

	"checkinkeeper/internal/model"
)

func TestUserUpdatePreservesExistingFields(t *testing.T) {
	svc := newTestService()
	user, err := svc.CreateUser(model.User{Username: "dora", Nickname: "Dora Explorer"})
	if err != nil {
		t.Fatal(err)
	}

	updated, err := svc.UpdateUser(user.ID, "", model.UserDisabled)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Nickname != "Dora Explorer" {
		t.Fatalf("status-only update should preserve nickname, got %q", updated.Nickname)
	}
	if updated.Status != model.UserDisabled {
		t.Fatalf("status should be disabled, got %q", updated.Status)
	}

	renamed, err := svc.UpdateUser(user.ID, "Dora Restored", "")
	if err != nil {
		t.Fatal(err)
	}
	if renamed.Status != model.UserDisabled {
		t.Fatalf("nickname-only update should preserve status, got %q", renamed.Status)
	}
	if renamed.Nickname != "Dora Restored" {
		t.Fatalf("nickname should be updated, got %q", renamed.Nickname)
	}
}
