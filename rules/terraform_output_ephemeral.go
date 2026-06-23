package rules

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// TerraformOutputEphemeralRule checks if "ephemeral" is only set to true when used.
// If ephemeral is set to false, it should emit an error (false is the default).
type TerraformOutputEphemeralRule struct {
	tflint.DefaultRule
}

func NewTerraformOutputEphemeralRule() *TerraformOutputEphemeralRule {
	return &TerraformOutputEphemeralRule{}
}

func (r *TerraformOutputEphemeralRule) Name() string {
	return "terraform_output_ephemeral"
}

func (r *TerraformOutputEphemeralRule) Enabled() bool {
	return true
}

func (r *TerraformOutputEphemeralRule) Severity() tflint.Severity {
	return tflint.ERROR
}

func (r *TerraformOutputEphemeralRule) Link() string {
	return GetRuleDocLink(r.Name())
}

func (r *TerraformOutputEphemeralRule) Check(runner tflint.Runner) error {
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

func (r *TerraformOutputEphemeralRule) processBody(
	body *hclsyntax.Body,
	filename string,
	runner tflint.Runner,
) error {
	for _, block := range body.Blocks {
		// Only examine blocks of type "output" (block types are always lowercase in Terraform)
		if block.Type == TypeOutput {
			if err := r.checkOutputBlock(block, runner); err != nil {
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

func (r *TerraformOutputEphemeralRule) checkOutputBlock(
	block *hclsyntax.Block,
	runner tflint.Runner,
) error {
	// Attribute names are always lowercase in Terraform
	ephemeralAttr := block.Body.Attributes[ArgEphemeral]
	if ephemeralAttr == nil {
		return nil // ephemeral not defined => no problem
	}

	// Use the new expression utility for boolean evaluation
	value, isLiteral, err := EvaluateBoolLiteral(ephemeralAttr.Expr)
	if err != nil {
		return err
	}

	if isLiteral && !value {
		return runner.EmitIssue(
			r,
			"ephemeral should not be set to false (omit instead)",
			ephemeralAttr.Range(),
		)
	}

	return nil
}
