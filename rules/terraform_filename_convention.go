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
	files := runner.GetFiles()
	for filename, file := range files {
		if filepath.Ext(filename) != ".tf" {
			continue
		}

		base := filepath.Base(filename)
		if !filenamePattern.MatchString(base) {
			if err := runner.EmitIssue(
				r,
				fmt.Sprintf("Terraform filename '%s' does not match the pattern '<name>.<area>.tf' with snake_case alphanumerics only", base),
				file.Range(),
			); err != nil {
				return err
			}
		}
	}
	return nil
}
