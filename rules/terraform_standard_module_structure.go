package rules

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

type TerraformStandardModuleStructureRule struct {
	tflint.DefaultRule
}

func NewTerraformStandardModuleStructureRule() *TerraformStandardModuleStructureRule {
	return &TerraformStandardModuleStructureRule{}
}

func (r *TerraformStandardModuleStructureRule) Name() string {
	return "terraform_standard_module_structure"
}

func (r *TerraformStandardModuleStructureRule) Enabled() bool {
	return true
}

func (r *TerraformStandardModuleStructureRule) Severity() tflint.Severity {
	return tflint.WARNING
}

func (r *TerraformStandardModuleStructureRule) Link() string {
	return ""
}

func (r *TerraformStandardModuleStructureRule) Check(runner tflint.Runner) error {
	requiredFiles := []string{
		"main.tf",
		"variables.tf",
		"locals.tf",
		"outputs.tf",
		"terraform.tf",
	}

	files, err := runner.GetFiles()
	if err != nil {
		return err
	}

	for _, required := range requiredFiles {
		if _, ok := files[required]; !ok {
			if err := runner.EmitIssue(
				r,
				fmt.Sprintf("Missing required file: %s", required),
				hcl.Range{Filename: required},
			); err != nil {
				return err
			}
		}
	}

	return nil
}
