package service

import (
	"testing"

	"checkinkeeper/internal/model"
)

func TestUserListFilterSortAndPaginationRegression(t *testing.T) {
	svc := newTestService()
	alice, _ := svc.CreateUser(model.User{Username: "alice", Nickname: "Alpha"})
	bob, _ := svc.CreateUser(model.User{Username: "bob", Nickname: "Bravo"})
	carol, _ := svc.CreateUser(model.User{Username: "carol", Nickname: "Alpha Team"})
	_, _ = svc.UpdateUser(bob.ID, "", model.UserDisabled)
	alice.Points = 30
	bob.Points = 50
	carol.Points = 30
	_, _ = svc.UpdateUser(alice.ID, alice.Nickname, alice.Status)
	_, _ = svc.UpdateUser(bob.ID, bob.Nickname, bob.Status)
	_, _ = svc.UpdateUser(carol.ID, carol.Nickname, carol.Status)

	active, total, err := svc.ListUsers(model.UserFilter{Status: model.UserActive}, 1, 10)
	if err != nil {
		t.Fatal(err)
	}
	if total != 2 || len(active) != 2 {
		t.Fatalf("active filter should return 2 users, total=%d users=%v", total, active)
	}
	if active[0].Username != "alice" || active[1].Username != "carol" {
		t.Fatalf("active users with equal points should be sorted by username, got %s then %s", active[0].Username, active[1].Username)
	}

	keyword, total, err := svc.ListUsers(model.UserFilter{Keyword: "alpha"}, 1, 10)
	if err != nil {
		t.Fatal(err)
	}
	if total != 2 || keyword[0].Username != "alice" || keyword[1].Username != "carol" {
		t.Fatalf("keyword should match username or nickname and keep deterministic order, total=%d users=%v", total, keyword)
	}

	page, total, err := svc.ListUsers(model.UserFilter{}, 2, 2)
	if err != nil {
		t.Fatal(err)
	}
	if total != 3 || len(page) != 1 || page[0].Username != "carol" {
		t.Fatalf("second page should contain carol after points desc sort, total=%d page=%v", total, page)
	}
}
