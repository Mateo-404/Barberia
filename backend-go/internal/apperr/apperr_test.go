package apperr

import "testing"

func TestNewFormateaMensaje(t *testing.T) {
	err := New(KindAlreadyExists, "Ya existe un cliente con el teléfono %s", "115")
	if err.Kind != KindAlreadyExists {
		t.Errorf("Kind = %v", err.Kind)
	}
	if err.Error() != "Ya existe un cliente con el teléfono 115" {
		t.Errorf("mensaje = %q", err.Error())
	}
}

func TestKindsDistintos(t *testing.T) {
	vistos := map[Kind]bool{}
	for _, k := range []Kind{KindBusinessRule, KindNotFound, KindAlreadyExists, KindInvalidCredentials, KindValidation, KindConflict} {
		if vistos[k] {
			t.Fatalf("kind duplicado: %v", k)
		}
		vistos[k] = true
	}
	if len(vistos) != 6 {
		t.Fatalf("kinds = %d, want 6", len(vistos))
	}
}
