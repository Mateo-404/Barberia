package turnos

import (
	"testing"
	"time"
)

// parseLocal replica el formato LocalDateTime del contrato (sin timezone).
func parseLocal(t *testing.T, s string) time.Time {
	t.Helper()
	tm, err := time.ParseInLocation("2006-01-02T15:04:05", s, time.Local)
	if err != nil {
		t.Fatal(err)
	}
	return tm
}

// validarFechaHora replica las reglas de TurnoService: 13:00–20:00 en
// intervalos de 30 minutos.
func TestValidarFechaHora(t *testing.T) {
	casos := []struct {
		fecha  string
		valida bool
	}{
		{"2026-10-22T13:00:00", true},  // apertura
		{"2026-10-22T13:30:00", true},  // media hora
		{"2026-10-22T16:00:00", true},  // media jornada
		{"2026-10-22T20:00:00", true},  // cierre
		{"2026-10-22T12:30:00", false}, // antes de la apertura
		{"2026-10-22T20:30:00", false}, // después del cierre
		{"2026-10-22T00:00:00", false}, // madrugada
		{"2026-10-22T14:15:00", false}, // no es :00 ni :30
		{"2026-10-22T14:45:00", false}, // no es :00 ni :30
	}
	for _, c := range casos {
		fecha := parseLocal(t, c.fecha)
		err := validarFechaHora(fecha)
		if c.valida && err != nil {
			t.Errorf("%s debería ser válida, got %v", c.fecha, err)
		}
		if !c.valida {
			if err == nil {
				t.Errorf("%s debería ser inválida", c.fecha)
			} else if err.Error() != "La hora debe estar entre 13:00 y 20:00." &&
				err.Error() != "La hora debe ser en intervalos de 30 minutos." {
				t.Errorf("mensaje inesperado: %q", err.Error())
			}
		}
	}
}

// TestTurnoDuplicadoMensaje replica el mensaje exacto del Java:
// "Ya existe un turno para la fecha y hora especificada. Fecha y hora: ..."
func TestTurnoDuplicadoMensaje(t *testing.T) {
	fecha := parseLocal(t, "2026-10-22T15:00:00")
	err := turnoDuplicado(fecha)
	want := "Ya existe un turno para la fecha y hora especificada. Fecha y hora: 2026-10-22T15:00"
	if err.Error() != want {
		t.Fatalf("mensaje = %q, want %q", err.Error(), want)
	}
}
