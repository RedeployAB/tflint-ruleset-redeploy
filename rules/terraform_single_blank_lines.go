package rules

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

type TerraformSingleBlankLinesRule struct {
	tflint.DefaultRule
}

func NewTerraformSingleBlankLinesRule() *TerraformSingleBlankLinesRule {
	return &TerraformSingleBlankLinesRule{}
}

func (r *TerraformSingleBlankLinesRule) Name() string {
	return "terraform_single_blank_lines"
}

func (r *TerraformSingleBlankLinesRule) Enabled() bool {
	return true
}

func (r *TerraformSingleBlankLinesRule) Severity() tflint.Severity {
	return tflint.ERROR
}

func (r *TerraformSingleBlankLinesRule) Link() string {
	return GetRuleDocLink(r.Name())
}

func (r *TerraformSingleBlankLinesRule) Check(runner tflint.Runner) error {
	files, err := runner.GetFiles()
	if err != nil {
		return err
	}

	for filename, hclFile := range files {
		if hclFile == nil || hclFile.Bytes == nil {
			continue
		}
		// Validate the file can be parsed before checking
		_, diags := hclsyntax.ParseConfig(hclFile.Bytes, filename, hcl.InitialPos)
		if diags.HasErrors() {
			// Skip if parse error
			continue
		}
		if err := r.checkBody(filename, string(hclFile.Bytes), runner); err != nil {
			return err
		}
	}
	return nil
}

// checkBody scans lines for consecutive blank lines (2+) in any file section.
// We do not parse the HCL structure deeply; we simply check raw lines for
// multiple consecutive blank lines. If found, we emit an issue at that line.
func (r *TerraformSingleBlankLinesRule) checkBody(
	filename string,
	content string,
	runner tflint.Runner,
) error {
	// Grab the file content lines (split once and reuse)
	lines := strings.Split(content, "\n")

	blankCount := 0
	blankStartLine := -1

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			if blankCount == 0 {
				// First blank line in sequence
				blankStartLine = i
			}
			blankCount++
		} else {
			// Check if we just finished a sequence of multiple blank lines
			if blankCount > 1 {
				if err := r.emitIssueForMultipleBlankLines(
					runner, filename, lines, blankStartLine, i-1,
				); err != nil {
					return err
				}
			}
			blankCount = 0
			blankStartLine = -1
		}
	}

	// Check if file ends with multiple blank lines
	if blankCount > 1 {
		if err := r.emitIssueForMultipleBlankLines(
			runner, filename, lines, blankStartLine, len(lines)-1,
		); err != nil {
			return err
		}
	}

	return nil
}

func (r *TerraformSingleBlankLinesRule) emitIssueForMultipleBlankLines(
	runner tflint.Runner,
	filename string,
	lines []string,
	startLine int,
	endLine int,
) error {
	// Calculate byte positions for the range
	startByte := 0

	// Calculate start byte position
	for i := 0; i < startLine && i < len(lines); i++ {
		startByte += len(lines[i]) + 1 // +1 for newline
	}

	// Calculate end byte position
	endByte := startByte
	for i := startLine; i <= endLine && i < len(lines); i++ {
		endByte += len(lines[i])
		// Add 1 for newline, but not after the last line of the file
		if i < len(lines)-1 {
			endByte++
		}
	}

	// Adjust for the case where we want to keep one newline
	// The range should cover all the blank lines but we'll replace with one newline
	endLinePlus := endLine + 1
	if endLinePlus > len(lines) {
		endLinePlus = len(lines)
	}

	issueRange := hcl.Range{
		Filename: filename,
		Start: hcl.Pos{
			Line:   startLine + 1, // HCL uses 1-based line numbers
			Column: 1,
			Byte:   startByte,
		},
		End: hcl.Pos{
			Line:   endLinePlus,
			Column: 1,
			Byte:   endByte,
		},
	}

	return runner.EmitIssueWithFix(
		r,
		fmt.Sprintf("More than one consecutive blank line found at lines %d-%d", startLine+1, endLine+1),
		issueRange,
		func(f tflint.Fixer) error {
			// Replace multiple blank lines with a single newline
			// If at the end of file and no trailing newline, don't add one
			if endLine == len(lines)-1 && lines[endLine] == "" {
				// Last line is blank, check if there's content after startLine
				if startLine > 0 || (startLine == 0 && len(lines) > endLine+1) {
					return f.ReplaceText(issueRange, "\n")
				}
				// If it's only blank lines at the start, keep one
				return f.ReplaceText(issueRange, "\n")
			}
			return f.ReplaceText(issueRange, "\n")
		},
	)
}
