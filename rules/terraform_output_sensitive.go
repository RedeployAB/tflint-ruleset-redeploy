package rules

import (
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// TerraformOutputSensitiveRule checks for correct usage of the "sensitive" attribute in output blocks.
// "sensitive" should only be set to true; if false, it should be omitted.
type TerraformOutputSensitiveRule struct {
	tflint.DefaultRule
}

func NewTerraformOutputSensitiveRule() *TerraformOutputSensitiveRule {
	return &TerraformOutputSensitiveRule{}
}

func (r *TerraformOutputSensitiveRule) Name() string {
	return "terraform_output_sensitive"
}

func (r *TerraformOutputSensitiveRule) Enabled() bool {
	return true
}

func (r *TerraformOutputSensitiveRule) Severity() tflint.Severity {
	return tflint.ERROR
}

func (r *TerraformOutputSensitiveRule) Link() string {
	return GetRuleDocLink(r.Name())
}

func (r *TerraformOutputSensitiveRule) Check(runner tflint.Runner) error {
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

func (r *TerraformOutputSensitiveRule) processBody(
	body *hclsyntax.Body,
	filename string,
	runner tflint.Runner,
) error {
	for _, block := range body.Blocks {
		if strings.ToLower(block.Type) == "output" {
			if err := r.checkOutputBlock(block, runner); err != nil {
				return err
			}
		}
		// Recurse into nested blocks
		if err := r.processBody(block.Body, filename, runner); err != nil {
			return err
		}
	}
	return nil
}

func (r *TerraformOutputSensitiveRule) checkOutputBlock(
	block *hclsyntax.Block,
	runner tflint.Runner,
) error {
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

	// Use the new expression utility for boolean evaluation
	value, isLiteral, err := EvaluateBoolLiteral(sensitiveAttr.Expr)
	if err != nil {
		return err
	}

	// If we see 'false', that's invalid => prefer omitting "sensitive"
	if isLiteral && !value {
		return runner.EmitIssueWithFix(
			r,
			"sensitive should not be set to false (omit instead)",
			sensitiveAttr.Range(),
			func(f tflint.Fixer) error {
				return removeAttributeLine(f, runner, sensitiveAttr.Range())
			},
		)
	}
	return nil
}
