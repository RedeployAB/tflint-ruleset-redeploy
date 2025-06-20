package rules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func TestTerraformBasicModuleStructure(t *testing.T) {
	tests := []struct {
		Name   string
		Files  map[string]string
		Issues helper.Issues
	}{
		{
			Name: "all required files exist",
			Files: map[string]string{
				"main.tf":      `# dummy`,
				"variables.tf": `# dummy`,
				"locals.tf":    `# dummy`,
				"outputs.tf":   `# dummy`,
				"terraform.tf": `# dummy`,
			},
			Issues: helper.Issues{},
		},
		{
			Name: "missing locals.tf",
			Files: map[string]string{
				"main.tf":      `# dummy`,
				"variables.tf": `# dummy`,
				"outputs.tf":   `# dummy`,
				"terraform.tf": `# dummy`,
			},
			Issues: helper.Issues{
				{
					Rule:    NewTerraformBasicModuleStructureRule(),
					Message: "Missing required file: locals.tf",
					Range:   hcl.Range{Filename: "locals.tf"},
				},
			},
		},
		{
			Name: "missing multiple files",
			Files: map[string]string{
				"main.tf": `# dummy`,
			},
			Issues: helper.Issues{
				{
					Rule:    NewTerraformBasicModuleStructureRule(),
					Message: "Missing required file: variables.tf",
					Range:   hcl.Range{Filename: "variables.tf"},
				},
				{
					Rule:    NewTerraformBasicModuleStructureRule(),
					Message: "Missing required file: locals.tf",
					Range:   hcl.Range{Filename: "locals.tf"},
				},
				{
					Rule:    NewTerraformBasicModuleStructureRule(),
					Message: "Missing required file: outputs.tf",
					Range:   hcl.Range{Filename: "outputs.tf"},
				},
				{
					Rule:    NewTerraformBasicModuleStructureRule(),
					Message: "Missing required file: terraform.tf",
					Range:   hcl.Range{Filename: "terraform.tf"},
				},
			},
		},
		{
			Name: "files in subdirectories should not count",
			Files: map[string]string{
				"main.tf":                     `# dummy`,
				"modules/submodule/main.tf":   `# dummy`,
				"modules/submodule/locals.tf": `# dummy`,
			},
			Issues: helper.Issues{
				{
					Rule:    NewTerraformBasicModuleStructureRule(),
					Message: "Missing required file: variables.tf",
					Range:   hcl.Range{Filename: "variables.tf"},
				},
				{
					Rule:    NewTerraformBasicModuleStructureRule(),
					Message: "Missing required file: locals.tf",
					Range:   hcl.Range{Filename: "locals.tf"},
				},
				{
					Rule:    NewTerraformBasicModuleStructureRule(),
					Message: "Missing required file: outputs.tf",
					Range:   hcl.Range{Filename: "outputs.tf"},
				},
				{
					Rule:    NewTerraformBasicModuleStructureRule(),
					Message: "Missing required file: terraform.tf",
					Range:   hcl.Range{Filename: "terraform.tf"},
				},
			},
		},
		{
			Name: "files with module path prefix (simulating --chdir) - all files present",
			Files: map[string]string{
				"modules/terraform-aws-monitor-ec2-metrics/main.tf":      `# dummy`,
				"modules/terraform-aws-monitor-ec2-metrics/variables.tf": `# dummy`,
				"modules/terraform-aws-monitor-ec2-metrics/locals.tf":    `# dummy`,
				"modules/terraform-aws-monitor-ec2-metrics/outputs.tf":   `# dummy`,
				"modules/terraform-aws-monitor-ec2-metrics/terraform.tf": `# dummy`,
			},
			Issues: helper.Issues{},
		},
		{
			Name: "files with module path prefix (simulating --chdir) - missing some files",
			Files: map[string]string{
				"modules/terraform-aws-monitor-ec2-metrics/main.tf":      `# dummy`,
				"modules/terraform-aws-monitor-ec2-metrics/variables.tf": `# dummy`,
			},
			Issues: helper.Issues{
				{
					Rule:    NewTerraformBasicModuleStructureRule(),
					Message: "Missing required file: locals.tf",
					Range:   hcl.Range{Filename: "locals.tf"},
				},
				{
					Rule:    NewTerraformBasicModuleStructureRule(),
					Message: "Missing required file: outputs.tf",
					Range:   hcl.Range{Filename: "outputs.tf"},
				},
				{
					Rule:    NewTerraformBasicModuleStructureRule(),
					Message: "Missing required file: terraform.tf",
					Range:   hcl.Range{Filename: "terraform.tf"},
				},
			},
		},
		{
			Name: "files with area pattern not accepted",
			Files: map[string]string{
				"main.area.tf":      `# dummy`,
				"variables.area.tf": `# dummy`,
				"locals.area.tf":    `# dummy`,
				"outputs.area.tf":   `# dummy`,
				"terraform.tf":      `# dummy`,
			},
			Issues: helper.Issues{
				{
					Rule:    NewTerraformBasicModuleStructureRule(),
					Message: "Missing required file: main.tf",
					Range:   hcl.Range{Filename: "main.tf"},
				},
				{
					Rule:    NewTerraformBasicModuleStructureRule(),
					Message: "Missing required file: variables.tf",
					Range:   hcl.Range{Filename: "variables.tf"},
				},
				{
					Rule:    NewTerraformBasicModuleStructureRule(),
					Message: "Missing required file: locals.tf",
					Range:   hcl.Range{Filename: "locals.tf"},
				},
				{
					Rule:    NewTerraformBasicModuleStructureRule(),
					Message: "Missing required file: outputs.tf",
					Range:   hcl.Range{Filename: "outputs.tf"},
				},
			},
		},
	}

	rule := NewTerraformBasicModuleStructureRule()

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runner := helper.TestRunner(t, tc.Files)

			if err := rule.Check(runner); err != nil {
				t.Fatalf("Unexpected error occurred: %s", err)
			}

			helper.AssertIssues(t, tc.Issues, runner.Issues)
		})
	}
}
