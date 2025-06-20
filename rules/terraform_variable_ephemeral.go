package rules

import (
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// TerraformVariableEphemeralRule checks if "ephemeral" is only set to true when used.
// If ephemeral is set to false, it should emit an error (false is the default).
type TerraformVariableEphemeralRule struct {
	tflint.DefaultRule
}

func NewTerraformVariableEphemeralRule() *TerraformVariableEphemeralRule {
	return &TerraformVariableEphemeralRule{}
}

func (r *TerraformVariableEphemeralRule) Name() string {
	return "terraform_variable_ephemeral"
}

func (r *TerraformVariableEphemeralRule) Enabled() bool {
	return true
}

func (r *TerraformVariableEphemeralRule) Severity() tflint.Severity {
	return tflint.ERROR
}

func (r *TerraformVariableEphemeralRule) Link() string {
	return GetRuleDocLink(r.Name())
}

func (r *TerraformVariableEphemeralRule) Check(runner tflint.Runner) error {
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

func (r *TerraformVariableEphemeralRule) processBody(
	body *hclsyntax.Body,
	filename string,
	runner tflint.Runner,
) error {
	for _, block := range body.Blocks {
		if strings.ToLower(block.Type) == TypeVariable {
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

func (r *TerraformVariableEphemeralRule) checkVariableBlock(
	block *hclsyntax.Block,
	runner tflint.Runner,
) error {
	var ephemeralAttr *hclsyntax.Attribute

	for _, attr := range block.Body.Attributes {
		if strings.EqualFold(attr.Name, "ephemeral") {
			ephemeralAttr = attr
			break
		}
	}
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
