package rules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func TestTerraformVariableEphemeralRule(t *testing.T) {
	tests := []struct {
		Name    string
		Content string
		Issues  helper.Issues
	}{
		{
			Name: "OK - ephemeral not set",
			Content: `
variable "test" {
  description = "some test"
}
`,
			Issues: helper.Issues{},
		},
		{
			Name: "OK - ephemeral = true",
			Content: `
variable "test" {
  description = "another test"
  ephemeral   = true
}
`,
			Issues: helper.Issues{},
		},
		{
			Name: "NOT OK - ephemeral = false",
			Content: `
variable "test" {
  description = "another test"

  ephemeral = false     
}
`,
			Issues: helper.Issues{
				{
					Rule:    NewTerraformVariableEphemeralRule(),
					Message: "ephemeral should not be set to false (omit instead)",
					Range: hcl.Range{
						Filename: "variables.tf",
						Start:    hcl.Pos{Line: 4, Column: 3},
						End:      hcl.Pos{Line: 4, Column: 24},
					},
				},
			},
		},
	}

	rule := NewTerraformVariableEphemeralRule()

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
