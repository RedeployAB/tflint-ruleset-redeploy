package rules

import (
	"fmt"
	"path/filepath"
	"regexp"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// TerraformVariableFileRule enforces that any "variable" block appears only in files
// named "variables.tf" or "variables.*.tf".
type TerraformVariableFileRule struct {
	tflint.DefaultRule
}

// NewTerraformVariableFileRule creates a new rule instance.
func NewTerraformVariableFileRule() *TerraformVariableFileRule {
	return &TerraformVariableFileRule{}
}

// Name returns the rule name.
func (r *TerraformVariableFileRule) Name() string {
	return "terraform_variable_file"
}

// Enabled returns whether the rule is enabled by default.
func (r *TerraformVariableFileRule) Enabled() bool {
	return true
}

// Severity returns the severity of the rule.
func (r *TerraformVariableFileRule) Severity() tflint.Severity {
	return tflint.ERROR
}

// Link returns the rule's reference link.
func (r *TerraformVariableFileRule) Link() string {
	return GetRuleDocLink(r.Name())
}

var validVariablePattern = regexp.MustCompile(`^variables(\.[^.]+)?\.tf$`)

// Check checks whether "variable" blocks are in the correct files.
func (r *TerraformVariableFileRule) Check(runner tflint.Runner) error {
	files, err := runner.GetFiles()
	if err != nil {
		return err
	}

	for filename, hclFile := range files {
		if hclFile == nil || hclFile.Bytes == nil {
			continue
		}
		// If the file name matches valid variable file pattern, skip
		if validVariablePattern.MatchString(filepath.Base(filename)) {
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

func (r *TerraformVariableFileRule) processBody(
	body *hclsyntax.Body,
	filename string,
	runner tflint.Runner,
) error {
	for _, blk := range body.Blocks {
		// Block types are always lowercase in Terraform
		if blk.Type == TypeVariable {
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

func (r *TerraformVariableFileRule) emitIssue(
	runner tflint.Runner,
	rng hcl.Range,
	filename string,
) error {
	msg := fmt.Sprintf(
		`"variable" block must be placed in "variables.tf" or "variables.<area>.tf", not %q`,
		filepath.Base(filename),
	)
	return runner.EmitIssue(r, msg, rng)
}
