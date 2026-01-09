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

func TestLineOffsets(t *testing.T) {
	tests := []struct {
		name           string
		content        string
		expectedLines  []string
		expectedEOF    int
		lineOffsets    map[int]int // line -> expected byte offset
		lineOffsetEnds map[int]int // line -> expected end byte offset
	}{
		{
			name:          "multiple lines without trailing newline",
			content:       "line1\nline2\nline3",
			expectedLines: []string{"line1", "line2", "line3"},
			expectedEOF:   17,
			lineOffsets:   map[int]int{0: 0, 1: 6, 2: 12},
			lineOffsetEnds: map[int]int{0: 6, 1: 12, 2: 17}, // last line has no newline
		},
		{
			name:          "multiple lines with trailing newline",
			content:       "line1\nline2\nline3\n",
			expectedLines: []string{"line1", "line2", "line3", ""},
			expectedEOF:   18,
			lineOffsets:   map[int]int{0: 0, 1: 6, 2: 12, 3: 18},
			lineOffsetEnds: map[int]int{0: 6, 1: 12, 2: 18, 3: 18},
		},
		{
			name:           "single line without newline",
			content:        "hello",
			expectedLines:  []string{"hello"},
			expectedEOF:    5,
			lineOffsets:    map[int]int{0: 0},
			lineOffsetEnds: map[int]int{0: 5},
		},
		{
			name:           "single line with newline",
			content:        "hello\n",
			expectedLines:  []string{"hello", ""},
			expectedEOF:    6,
			lineOffsets:    map[int]int{0: 0, 1: 6},
			lineOffsetEnds: map[int]int{0: 6, 1: 6},
		},
		{
			name:           "empty content",
			content:        "",
			expectedLines:  []string{""},
			expectedEOF:    0,
			lineOffsets:    map[int]int{0: 0},
			lineOffsetEnds: map[int]int{0: 0},
		},
		{
			name:           "only newline",
			content:        "\n",
			expectedLines:  []string{"", ""},
			expectedEOF:    1,
			lineOffsets:    map[int]int{0: 0, 1: 1},
			lineOffsetEnds: map[int]int{0: 1, 1: 1},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			lo := NewLineOffsets(tc.content)

			// Check line count
			if lo.LineCount() != len(tc.expectedLines) {
				t.Errorf("LineCount() = %d; want %d", lo.LineCount(), len(tc.expectedLines))
			}

			// Check lines content
			lines := lo.Lines()
			for i, expected := range tc.expectedLines {
				if i >= len(lines) || lines[i] != expected {
					t.Errorf("Lines()[%d] = %q; want %q", i, lines[i], expected)
				}
			}

			// Check EOF offset (last element in offsets slice)
			eofOffset := lo.ByteOffset(len(tc.expectedLines))
			if eofOffset != tc.expectedEOF {
				t.Errorf("EOF offset = %d; want %d", eofOffset, tc.expectedEOF)
			}

			// Check individual line offsets
			for line, expected := range tc.lineOffsets {
				got := lo.ByteOffset(line)
				if got != expected {
					t.Errorf("ByteOffset(%d) = %d; want %d", line, got, expected)
				}
			}

			// Check individual line end offsets
			for line, expected := range tc.lineOffsetEnds {
				got := lo.ByteOffsetEnd(line)
				if got != expected {
					t.Errorf("ByteOffsetEnd(%d) = %d; want %d", line, got, expected)
				}
			}
		})
	}
}

func TestLineOffsets_OutOfRange(t *testing.T) {
	lo := NewLineOffsets("line1\nline2")

	// ByteOffset returns 0 for out of range
	if got := lo.ByteOffset(-1); got != 0 {
		t.Errorf("ByteOffset(-1) = %d; want 0", got)
	}
	if got := lo.ByteOffset(100); got != 0 {
		t.Errorf("ByteOffset(100) = %d; want 0", got)
	}

	// ByteOffsetEnd returns 0 for out of range
	if got := lo.ByteOffsetEnd(-1); got != 0 {
		t.Errorf("ByteOffsetEnd(-1) = %d; want 0", got)
	}
	if got := lo.ByteOffsetEnd(100); got != 0 {
		t.Errorf("ByteOffsetEnd(100) = %d; want 0", got)
	}
}
