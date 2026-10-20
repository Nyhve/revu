package habit

import (
	"testing"
	"time"
)

func TestPercentMatchesJava(t *testing.T) {
	cases := []struct {
		completed, total int
		want             float64
	}{
		{1, 28, 3.6}, {0, 28, 0}, {28, 28, 100}, {5, 4, 100}, {3, 0, 0}, {1, 3, 33.3}, {2, 3, 66.7}, {1, 8, 12.5},
	}
	for _, c := range cases {
		if got := Percent(c.completed, c.total); got != c.want {
			t.Errorf("Percent(%d,%d) = %v, want %v", c.completed, c.total, got, c.want)
		}
	}
}

func TestWeekBounds(t *testing.T) {
	day := func(s string) time.Time { d, _ := time.Parse("2006-01-02", s); return d }
	cases := map[string][2]string{
		"2026-10-20": {"2026-10-19", "2026-10-25"}, // Tuesday
		"2026-10-19": {"2026-10-19", "2026-10-25"}, // Monday
		"2026-10-25": {"2026-10-19", "2026-10-25"}, // Sunday
		"2027-01-01": {"2026-12-28", "2027-01-03"}, // across new year
	}
	for in, want := range cases {
		from, to := WeekBounds(day(in))
		if from.Format("2006-01-02") != want[0] || to.Format("2006-01-02") != want[1] {
			t.Errorf("WeekBounds(%s) = %s..%s", in, from.Format("2006-01-02"), to.Format("2006-01-02"))
		}
	}
}
