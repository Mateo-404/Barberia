package timeutil

import (
	"testing"
	"time"
)

func TestFormatJavaToString(t *testing.T) {
	casos := []struct {
		in   string
		want string
	}{
		{"2026-10-22T15:00:00", "2026-10-22T15:00"},
		{"2026-10-22T15:30:00", "2026-10-22T15:30"},
		{"2026-10-22T15:00:30", "2026-10-22T15:00:30"},
		{"2026-10-22T15:00:00.500", "2026-10-22T15:00:00.5"},
	}
	for _, c := range casos {
		tm, err := ParseLocalDateTime(c.in)
		if err != nil {
			t.Fatalf("parse %s: %v", c.in, err)
		}
		if got := FormatJavaToString(tm); got != c.want {
			t.Errorf("FormatJavaToString(%s) = %s, want %s", c.in, got, c.want)
		}
	}
}

func TestFormatFechaHora(t *testing.T) {
	tm, _ := ParseLocalDateTime("2026-10-22T15:00:00")
	if got := FormatFechaHora(tm); got != "2026-10-22T15:00:00" {
		t.Errorf("got %s", got)
	}
}

func TestMinusMonthsClampaDia(t *testing.T) {
	// 31 de marzo − 1 mes = 28 de febrero (año no bisiesto).
	marzo31 := time.Date(2025, time.March, 31, 14, 30, 0, 0, time.Local)
	if got := MinusMonths(marzo31, 1); got.Day() != 28 || got.Month() != time.February {
		t.Errorf("31 mar 2025 − 1 = %v, want 28 feb", got)
	}
	// Año bisiesto.
	marzo31b := time.Date(2024, time.March, 31, 14, 30, 0, 0, time.Local)
	if got := MinusMonths(marzo31b, 1); got.Day() != 29 {
		t.Errorf("31 mar 2024 − 1 = %v, want 29 feb", got)
	}
	// Enero − 1 mes = diciembre del año anterior.
	enero := time.Date(2026, time.January, 15, 0, 0, 0, 0, time.Local)
	if got := MinusMonths(enero, 1); got.Month() != time.December || got.Year() != 2025 {
		t.Errorf("15 ene 2026 − 1 = %v, want 15 dic 2025", got)
	}
	if got := PlusMonths(enero, 1); got.Month() != time.February || got.Year() != 2026 {
		t.Errorf("15 ene 2026 + 1 = %v, want 15 feb 2026", got)
	}
}
