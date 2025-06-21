package rules

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Used for reading test fixtures
func readFixture(t *testing.T, filename string) string {
	path := filepath.Join("testdata", filename)
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed reading %s: %v", path, err)
	}
	// Normalize line endings to Unix style (\n) for cross-platform compatibility
	return strings.ReplaceAll(string(content), "\r\n", "\n")
}

func TestMax(t *testing.T) {
	tests := []struct {
		a, b     int
		expected int
	}{
		{1, 2, 2},
		{2, 1, 2},
		{-1, 1, 1},
		{0, 0, 0},
	}

	for _, tc := range tests {
		got := Max(tc.a, tc.b)
		if got != tc.expected {
			t.Errorf("Max(%d, %d) = %d; want %d", tc.a, tc.b, got, tc.expected)
		}
	}
}
