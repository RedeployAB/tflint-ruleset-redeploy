package rules

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

//nolint:gocyclo
type TerraformSourceFormatRule struct {
	tflint.DefaultRule
}

func NewTerraformSourceFormatRule() *TerraformSourceFormatRule {
	return &TerraformSourceFormatRule{}
}

func (r *TerraformSourceFormatRule) Name() string {
	return "terraform_source_format"
}

func (r *TerraformSourceFormatRule) Enabled() bool {
	return true
}

func (r *TerraformSourceFormatRule) Severity() tflint.Severity {
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

		if body, ok := syntaxFile.Body.(*hclsyntax.Body); ok {
			if err := r.processBody(body, filename, runner); err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *TerraformSourceFormatRule) processBody(body *hclsyntax.Body, filename string, runner tflint.Runner) error {
	for _, block := range body.Blocks {
		if block.Type == "module" {
			if err := r.checkModuleBlock(block, runner); err != nil {
				return err
			}
		}
		if err := r.processBody(block.Body, filename, runner); err != nil {
			return err
		}
	}
	return nil
}

//nolint:gocyclo
func (r *TerraformSourceFormatRule) checkModuleBlock(block *hclsyntax.Block, runner tflint.Runner) error {
	srcRange := block.Body.Range()

	files, err := runner.GetFiles()
	if err != nil {
		return err
	}

	f, ok := files[srcRange.Filename]
	if !ok || f.Bytes == nil {
		return nil
	}

	lines := strings.Split(string(f.Bytes), "\n")

	startLine := srcRange.Start.Line - 1
	endLine := srcRange.End.Line - 1
	if endLine >= len(lines) {
		endLine = len(lines) - 1
	}

	sourceLine := -1
	versionLine := -1

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

	if sourceLine < 0 && versionLine < 0 {
		return nil
	}

	lastOfTheTwo := Max(sourceLine, versionLine)
	if lastOfTheTwo < 0 {
		return nil
	}

	nextLineIdx := lastOfTheTwo + 1
	if nextLineIdx > endLine {
		return nil
	}

	for nextLineIdx <= endLine {
		nextText := strings.TrimSpace(lines[nextLineIdx])
		switch {
		case nextText == "":
			tmp := nextLineIdx + 1
			for tmp <= endLine {
				lineCheck := strings.TrimSpace(lines[tmp])
				if lineCheck == "" || strings.HasPrefix(lineCheck, "//") || strings.HasPrefix(lineCheck, "#") {
					tmp++
					continue
				}
				if lineCheck == "}" {
					rng := hcl.Range{
						Filename: srcRange.Filename,
						Start:    hcl.Pos{Line: nextLineIdx + 1, Column: 1},
						End:      hcl.Pos{Line: nextLineIdx + 1, Column: 1},
					}
					return runner.EmitIssueWithFix(
						r,
						fmt.Sprintf("Unexpected blank line after '%s' when block ends", pickAttrName(sourceLine, versionLine, lastOfTheTwo)),
						rng,
						func(f tflint.Fixer) error {
							// Calculate byte position for the blank line
							bytePos := 0
							for i := 0; i < nextLineIdx; i++ {
								bytePos += len(lines[i]) + 1 // +1 for newline
							}

							// Range for the blank line (entire line including newline)
							lineStart := bytePos
							lineEnd := bytePos + len(lines[nextLineIdx])
							if nextLineIdx < len(lines)-1 {
								lineEnd++ // Include the newline
							}

							removeRange := hcl.Range{
								Filename: srcRange.Filename,
								Start: hcl.Pos{
									Line:   nextLineIdx + 1,
									Column: 1,
									Byte:   lineStart,
								},
								End: hcl.Pos{
									Line:   nextLineIdx + 2,
									Column: 1,
									Byte:   lineEnd,
								},
							}

							// Remove the blank line
							return f.Remove(removeRange)
						},
					)
				}
				return nil
			}
			rng := hcl.Range{
				Filename: srcRange.Filename,
				Start:    hcl.Pos{Line: nextLineIdx + 1, Column: 1},
				End:      hcl.Pos{Line: nextLineIdx + 1, Column: 1},
			}
			return runner.EmitIssueWithFix(
				r,
				fmt.Sprintf("Unexpected blank line after '%s' when block ends", pickAttrName(sourceLine, versionLine, lastOfTheTwo)),
				rng,
				func(f tflint.Fixer) error {
					// Calculate byte position for the blank line
					bytePos := 0
					for i := 0; i < nextLineIdx; i++ {
						bytePos += len(lines[i]) + 1 // +1 for newline
					}

					// Range for the blank line (entire line including newline)
					lineStart := bytePos
					lineEnd := bytePos + len(lines[nextLineIdx])
					if nextLineIdx < len(lines)-1 {
						lineEnd++ // Include the newline
					}

					removeRange := hcl.Range{
						Filename: srcRange.Filename,
						Start: hcl.Pos{
							Line:   nextLineIdx + 1,
							Column: 1,
							Byte:   lineStart,
						},
						End: hcl.Pos{
							Line:   nextLineIdx + 2,
							Column: 1,
							Byte:   lineEnd,
						},
					}

					// Remove the blank line
					return f.Remove(removeRange)
				},
			)
		case strings.HasPrefix(nextText, "//"), strings.HasPrefix(nextText, "#"):
			nextLineIdx++
			continue
		case nextText == "}":
			return nil
		default:
			if nextLineIdx > lastOfTheTwo+1 {
				return nil
			}
			rng := hcl.Range{
				Filename: srcRange.Filename,
				Start:    hcl.Pos{Line: nextLineIdx + 1, Column: 1},
				End:      hcl.Pos{Line: nextLineIdx + 1, Column: 1},
			}
			return runner.EmitIssueWithFix(
				r,
				fmt.Sprintf("Expected a blank line after '%s'", pickAttrName(sourceLine, versionLine, lastOfTheTwo)),
				rng,
				func(f tflint.Fixer) error {
					// Calculate byte position for insertion
					bytePos := 0
					for i := 0; i < nextLineIdx-1; i++ {
						bytePos += len(lines[i]) + 1 // +1 for newline
					}
					// Add the length of the previous line (line with source/version)
					bytePos += len(lines[nextLineIdx-1]) + 1

					insertPos := hcl.Range{
						Filename: srcRange.Filename,
						Start: hcl.Pos{
							Line:   nextLineIdx,
							Column: len(lines[nextLineIdx-1]) + 1,
							Byte:   bytePos - 1, // Position at end of previous line
						},
						End: hcl.Pos{
							Line:   nextLineIdx,
							Column: len(lines[nextLineIdx-1]) + 1,
							Byte:   bytePos - 1,
						},
					}

					// Insert a newline to create a blank line
					return f.InsertTextAfter(insertPos, "\n")
				},
			)
		}
	}

	return nil
}

func pickAttrName(srcLine, verLine, last int) string {
	switch last {
	case srcLine:
		return "source"
	case verLine:
		return "version"
	default:
		return "source"
	}
}
