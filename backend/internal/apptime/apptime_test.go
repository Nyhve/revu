package apptime

import (
	"testing"
	"time"
)

func TestTodayUsesUserTimeZone(t *testing.T) {
	berlin, _ := time.LoadLocation("Europe/Berlin")
	// 22:30 UTC on Oct 19 is already 00:30 on Oct 20 in Berlin (CEST, UTC+2).
	c := NewFixed(berlin, time.Date(2026, 10, 19, 22, 30, 0, 0, time.UTC))
	if got := c.Today().Format("2006-01-02"); got != "2026-10-20" {
		t.Fatalf("Today = %s", got)
	}
	if got := c.Now(); got.Location() != time.UTC || got.Hour() != 22 {
		t.Fatalf("Now must be UTC: %v", got)
	}
	stored := time.Date(2026, 10, 19, 22, 30, 0, 0, time.UTC)
	if got := c.LocalDate(stored).Format("2006-01-02"); got != "2026-10-20" {
		t.Fatalf("LocalDate = %s", got)
	}
}
