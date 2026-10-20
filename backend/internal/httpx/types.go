package httpx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"
)

const (
	dateLayout          = "2006-01-02"
	localDateTimeLayout = "2006-01-02T15:04:05.999999"
)

// Date is a calendar date serialised as "2006-01-02" (Java LocalDate).
type Date time.Time

func (d Date) Time() time.Time { return time.Time(d) }

func (d Date) MarshalJSON() ([]byte, error) {
	return []byte(`"` + time.Time(d).Format(dateLayout) + `"`), nil
}

func (d *Date) UnmarshalJSON(b []byte) error {
	if bytes.Equal(bytes.TrimSpace(b), []byte("null")) {
		return nil
	}
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return fmt.Errorf("date must be a string: %w", err)
	}
	t, err := time.Parse(dateLayout, s)
	if err != nil {
		return fmt.Errorf("invalid date %q", s)
	}
	*d = Date(t)
	return nil
}

// DatePtr converts an optional database date.
func DatePtr(t *time.Time) *Date {
	if t == nil {
		return nil
	}
	d := Date(*t)
	return &d
}

// TimePtr converts an optional request date for the database.
func (d *Date) TimePtr() *time.Time {
	if d == nil {
		return nil
	}
	t := time.Time(*d)
	return &t
}

// LocalDateTime is a zone-less timestamp, e.g. "2026-10-20T07:04:00.123456"
// (Java LocalDateTime). Values are UTC wall clock times.
type LocalDateTime time.Time

func (t LocalDateTime) MarshalJSON() ([]byte, error) {
	return []byte(`"` + time.Time(t).Format(localDateTimeLayout) + `"`), nil
}
