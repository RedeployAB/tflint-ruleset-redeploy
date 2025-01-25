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
	return ""
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
		// Parse each file
		syntaxFile, diags := hclsyntax.ParseConfig(hclFile.Bytes, filename, hcl.InitialPos)
		if diags.HasErrors() {
			// Skip if parse error
			continue
		}
		if body, ok := syntaxFile.Body.(*hclsyntax.Body); ok {
			if err := r.checkBody(body, filename, runner); err != nil {
				return err
			}
		}
	}
	return nil
}

// checkBody scans lines for consecutive blank lines (2+) in any file section.
// We do not parse the HCL structure deeply; we simply check raw lines for
// multiple consecutive blank lines. If found, we emit an issue at that line.
func (r *TerraformSingleBlankLinesRule) checkBody(
	body *hclsyntax.Body,
	filename string,
	runner tflint.Runner,
) error {
	// Grab the file content lines
	content := runner.GetFileContent(filename)
	lines := strings.Split(content, "\n")

	blankCount := 0
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			blankCount++
		} else {
			blankCount = 0
		}
		// If we ever see more than 1 consecutive blank line => issue
		if blankCount > 1 {
			pos := hcl.Pos{Line: i + 1, Column: 1}
			rng := hcl.Range{
				Filename: filename,
				Start:    pos,
				End:      pos,
			}
			return runner.EmitIssue(r,
				fmt.Sprintf("More than one consecutive blank line found at line %d", i+1),
				rng,
			)
		}
	}
	return nil
}
