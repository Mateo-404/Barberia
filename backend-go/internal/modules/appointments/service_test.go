package appointments

import (
	"errors"
	"testing"
	"time"

	"github.com/Mateo-404/Barberia/backend-go/internal/apperr"
	"github.com/Mateo-404/Barberia/backend-go/internal/settings"
)

var businessTZ, _ = time.LoadLocation(settings.DefaultTimezone)

// at builds a local date-time of the business on 2030-10-22.
func at(hhmm string) time.Time {
	t, err := time.Parse("15:04", hhmm)
	if err != nil {
		panic(err)
	}
	return time.Date(2030, 10, 22, t.Hour(), t.Minute(), 0, 0, time.UTC)
}

func code(err error) string {
	var ae *apperr.Error
	if errors.As(err, &ae) {
		return ae.Code
	}
	return ""
}

// now is far before every test date, so only the hour rules are exercised.
var now = time.Date(2030, 1, 1, 0, 0, 0, 0, businessTZ)

func TestValidateStartsAtDefaultHours(t *testing.T) {
	b := settings.DefaultBusiness() // 13:00–20:00, 30-minute slots
	cases := []struct {
		at, wantCode string
	}{
		{"13:00", ""}, {"13:30", ""}, {"16:00", ""}, {"19:30", ""},
		{"20:00", "appointment.outside_hours"}, // would end after closing
		{"12:30", "appointment.outside_hours"}, {"20:30", "appointment.outside_hours"}, {"00:00", "appointment.outside_hours"},
		{"14:15", "appointment.invalid_interval"}, {"14:45", "appointment.invalid_interval"},
	}
	for _, c := range cases {
		if got := code(validateStartsAt(at(c.at), b, now)); got != c.wantCode {
			t.Errorf("%s: code = %q, want %q", c.at, got, c.wantCode)
		}
	}
}

func TestValidateStartsAtFollowsTheBusiness(t *testing.T) {
	b := settings.DefaultBusiness()
	b.Open, b.Close, b.Slot = "09:00", "12:00", 45
	for at_, want := range map[string]string{
		"09:00": "", "09:45": "", "10:30": "", "11:15": "",
		"09:30": "appointment.invalid_interval", // not on a 45-minute boundary
		"11:30": "appointment.outside_hours",    // would end after closing
		"12:00": "appointment.outside_hours", "08:45": "appointment.outside_hours",
	} {
		if got := code(validateStartsAt(at(at_), b, now)); got != want {
			t.Errorf("%s: code = %q, want %q", at_, got, want)
		}
	}
}

func TestValidateStartsAtPastAndNotice(t *testing.T) {
	b := settings.DefaultBusiness()
	b.Notice = 2
	clock := time.Date(2030, 10, 22, 14, 0, 0, 0, businessTZ)
	cases := []struct{ at, wantCode string }{
		{"13:00", "appointment.in_the_past"},
		{"15:30", "appointment.too_soon"}, // inside the 2-hour notice
		{"16:00", ""},
	}
	for _, c := range cases {
		if got := code(validateStartsAt(at(c.at), b, clock)); got != c.wantCode {
			t.Errorf("%s: code = %q, want %q", c.at, got, c.wantCode)
		}
	}
}

func TestSlotTakenError(t *testing.T) {
	err := slotTakenError(time.Date(2026, 10, 22, 15, 0, 0, 0, time.UTC))
	var ae *apperr.Error
	if !errors.As(err, &ae) {
		t.Fatalf("not an apperr: %v", err)
	}
	if ae.Kind != apperr.KindConflict || ae.Code != "appointment.slot_taken" || ae.Params["startsAt"] != "2026-10-22T15:00" {
		t.Fatalf("unexpected error: %+v", ae)
	}
}
