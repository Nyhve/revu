// Package apptime centralises how the application reads the clock.
//
// Timestamps are stored as UTC wall clock values in "timestamp without time
// zone" columns, exactly like the former Spring backend running on Cloud Run.
// Calendar decisions ("today", "this month") are made in the configured user
// time zone, so that a habit ticked at 00:30 in Berlin counts for that day.
package apptime

import "time"

type Clock struct {
	Location *time.Location
	now      func() time.Time
}

func New(loc *time.Location) *Clock {
	return &Clock{Location: loc, now: time.Now}
}

// NewFixed returns a clock frozen at t (for tests).
func NewFixed(loc *time.Location, t time.Time) *Clock {
	return &Clock{Location: loc, now: func() time.Time { return t }}
}

// Now returns the current instant as UTC, truncated to the database precision.
func (c *Clock) Now() time.Time {
	return c.now().UTC().Truncate(time.Microsecond)
}

// Today returns the current calendar date in the user time zone (as a UTC midnight value).
func (c *Clock) Today() time.Time {
	return DateOf(c.now().In(c.Location))
}

// DateOf strips the time of day, keeping the calendar date as seen in t's location.
func DateOf(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// LocalDate converts a stored UTC timestamp to the calendar date in the user time zone.
func (c *Clock) LocalDate(storedUTC time.Time) time.Time {
	return DateOf(storedUTC.In(c.Location))
}

// StartOfLocalDay returns the UTC instant at which the given local calendar date begins.
func (c *Clock) StartOfLocalDay(date time.Time) time.Time {
	y, m, d := date.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, c.Location).UTC()
}
