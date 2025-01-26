package rules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func TestTerraformOutputFileRule(t *testing.T) {
	tests := []struct {
		Name     string
		FileMap  map[string]string
		Expected helper.Issues
	}{
		{
			Name: "Valid - output in outputs.tf",
			FileMap: map[string]string{
				"outputs.tf": `output "foo" { value = "bar" }`,
			},
			Expected: helper.Issues{},
		},
		{
			Name: "Valid - output in outputs.prod.tf",
			FileMap: map[string]string{
				"outputs.prod.tf": `output "foo" { value = "bar" }`,
			},
			Expected: helper.Issues{},
		},
		{
			Name: "Invalid - output in main.tf",
			FileMap: map[string]string{
				"main.tf": `output "foo" { value = "bar" }`,
			},
			Expected: helper.Issues{
				{
					Rule:    NewTerraformOutputFileRule(),
					Message: `"output" block must be placed in "outputs.tf" or "outputs.<area>.tf", not "main.tf"`,
					Range: hcl.Range{
						Filename: "main.tf",
						Start:    hcl.Pos{Line: 1, Column: 1},
						End:      hcl.Pos{Line: 1, Column: 7},
					},
				},
			},
		},
	}

	rule := NewTerraformOutputFileRule()
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
