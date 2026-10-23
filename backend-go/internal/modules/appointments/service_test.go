package appointments

import (
	"errors"
	"testing"
	"time"

	"github.com/Mateo-404/Barberia/backend-go/internal/apperr"
)

// parseLocal parses the contract's local date-time format (no zone).
func parseLocal(t *testing.T, s string) time.Time {
	t.Helper()
	tm, err := time.ParseInLocation("2006-01-02T15:04:05", s, time.Local)
	if err != nil {
		t.Fatal(err)
	}
	return tm
}

// Bookings are allowed 13:00–20:00 on 30-minute intervals.
func TestValidateStartsAt(t *testing.T) {
	cases := []struct {
		at       string
		wantCode string
	}{
		{"2026-10-22T13:00:00", ""},
		{"2026-10-22T13:30:00", ""},
		{"2026-10-22T16:00:00", ""},
		{"2026-10-22T20:00:00", ""},
		{"2026-10-22T12:30:00", "appointment.outside_hours"},
		{"2026-10-22T20:30:00", "appointment.outside_hours"},
		{"2026-10-22T00:00:00", "appointment.outside_hours"},
		{"2026-10-22T14:15:00", "appointment.invalid_interval"},
		{"2026-10-22T14:45:00", "appointment.invalid_interval"},
	}
	for _, c := range cases {
		err := validateStartsAt(parseLocal(t, c.at))
		if c.wantCode == "" {
			if err != nil {
				t.Errorf("%s should be valid, got %v", c.at, err)
			}
			continue
		}
		var ae *apperr.Error
		if !errors.As(err, &ae) || ae.Code != c.wantCode {
			t.Errorf("%s: got %v, want code %s", c.at, err, c.wantCode)
		}
	}
}

func TestSlotTakenError(t *testing.T) {
	err := slotTakenError(parseLocal(t, "2026-10-22T15:00:00"))
	var ae *apperr.Error
	if !errors.As(err, &ae) {
		t.Fatalf("not an apperr: %v", err)
	}
	if ae.Kind != apperr.KindConflict || ae.Code != "appointment.slot_taken" || ae.Params["startsAt"] != "2026-10-22T15:00" {
		t.Fatalf("unexpected error: %+v", ae)
	}
}
