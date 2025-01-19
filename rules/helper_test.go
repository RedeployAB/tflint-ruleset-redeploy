package rules

import (
	"os"
	"path/filepath"
	"testing"
)

func readFixture(t *testing.T, filename string) string {
	path := filepath.Join("testdata", filename)
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed reading %s: %v", path, err)
	}
	return string(content)
}
