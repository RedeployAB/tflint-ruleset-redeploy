package rules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func TestTerraformVariableFileRule(t *testing.T) {
	tests := []struct {
		Name     string
		FileMap  map[string]string
		Expected helper.Issues
	}{
		{
			Name: "Valid - variable in variables.tf",
			FileMap: map[string]string{
				"variables.tf": `variable "foo" {}`,
			},
			Expected: helper.Issues{},
		},
		{
			Name: "Valid - variable in variables.prod.tf",
			FileMap: map[string]string{
				"variables.prod.tf": `variable "foo" {}`,
			},
			Expected: helper.Issues{},
		},
		{
			Name: "Invalid - variable in main.tf",
			FileMap: map[string]string{
				"main.tf": `variable "foo" {}`,
			},
			Expected: helper.Issues{
				{
					Rule:    NewTerraformVariableFileRule(),
					Message: `"variable" block must be placed in "variables.tf" or "variables.<area>.tf", not "main.tf"`,
					Range: hcl.Range{
						Filename: "main.tf",
						Start:    hcl.Pos{Line: 1, Column: 1},
						End:      hcl.Pos{Line: 1, Column: 15},
					},
				},
			},
		},
	}

	rule := NewTerraformVariableFileRule()
	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runner := helper.TestRunner(t, tc.FileMap)
			if err := rule.Check(runner); err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			helper.AssertIssues(t, tc.Expected, runner.Issues)
		})
	}
}
