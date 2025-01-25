package rules

import (
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// TerraformVariableSensitiveRule checks for correct usage of the "sensitive" attribute in variable blocks.
// "sensitive" should only be set to true; if false, it should be omitted.
type TerraformVariableSensitiveRule struct {
	tflint.DefaultRule
}

func NewTerraformVariableSensitiveRule() *TerraformVariableSensitiveRule {
	return &TerraformVariableSensitiveRule{}
}

func (r *TerraformVariableSensitiveRule) Name() string {
	return "terraform_variable_sensitive"
}

func (r *TerraformVariableSensitiveRule) Enabled() bool {
	return true
}

func (r *TerraformVariableSensitiveRule) Severity() tflint.Severity {
	return tflint.ERROR
}

func (r *TerraformVariableSensitiveRule) Link() string {
	return ""
}

func (r *TerraformVariableSensitiveRule) Check(runner tflint.Runner) error {
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
			// skip parse errors
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

func (r *TerraformVariableSensitiveRule) processBody(
	body *hclsyntax.Body,
	filename string,
	runner tflint.Runner,
) error {
	for _, block := range body.Blocks {
		if strings.ToLower(block.Type) == "variable" {
			if err := r.checkVariableBlock(block, runner); err != nil {
				return err
			}
		}
		// Recurse
		if err := r.processBody(block.Body, filename, runner); err != nil {
			return err
		}
	}
	return nil
}

func (r *TerraformVariableSensitiveRule) checkVariableBlock(
	block *hclsyntax.Block,
	runner tflint.Runner,
) error {
	// We'll gather the "sensitive" attribute text
	var sensitiveAttr *hclsyntax.Attribute

	for _, attr := range block.Body.Attributes {
		if strings.EqualFold(attr.Name, "sensitive") {
			sensitiveAttr = attr
			break
		}
	}

	if sensitiveAttr == nil {
		return nil // No "sensitive" => fine
	}

	// Slice the raw text for "sensitive"
	files, err := runner.GetFiles()
	if err != nil {
		return err
	}
	fileBytes := files[block.DefRange().Filename].Bytes

	src := GetAttributeRawText(sensitiveAttr, fileBytes)
	src = strings.ToLower(strings.TrimSpace(src))

	// If we see 'false', that's invalid => prefer omitting "sensitive"
	if src == "false" {
		return runner.EmitIssue(
			r,
			"sensitive should not be set to false (omit instead)",
			sensitiveAttr.Range(),
		)
	}

	return nil
}
