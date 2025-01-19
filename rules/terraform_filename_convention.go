package rules

import (
	"fmt"
	"path/filepath"
	"regexp"

	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

var filenamePattern = regexp.MustCompile(`^[a-z0-9]+(?:_[a-z0-9]+)*(?:\.[a-z0-9]+(?:_[a-z0-9]+)*)?\.tf$`)

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
	files, err := runner.GetFiles()
	if err != nil {
		return err
	}
	for filename := range files {
		if filepath.Ext(filename) != ".tf" {
			continue
		}

		base := filepath.Base(filename)
		if !filenamePattern.MatchString(base) {
			if err := runner.EmitIssue(
				r,
				fmt.Sprintf(
					"Terraform filename '%s' does not match the pattern '<name>.tf' or '<name>.<area>.tf' (all snake_case alphanumerics)",
					base,
				),
				hcl.Range{Filename: filename},
			); err != nil {
				return err
			}
		}
	}
	return nil
}
