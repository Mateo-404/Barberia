package settings

import (
	"os"
	"path/filepath"
	"testing"
)

// readTestdata reads a fixture from backend-go/testdata and fails when it is
// missing (an empty fixture would make tests pass without checking anything).
func readTestdata(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", "testdata", name))
	if err != nil {
		t.Fatalf("read testdata %s: %v", name, err)
	}
	return b
}
