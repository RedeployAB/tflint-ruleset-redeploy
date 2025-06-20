package rules

import (
	"fmt"
	"path/filepath"

	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

type TerraformBasicModuleStructureRule struct {
	tflint.DefaultRule
}

func NewTerraformBasicModuleStructureRule() *TerraformBasicModuleStructureRule {
	return &TerraformBasicModuleStructureRule{}
}

func (r *TerraformBasicModuleStructureRule) Name() string {
	return "terraform_basic_module_structure"
}

func (r *TerraformBasicModuleStructureRule) Enabled() bool {
	return true
}

func (r *TerraformBasicModuleStructureRule) Severity() tflint.Severity {
	return tflint.WARNING
}

func (r *TerraformBasicModuleStructureRule) Link() string {
	return GetRuleDocLink(r.Name())
}

func (r *TerraformBasicModuleStructureRule) Check(runner tflint.Runner) error {
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

	// Determine the common directory prefix for all files (module root)
	var moduleRoot string
	for filename := range files {
		dir := filepath.Dir(filename)
		if moduleRoot == "" {
			moduleRoot = dir
		} else if len(dir) < len(moduleRoot) {
			moduleRoot = dir
		}
	}

	// Build a set of base filenames that exist in the module root
	foundFiles := make(map[string]bool)

	for filename := range files {
		// Get the directory of this file
		dir := filepath.Dir(filename)
		base := filepath.Base(filename)

		// Check if this file is in the module root directory
		if dir == moduleRoot || dir == "." {
			foundFiles[base] = true
		}
	}

	for _, required := range requiredFiles {
		if !foundFiles[required] {
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
