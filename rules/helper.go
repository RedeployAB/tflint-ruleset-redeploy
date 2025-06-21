package rules

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
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
