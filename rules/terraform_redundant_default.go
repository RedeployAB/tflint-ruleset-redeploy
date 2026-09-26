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

func (*TerraformRedundantDefaultRule) Name() string {
	return "terraform_redundant_default"
}

func (*TerraformRedundantDefaultRule) Enabled() bool {
	return true
}

func (*TerraformRedundantDefaultRule) Severity() tflint.Severity {
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

func (r *TerraformRedundantDefaultRule) processBody(
	body *hclsyntax.Body,
	config *redundantDefaultConfig,
	runner tflint.Runner,
) error {
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

func (r *TerraformRedundantDefaultRule) checkBlock(
	block *hclsyntax.Block,
	config *redundantDefaultConfig,
	runner tflint.Runner,
) error {
	for _, name := range redundantDefaultArgNames(block.Type, config) {
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
		err = runner.EmitIssueWithFix(
			r,
			name+" should not be set to false (omit instead)",
			attr.Range(),
			func(f tflint.Fixer) error {
				return removeAttributeLine(f, runner, attr.Range())
			},
		)
		if err != nil {
			return err
		}
	}
	return nil
}

// redundantDefaultArgNames returns the enabled argument names to check for
// the given block type, or nil if the block type is not checked.
func redundantDefaultArgNames(blockType string, config *redundantDefaultConfig) []string {
	switch blockType {
	case TypeVariable, TypeOutput:
		return enabledRedundantDefaultArgs(
			redundantDefaultToggle{ArgSensitive, config.Sensitive},
			redundantDefaultToggle{ArgEphemeral, config.Ephemeral},
		)
	case ArgLifecycle:
		return enabledRedundantDefaultArgs(
			redundantDefaultToggle{ArgPreventDestroy, config.PreventDestroy},
			redundantDefaultToggle{ArgCreateBeforeDestroy, config.CreateBeforeDestroy},
		)
	default:
		return nil
	}
}

// redundantDefaultToggle pairs an argument name with its config toggle.
type redundantDefaultToggle struct {
	name    string
	enabled *bool
}

// enabledRedundantDefaultArgs returns, in order, the names of the toggles
// that are enabled.
func enabledRedundantDefaultArgs(toggles ...redundantDefaultToggle) []string {
	var names []string
	for _, toggle := range toggles {
		if checkEnabled(toggle.enabled) {
			names = append(names, toggle.name)
		}
	}
	return names
}

// checkEnabled reports whether a config toggle is on. An unset (nil) value
// defaults to enabled.
func checkEnabled(b *bool) bool {
	return b == nil || *b
}
