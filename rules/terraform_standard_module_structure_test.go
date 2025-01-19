package rules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func TestTerraformStandardModuleStructure(t *testing.T) {
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
					Rule:    NewTerraformStandardModuleStructureRule(),
					Message: "Missing required file: locals.tf",
					Range:   hcl.Range{Filename: "locals.tf"},
				},
			},
		},
	}

	rule := NewTerraformStandardModuleStructureRule()

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
