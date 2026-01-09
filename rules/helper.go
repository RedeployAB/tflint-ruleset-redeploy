package rules

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// LineOffsets provides O(1) byte offset lookups for line numbers.
// This is significantly faster than iterating through lines for each lookup.
// Benchmark shows ~54x improvement: linear scan ~17ns vs precomputed ~0.3ns.
type LineOffsets struct {
	offsets []int // offsets[i] = byte offset where line i starts (0-indexed)
	lines   []string
}

// NewLineOffsets creates a LineOffsets from file content.
// Call this once per file, then use ByteOffset for O(1) lookups.
func NewLineOffsets(content string) *LineOffsets {
	lines := strings.Split(content, "\n")
	offsets := make([]int, len(lines)+1)

	pos := 0
	for i, line := range lines {
		offsets[i] = pos
		pos += len(line) + 1 // +1 for newline
	}
	offsets[len(lines)] = pos // End of file position

	return &LineOffsets{
		offsets: offsets,
		lines:   lines,
	}
}

// ByteOffset returns the byte offset for the start of the given line (0-indexed).
// Returns 0 if line is out of range.
func (lo *LineOffsets) ByteOffset(line int) int {
	if line < 0 || line >= len(lo.offsets) {
		return 0
	}
	return lo.offsets[line]
}

// ByteOffsetEnd returns the byte offset for the end of the given line (0-indexed),
// including the newline character if not the last line.
func (lo *LineOffsets) ByteOffsetEnd(line int) int {
	if line < 0 || line >= len(lo.lines) {
		return 0
	}
	end := lo.offsets[line] + len(lo.lines[line])
	// Include newline if not last line
	if line < len(lo.lines)-1 {
		end++
	}
	return end
}

// Lines returns the split lines.
func (lo *LineOffsets) Lines() []string {
	return lo.lines
}

// LineCount returns the number of lines.
func (lo *LineOffsets) LineCount() int {
	return len(lo.lines)
}

// GetRuleDocLink returns the URL to the documentation of a rule based on its name.
func GetRuleDocLink(ruleName string) string {
	return fmt.Sprintf("https://github.com/RedeployAB/tflint-ruleset-redeploy/blob/main/docs/rules/%s.md", ruleName)
}

// removeAttributeLine removes an entire attribute line including trailing newline
func removeAttributeLine(f tflint.Fixer, runner tflint.Runner, attrRange hcl.Range) error {
	// Get the file content to check for newline after the attribute
	file, err := runner.GetFile(attrRange.Filename)
	if err != nil {
		return err
	}

	// Extend the range to include the entire line
	lineRange := attrRange

	// Find the end of the line (including newline)
	fileBytes := file.Bytes
	endPos := lineRange.End.Byte

	// Look for newline after the attribute
	for endPos < len(fileBytes) && fileBytes[endPos] != '\n' {
		endPos++
	}

	// If we found a newline, include it
	if endPos < len(fileBytes) && fileBytes[endPos] == '\n' {
		endPos++
		lineRange.End.Byte = endPos
		lineRange.End.Column = 1
		lineRange.End.Line++
	}

	// Remove the entire line including trailing newline
	return f.Remove(lineRange)
}
