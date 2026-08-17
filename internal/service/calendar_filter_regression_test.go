package service

import (
	"reflect"
	"testing"
	"time"

	"checkinkeeper/internal/model"
)

func TestCalendarAndDateRangeRegression(t *testing.T) {
	svc := newTestService()
	activity, user := setupActiveActivity(t, svc)
	for i := 0; i < 5; i++ {
		if _, err := svc.Checkin(activity.ID, user.ID, baseTime.AddDate(0, 0, i)); err != nil {
			t.Fatalf("checkin day %d: %v", i+1, err)
		}
	}

	items, total, err := svc.ListCheckinRecords(model.CheckinFilter{
		UserID:   user.ID,
		FromDate: "2026-08-11",
		ToDate:   "2026-08-13",
	}, 1, 10)
	if err != nil {
		t.Fatal(err)
	}
	if total != 3 {
		t.Fatalf("date range should include exactly 3 records, got %d", total)
	}
	gotOrder := []string{items[0].CheckinDate, items[1].CheckinDate, items[2].CheckinDate}
	if !reflect.DeepEqual(gotOrder, []string{"2026-08-13", "2026-08-12", "2026-08-11"}) {
		t.Fatalf("list should be sorted newest first inside the requested range, got %v", gotOrder)
	}

	calendar, err := svc.GetUserCalendar(user.ID, activity.ID, "2026-08")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(calendar.Dates, []string{
		"2026-08-10", "2026-08-11", "2026-08-12", "2026-08-13", "2026-08-14",
	}) {
		t.Fatalf("calendar should return all August dates in ascending order, got %v", calendar.Dates)
	}

	late := time.Date(2026, 8, 14, 23, 59, 0, 0, time.UTC)
	latest, err := svc.GetUserStreak(user.ID, activity.ID)
	if err != nil {
		t.Fatal(err)
	}
	if latest.CurrentStreak != 5 {
		t.Fatalf("latest streak after %s should be 5, got %+v", late.Format(time.RFC3339), latest)
	}
}
