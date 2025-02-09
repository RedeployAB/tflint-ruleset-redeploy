package rules

import (
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

type TerraformNoLeadingTrailingBlankLinesRule struct {
	tflint.DefaultRule
}

func NewTerraformNoLeadingTrailingBlankLinesRule() *TerraformNoLeadingTrailingBlankLinesRule {
	return &TerraformNoLeadingTrailingBlankLinesRule{}
}

func (r *TerraformNoLeadingTrailingBlankLinesRule) Name() string {
	return "terraform_no_leading_trailing_blank_lines"
}

func (r *TerraformNoLeadingTrailingBlankLinesRule) Enabled() bool {
	return true
}

func (r *TerraformNoLeadingTrailingBlankLinesRule) Severity() tflint.Severity {
	return tflint.ERROR
}

func (r *TerraformNoLeadingTrailingBlankLinesRule) Link() string {
	return GetRuleDocLink(r.Name())
}

func (r *TerraformNoLeadingTrailingBlankLinesRule) Check(runner tflint.Runner) error {
	files, err := runner.GetFiles()
	if err != nil {
		return err
	}

	for filename, hclFile := range files {
		if hclFile == nil || hclFile.Bytes == nil {
			continue
		}
		syntaxFile, diags := hclsyntax.ParseConfig(hclFile.Bytes, filename, hcl.InitialPos)
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

func (r *TerraformNoLeadingTrailingBlankLinesRule) processBody(
	body *hclsyntax.Body,
	filename string,
	runner tflint.Runner,
) error {
	for _, blk := range body.Blocks {
		if blk.Type == TypeResource || blk.Type == TypeModule {
			if err := r.checkBlock(blk, filename, runner); err != nil {
				return err
			}
		}
		// Recurse into child blocks
		if err := r.processBody(blk.Body, filename, runner); err != nil {
			return err
		}
	}
	return nil
}

func (r *TerraformNoLeadingTrailingBlankLinesRule) checkBlock(
	block *hclsyntax.Block,
	filename string,
	runner tflint.Runner,
) error {
	hclFile, err := runner.GetFile(filename)
	if err != nil {
		return err
	}
	if hclFile.Bytes == nil {
		return nil
	}
	lines := strings.Split(string(hclFile.Bytes), "\n")
	startLine := block.Body.Range().Start.Line - 1
	endLine := block.Body.Range().End.Line - 1

	// If there's no actual interior lines (i.e. empty block),
	// skip checks so that an empty resource {} won't trigger errors.
	if (endLine - startLine) <= 1 {
		return nil
	}

	// 1) Check line right after opening brace => must NOT be blank
	if startLine+1 < len(lines) {
		next := strings.TrimSpace(lines[startLine+1])
		if next == "" {
			rng := hcl.Range{
				Filename: filename,
				Start:    hcl.Pos{Line: startLine + 2, Column: 1},
				End:      hcl.Pos{Line: startLine + 2, Column: 1},
			}
			return runner.EmitIssue(r,
				"No blank line allowed immediately after '{'",
				rng,
			)
		}
	}

	// 2) Check line right before closing brace => must NOT be blank
	if endLine-1 >= 0 {
		prev := strings.TrimSpace(lines[endLine-1])
		if prev == "" {
			rng := hcl.Range{
				Filename: filename,
				Start:    hcl.Pos{Line: endLine, Column: 1},
				End:      hcl.Pos{Line: endLine, Column: 1},
			}
			return runner.EmitIssue(r,
				"No blank line allowed immediately before '}'",
				rng,
			)
		}
	}
	return nil
}
