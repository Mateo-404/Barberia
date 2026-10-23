// Package timeutil replica la semántica de LocalDateTime de Java que el
// contrato de API expone (formatos y aritmética de meses con clamp de día).
package timeutil

import (
	"strings"
	"time"
)

const (
	// LayoutFechaHora es el formato Jackson de LocalDateTime:
	// "2006-01-02T15:04:05" (+ fracción solo si existe).
	LayoutFechaHora = "2006-01-02T15:04:05.999999999"
	// LayoutMinutos replica LocalDateTime.toString() sin segundos.
	layoutMinutos  = "2006-01-02T15:04"
	layoutSegundos = "2006-01-02T15:04:05"
)

// FormatFechaHora serializa como Jackson (siempre con segundos).
func FormatFechaHora(t time.Time) string {
	return t.Format(LayoutFechaHora)
}

// FormatJavaToString replica LocalDateTime.toString(): omite ":ss" cuando
// los segundos y la fracción son cero, y recorta la fracción a su precisión.
func FormatJavaToString(t time.Time) string {
	if t.Nanosecond() != 0 {
		return t.Format(LayoutFechaHora)
	}
	if t.Second() != 0 {
		return t.Format(layoutSegundos)
	}
	return t.Format(layoutMinutos)
}

// ParseLocalDateTime acepta "2006-01-02T15:04[:05[.frac]]" sin zona.
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

// MinusMonths replica LocalDateTime.minusMonths: si el día no existe en el
// mes destino se usa el último día de ese mes (31 mar − 1 mes = 28/29 feb).
func MinusMonths(t time.Time, months int) time.Time { return addMonths(t, -months) }

// PlusMonths replica LocalDateTime.plusMonths.
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

// StartOfDay trunca a las 00:00:00.0.
func StartOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
