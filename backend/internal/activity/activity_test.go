package activity

import (
	"testing"
	"time"

	"github.com/Nyhve/revu/backend/internal/apptime"
)

func TestGroupByDayInUserTimeZone(t *testing.T) {
	berlin, _ := time.LoadLocation("Europe/Berlin")
	clock := apptime.New(berlin)
	times := []time.Time{
		time.Date(2026, 10, 18, 10, 0, 0, 0, time.UTC),
		time.Date(2026, 10, 18, 21, 0, 0, 0, time.UTC),  // 23:00 Berlin, still Oct 18
		time.Date(2026, 10, 18, 22, 30, 0, 0, time.UTC), // 00:30 Berlin, Oct 19
		time.Date(2026, 10, 19, 8, 0, 0, 0, time.UTC),
	}
	got := groupByDay(clock, times)
	if len(got) != 2 || got[0].Count != 2 || got[1].Count != 2 {
		t.Fatalf("unexpected %+v", got)
	}
	if time.Time(got[1].Date).Format("2006-01-02") != "2026-10-19" {
		t.Fatalf("unexpected date %v", time.Time(got[1].Date))
	}
	if empty := groupByDay(clock, nil); empty == nil || len(empty) != 0 {
		t.Fatal("must return an empty, non-nil slice")
	}
}
