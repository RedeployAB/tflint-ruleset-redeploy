package rules

import (
	"fmt"
	"path/filepath"
	"regexp"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// TerraformProviderFileRule enforces that any "provider" block appears only in files
// named "providers.tf" or "providers.*.tf".
//
// Source: HashiCorp Style Guide - https://developer.hashicorp.com/terraform/language/style
// Source: AVM TFNFR27 - https://azure.github.io/Azure-Verified-Modules/specs/tf/res/
type TerraformProviderFileRule struct {
	tflint.DefaultRule
}

// NewTerraformProviderFileRule creates a new rule instance.
func NewTerraformProviderFileRule() *TerraformProviderFileRule {
	return &TerraformProviderFileRule{}
}

// Name returns the rule name.
func (r *TerraformProviderFileRule) Name() string {
	return "terraform_provider_file"
}

// Enabled returns whether the rule is enabled by default.
func (r *TerraformProviderFileRule) Enabled() bool {
	return true
}

// Severity returns the severity of the rule.
func (r *TerraformProviderFileRule) Severity() tflint.Severity {
	return tflint.ERROR
}

// Link returns the rule's reference link.
func (r *TerraformProviderFileRule) Link() string {
	return GetRuleDocLink(r.Name())
}

var validProviderPattern = regexp.MustCompile(`^providers(\.[^.]+)?\.tf$`)

// Check checks whether "provider" blocks are in the correct files.
func (r *TerraformProviderFileRule) Check(runner tflint.Runner) error {
	files, err := runner.GetFiles()
	if err != nil {
		return err
	}

	for filename, hclFile := range files {
		if hclFile == nil || hclFile.Bytes == nil {
			continue
		}
		// If the file name matches valid provider file pattern, skip
		if validProviderPattern.MatchString(filepath.Base(filename)) {
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

func (r *TerraformProviderFileRule) processBody(
	body *hclsyntax.Body,
	filename string,
	runner tflint.Runner,
) error {
	for _, blk := range body.Blocks {
		if blk.Type == TypeProvider {
			if err := r.emitIssue(runner, blk.DefRange(), filename); err != nil {
				return err
			}
		}
		// Recurse into nested blocks
		if err := r.processBody(blk.Body, filename, runner); err != nil {
			return err
		}
	}
	return nil
}

func (r *TerraformProviderFileRule) emitIssue(
	runner tflint.Runner,
	rng hcl.Range,
	filename string,
) error {
	msg := fmt.Sprintf(
		`"provider" block must be placed in "providers.tf" or "providers.<area>.tf", not %q`,
		filepath.Base(filename),
	)
	return runner.EmitIssue(r, msg, rng)
}
