package rules

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// TerraformLifecycleRedundantDefaultRule checks for lifecycle meta-arguments
// that are explicitly set to their default value of false. Like the other
// redundant-default rules in this ruleset, such declarations add noise and
// should be omitted.
type TerraformLifecycleRedundantDefaultRule struct {
	tflint.DefaultRule
}

func NewTerraformLifecycleRedundantDefaultRule() *TerraformLifecycleRedundantDefaultRule {
	return &TerraformLifecycleRedundantDefaultRule{}
}

func (r *TerraformLifecycleRedundantDefaultRule) Name() string {
	return "terraform_lifecycle_redundant_default"
}

func (r *TerraformLifecycleRedundantDefaultRule) Enabled() bool {
	return true
}

func (r *TerraformLifecycleRedundantDefaultRule) Severity() tflint.Severity {
	return tflint.ERROR
}

func (r *TerraformLifecycleRedundantDefaultRule) Link() string {
	return GetRuleDocLink(r.Name())
}

func (r *TerraformLifecycleRedundantDefaultRule) Check(runner tflint.Runner) error {
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
			if err := r.processBody(body, runner); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *TerraformLifecycleRedundantDefaultRule) processBody(body *hclsyntax.Body, runner tflint.Runner) error {
	for _, block := range body.Blocks {
		if block.Type == ArgLifecycle {
			if err := r.checkLifecycleBlock(block, runner); err != nil {
				return err
			}
		}
		if err := r.processBody(block.Body, runner); err != nil {
			return err
		}
	}
	return nil
}

func (r *TerraformLifecycleRedundantDefaultRule) checkLifecycleBlock(block *hclsyntax.Block, runner tflint.Runner) error {
	for _, name := range []string{ArgPreventDestroy, ArgCreateBeforeDestroy} {
		attr := block.Body.Attributes[name]
		if attr == nil {
			continue
		}
		// Only a literal false is redundant. Non-literal expressions (for
		// example referencing a variable) are left alone to avoid false
		// positives.
		value, isLiteral, err := EvaluateBoolLiteral(attr.Expr)
		if err != nil || !isLiteral || value {
			continue
		}
		if err := runner.EmitIssueWithFix(
			r,
			name+" should not be set to false (omit instead)",
			attr.Range(),
			func(f tflint.Fixer) error {
				return removeAttributeLine(f, runner, attr.Range())
			},
		); err != nil {
			return err
		}
	}
	return nil
}
