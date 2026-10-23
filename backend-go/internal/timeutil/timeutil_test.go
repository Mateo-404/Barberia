package timeutil

import (
	"testing"
	"time"
)

func TestFormatShort(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"2026-10-22T15:00:00", "2026-10-22T15:00"},
		{"2026-10-22T15:30:00", "2026-10-22T15:30"},
		{"2026-10-22T15:00:30", "2026-10-22T15:00:30"},
		{"2026-10-22T15:00:00.500", "2026-10-22T15:00:00.5"},
	}
	for _, c := range cases {
		tm, err := ParseLocalDateTime(c.in)
		if err != nil {
			t.Fatalf("parse %s: %v", c.in, err)
		}
		if got := FormatShort(tm); got != c.want {
			t.Errorf("FormatShort(%s) = %s, want %s", c.in, got, c.want)
		}
	}
}

func TestFormatDateTime(t *testing.T) {
	tm, _ := ParseLocalDateTime("2026-10-22T15:00:00")
	if got := FormatDateTime(tm); got != "2026-10-22T15:00:00" {
		t.Errorf("got %s", got)
	}
}

func TestMinusMonthsClampsDay(t *testing.T) {
	// 31 March − 1 month = 28 February (non-leap year).
	march31 := time.Date(2025, time.March, 31, 14, 30, 0, 0, time.Local)
	if got := MinusMonths(march31, 1); got.Day() != 28 || got.Month() != time.February {
		t.Errorf("31 Mar 2025 − 1 = %v, want 28 Feb", got)
	}
	// Leap year.
	march31Leap := time.Date(2024, time.March, 31, 14, 30, 0, 0, time.Local)
	if got := MinusMonths(march31Leap, 1); got.Day() != 29 {
		t.Errorf("31 Mar 2024 − 1 = %v, want 29 Feb", got)
	}
	// January − 1 month = December of the previous year.
	jan := time.Date(2026, time.January, 15, 0, 0, 0, 0, time.Local)
	if got := MinusMonths(jan, 1); got.Month() != time.December || got.Year() != 2025 {
		t.Errorf("15 Jan 2026 − 1 = %v, want 15 Dec 2025", got)
	}
	if got := PlusMonths(jan, 1); got.Month() != time.February || got.Year() != 2026 {
		t.Errorf("15 Jan 2026 + 1 = %v, want 15 Feb 2026", got)
	}
}
