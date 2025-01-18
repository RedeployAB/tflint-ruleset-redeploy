package rules

import (
	"fmt"
	"path/filepath"
	"regexp"

	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

var filenamePattern = regexp.MustCompile(`^[a-z0-9]+(?:_[a-z0-9]+)*\.[a-z0-9]+(?:_[a-z0-9]+)*\.tf$`)

type TerraformFilenameConventionRule struct {
	tflint.DefaultRule
}

func NewTerraformFilenameConventionRule() *TerraformFilenameConventionRule {
	return &TerraformFilenameConventionRule{}
}

func (r *TerraformFilenameConventionRule) Name() string {
	return "terraform_filename_convention"
}

func (r *TerraformFilenameConventionRule) Enabled() bool {
	return true
}

func (r *TerraformFilenameConventionRule) Severity() tflint.Severity {
	return tflint.ERROR
}

func (r *TerraformFilenameConventionRule) Link() string {
	return ""
}

func (r *TerraformFilenameConventionRule) Check(runner tflint.Runner) error {
	parser := runner.GetParser()
	if parser == nil {
		return nil
	}

	for _, file := range parser.Files() {
		// Only consider .tf files
		if filepath.Ext(file.Name) != ".tf" {
			continue
		}

		base := filepath.Base(file.Name)
		// Check pattern <name>.<area>.tf and ensure snake_case alphanumerics
		if !filenamePattern.MatchString(base) {
			if err := runner.EmitIssue(
				r,
				fmt.Sprintf("Terraform filename '%s' does not match the pattern '<name>.<area>.tf' with snake_case alphanumerics only", base),
				file.Range,
			); err != nil {
				return err
			}
		}
	}

	return nil
}
