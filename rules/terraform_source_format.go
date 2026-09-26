package rules

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

type TerraformSourceFormatRule struct {
	tflint.DefaultRule
}

func NewTerraformSourceFormatRule() *TerraformSourceFormatRule {
	return &TerraformSourceFormatRule{}
}

func (*TerraformSourceFormatRule) Name() string {
	return "terraform_source_format"
}

func (*TerraformSourceFormatRule) Enabled() bool {
	return true
}

func (*TerraformSourceFormatRule) Severity() tflint.Severity {
	return tflint.ERROR
}

func (r *TerraformSourceFormatRule) Link() string {
	return GetRuleDocLink(r.Name())
}

func (r *TerraformSourceFormatRule) Check(runner tflint.Runner) error {
	files, err := runner.GetFiles()
	if err != nil {
		return err
	}

	for filename, file := range files {
		if file == nil || file.Bytes == nil {
			continue
		}

		syntaxFile, diags := hclsyntax.ParseConfig(file.Bytes, filename, hcl.InitialPos)
		if diags.HasErrors() {
			continue
		}

		// Pre-split lines once per file to avoid repeated splitting
		lines := strings.Split(string(file.Bytes), "\n")

		if body, ok := syntaxFile.Body.(*hclsyntax.Body); ok {
			if err := r.processBody(body, filename, lines, runner); err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *TerraformSourceFormatRule) processBody(
	body *hclsyntax.Body,
	filename string,
	lines []string,
	runner tflint.Runner,
) error {
	for _, block := range body.Blocks {
		if block.Type == "module" {
			if err := r.checkModuleBlock(block, lines, runner); err != nil {
				return err
			}
		}
		if err := r.processBody(block.Body, filename, lines, runner); err != nil {
			return err
		}
	}
	return nil
}

func (r *TerraformSourceFormatRule) checkModuleBlock(
	block *hclsyntax.Block,
	lines []string,
	runner tflint.Runner,
) error {
	srcRange := block.Body.Range()

	startLine := srcRange.Start.Line - 1
	endLine := srcRange.End.Line - 1
	if endLine >= len(lines) {
		endLine = len(lines) - 1
	}

	sourceLine, versionLine := findSourceVersionLines(lines, startLine, endLine)

	lastOfTheTwo := Max(sourceLine, versionLine)
	if lastOfTheTwo < 0 {
		return nil
	}

	attrName := pickAttrName(sourceLine, versionLine, lastOfTheTwo)

	for nextLineIdx := lastOfTheTwo + 1; nextLineIdx <= endLine; nextLineIdx++ {
		nextText := strings.TrimSpace(lines[nextLineIdx])
		switch {
		case nextText == "":
			if !onlyCommentsUntilBlockEnd(lines, nextLineIdx+1, endLine) {
				return nil
			}
			return runner.EmitIssueWithFix(
				r,
				fmt.Sprintf("Unexpected blank line after '%s' when block ends", attrName),
				lineStartRange(srcRange.Filename, nextLineIdx),
				removeBlankLineFix(srcRange.Filename, lines, nextLineIdx),
			)
		case isSourceFormatComment(nextText):
			continue
		case nextText == "}":
			return nil
		default:
			if nextLineIdx > lastOfTheTwo+1 {
				return nil
			}
			return runner.EmitIssueWithFix(
				r,
				fmt.Sprintf("Expected a blank line after '%s'", attrName),
				lineStartRange(srcRange.Filename, nextLineIdx),
				insertBlankLineFix(srcRange.Filename, lines, nextLineIdx),
			)
		}
	}

	return nil
}

// findSourceVersionLines returns the 0-based indices of the last "source"
// and "version" attribute lines between startLine and endLine (inclusive),
// or -1 for an attribute that is not present.
func findSourceVersionLines(lines []string, startLine, endLine int) (sourceLine, versionLine int) {
	sourceLine = -1
	versionLine = -1

	for l := startLine; l <= endLine && l < len(lines); l++ {
		text := strings.TrimSpace(lines[l])
		if text == "" {
			continue
		}
		if strings.HasPrefix(text, "source ") || strings.HasPrefix(text, "source=") {
			sourceLine = l
		}
		if strings.HasPrefix(text, "version ") || strings.HasPrefix(text, "version=") {
			versionLine = l
		}
	}

	return sourceLine, versionLine
}

// isSourceFormatComment reports whether a trimmed line is a line comment.
func isSourceFormatComment(text string) bool {
	return strings.HasPrefix(text, "//") || strings.HasPrefix(text, "#")
}

// onlyCommentsUntilBlockEnd reports whether the lines from fromLine through
// endLine (inclusive) contain nothing but blank lines and comments before the
// block's closing brace, or before the end of the range.
func onlyCommentsUntilBlockEnd(lines []string, fromLine, endLine int) bool {
	for l := fromLine; l <= endLine; l++ {
		lineCheck := strings.TrimSpace(lines[l])
		if lineCheck == "" || isSourceFormatComment(lineCheck) {
			continue
		}
		return lineCheck == "}"
	}
	return true
}

// lineStartRange returns a zero-width range at the start of the 0-based line.
func lineStartRange(filename string, lineIdx int) hcl.Range {
	return hcl.Range{
		Filename: filename,
		Start:    hcl.Pos{Line: lineIdx + 1, Column: 1},
		End:      hcl.Pos{Line: lineIdx + 1, Column: 1},
	}
}

// lineByteOffset returns the byte offset of the start of the 0-based line.
func lineByteOffset(lines []string, lineIdx int) int {
	bytePos := 0
	for i := range lineIdx {
		bytePos += len(lines[i]) + 1 // +1 for newline
	}
	return bytePos
}

// removeBlankLineFix returns a fix that removes the blank 0-based line.
func removeBlankLineFix(filename string, lines []string, lineIdx int) func(tflint.Fixer) error {
	return func(f tflint.Fixer) error {
		// Range for the blank line (entire line including newline)
		lineStart := lineByteOffset(lines, lineIdx)
		lineEnd := lineStart + len(lines[lineIdx])
		if lineIdx < len(lines)-1 {
			lineEnd++ // Include the newline
		}

		removeRange := hcl.Range{
			Filename: filename,
			Start: hcl.Pos{
				Line:   lineIdx + 1,
				Column: 1,
				Byte:   lineStart,
			},
			End: hcl.Pos{
				Line:   lineIdx + 2,
				Column: 1,
				Byte:   lineEnd,
			},
		}

		// Remove the blank line
		return f.Remove(removeRange)
	}
}

// insertBlankLineFix returns a fix that inserts a blank line before the
// 0-based line, i.e. at the end of the line preceding it.
func insertBlankLineFix(filename string, lines []string, lineIdx int) func(tflint.Fixer) error {
	return func(f tflint.Fixer) error {
		// Calculate byte position for insertion: the start of the previous
		// line (with source/version) plus its length and newline
		bytePos := lineByteOffset(lines, lineIdx-1) + len(lines[lineIdx-1]) + 1

		insertPos := hcl.Range{
			Filename: filename,
			Start: hcl.Pos{
				Line:   lineIdx,
				Column: len(lines[lineIdx-1]) + 1,
				Byte:   bytePos - 1, // Position at end of previous line
			},
			End: hcl.Pos{
				Line:   lineIdx,
				Column: len(lines[lineIdx-1]) + 1,
				Byte:   bytePos - 1,
			},
		}

		// Insert a newline to create a blank line
		return f.InsertTextAfter(insertPos, "\n")
	}
}

// pickAttrName names the attribute on the last line of the two; "source"
// wins when both share that line or neither matches.
func pickAttrName(srcLine, verLine, last int) string {
	if last != srcLine && last == verLine {
		return "version"
	}
	return "source"
}
