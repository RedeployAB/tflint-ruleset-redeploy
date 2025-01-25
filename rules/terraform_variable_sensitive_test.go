package rules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func TestTerraformVariableSensitiveRule(t *testing.T) {
	tests := []struct {
		Name    string
		Content string
		Issues  helper.Issues
	}{
		{
			Name: "OK - no sensitive declared",
			Content: `
variable "test" {
  description = "no sensitive declared"
}
`,
			Issues: helper.Issues{},
		},
		{
			Name: "OK - sensitive = true",
			Content: `
variable "test" {
  description = "sensitive true"
  sensitive   = true
}
`,
			Issues: helper.Issues{},
		},
		{
			Name: "NOT OK - sensitive = false",
			Content: `
variable "test" {
  description = "sensitive false"

  sensitive = false     
}
`,
			Issues: helper.Issues{
				{
					Rule:    NewTerraformVariableSensitiveRule(),
					Message: "sensitive should not be set to false (omit instead)",
					Range: hcl.Range{
						Filename: "variables.tf",
						Start:    hcl.Pos{Line: 4, Column: 3},
						End:      hcl.Pos{Line: 4, Column: 22},
					},
				},
			},
		},
	}

	rule := NewTerraformVariableSensitiveRule()

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runner := helper.TestRunner(t, map[string]string{
				"variables.tf": tc.Content,
			})

			if err := rule.Check(runner); err != nil {
				t.Fatalf("Unexpected error: %s", err)
			}

			helper.AssertIssues(t, tc.Issues, runner.Issues)
		})
	}
}
