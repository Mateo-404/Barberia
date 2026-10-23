package settings

import (
	"os"
	"path/filepath"
	"testing"
)

// leerTestdata lee un fixture de backend-go/testdata y falla si no existe
// (un fixture vacío haría pasar los tests sin validar nada).
func leerTestdata(t *testing.T, nombre string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", "testdata", nombre))
	if err != nil {
		t.Fatalf("leer testdata %s: %v", nombre, err)
	}
	return b
}
