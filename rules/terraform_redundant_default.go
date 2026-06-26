package rules

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// TerraformRedundantDefaultRule flags meta-arguments that are explicitly set to
// their default value of false, which is redundant noise. It covers `sensitive`
// and `ephemeral` on variables and outputs, and `prevent_destroy` and
// `create_before_destroy` inside lifecycle blocks. Each check can be disabled
// individually via the rule configuration.
type TerraformRedundantDefaultRule struct {
	tflint.DefaultRule
}

// redundantDefaultConfig toggles individual checks. An unset (nil) value keeps
// the check enabled; setting it to false disables that check.
type redundantDefaultConfig struct {
	Sensitive           *bool `hclext:"sensitive,optional"`
	Ephemeral           *bool `hclext:"ephemeral,optional"`
	PreventDestroy      *bool `hclext:"prevent_destroy,optional"`
	CreateBeforeDestroy *bool `hclext:"create_before_destroy,optional"`
}

func NewTerraformRedundantDefaultRule() *TerraformRedundantDefaultRule {
	return &TerraformRedundantDefaultRule{}
}

func (r *TerraformRedundantDefaultRule) Name() string {
	return "terraform_redundant_default"
}

func (r *TerraformRedundantDefaultRule) Enabled() bool {
	return true
}

func (r *TerraformRedundantDefaultRule) Severity() tflint.Severity {
	return tflint.ERROR
}

func (r *TerraformRedundantDefaultRule) Link() string {
	return GetRuleDocLink(r.Name())
}

func (r *TerraformRedundantDefaultRule) Check(runner tflint.Runner) error {
	config := redundantDefaultConfig{}
	if err := runner.DecodeRuleConfig(r.Name(), &config); err != nil {
		return err
	}

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
			if err := r.processBody(body, &config, runner); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *TerraformRedundantDefaultRule) processBody(body *hclsyntax.Body, config *redundantDefaultConfig, runner tflint.Runner) error {
	for _, block := range body.Blocks {
		if err := r.checkBlock(block, config, runner); err != nil {
			return err
		}
		if err := r.processBody(block.Body, config, runner); err != nil {
			return err
		}
	}
	return nil
}

func (r *TerraformRedundantDefaultRule) checkBlock(block *hclsyntax.Block, config *redundantDefaultConfig, runner tflint.Runner) error {
	var names []string
	switch block.Type {
	case TypeVariable, TypeOutput:
		if checkEnabled(config.Sensitive) {
			names = append(names, ArgSensitive)
		}
		if checkEnabled(config.Ephemeral) {
			names = append(names, ArgEphemeral)
		}
	case ArgLifecycle:
		if checkEnabled(config.PreventDestroy) {
			names = append(names, ArgPreventDestroy)
		}
		if checkEnabled(config.CreateBeforeDestroy) {
			names = append(names, ArgCreateBeforeDestroy)
		}
	default:
		return nil
	}

	for _, name := range names {
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

// checkEnabled reports whether a config toggle is on. An unset (nil) value
// defaults to enabled.
func checkEnabled(b *bool) bool {
	return b == nil || *b
}
