// Package timeutil holds the local date-time formats of the API contract
// (no time zone) and month arithmetic with day clamping.
package timeutil

import (
	"strings"
	"time"
)

const (
	// LayoutDateTime is the API date-time format: "2006-01-02T15:04:05"
	// (plus a fraction only when present).
	LayoutDateTime = "2006-01-02T15:04:05.999999999"
	layoutMinutes  = "2006-01-02T15:04"
	layoutSeconds  = "2006-01-02T15:04:05"
)

// FormatDateTime always includes seconds.
func FormatDateTime(t time.Time) string {
	return t.Format(LayoutDateTime)
}

// FormatShort omits ":ss" when seconds and fraction are zero, and trims the
// fraction to its precision ("2026-10-22T15:00").
func FormatShort(t time.Time) string {
	if t.Nanosecond() != 0 {
		return t.Format(LayoutDateTime)
	}
	if t.Second() != 0 {
		return t.Format(layoutSeconds)
	}
	return t.Format(layoutMinutes)
}

// ParseLocalDateTime accepts "2006-01-02T15:04[:05[.frac]]" without a zone.
func ParseLocalDateTime(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if t, err := time.ParseInLocation("2006-01-02T15:04:05.999999999", s, time.Local); err == nil {
		return t, nil
	}
	if t, err := time.ParseInLocation("2006-01-02T15:04:05", s, time.Local); err == nil {
		return t, nil
	}
	return time.ParseInLocation("2006-01-02T15:04", s, time.Local)
}

// MinusMonths subtracts months; when the day does not exist in the target
// month the last day of that month is used (31 Mar − 1 month = 28/29 Feb).
func MinusMonths(t time.Time, months int) time.Time { return addMonths(t, -months) }

// PlusMonths adds months with the same clamping as MinusMonths.
func PlusMonths(t time.Time, months int) time.Time { return addMonths(t, months) }

func addMonths(t time.Time, months int) time.Time {
	y, m, d := t.Date()
	total := int(m) - 1 + months
	y += total / 12
	m = time.Month(total%12 + 1)
	if m < time.January {
		m += 12
		y--
	}
	if last := daysIn(y, m); d > last {
		d = last
	}
	return time.Date(y, m, d, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
}

func daysIn(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

// StartOfDay truncates to 00:00:00.0.
func StartOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
