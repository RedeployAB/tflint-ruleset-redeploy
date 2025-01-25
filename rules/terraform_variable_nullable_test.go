package rules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func TestTerraformVariableNullableRule(t *testing.T) {
	tests := []struct {
		Name    string
		Content string
		Issues  helper.Issues
	}{
		{
			Name: "OK - nullable = false, bool default = true",
			Content: `
variable "test" {
  type     = bool
  default  = true
  nullable = false
}
`,
			Issues: helper.Issues{},
		},
		{
			Name: "OK - no default, nullable = false",
			Content: `
variable "test" {
  type     = bool
  nullable = false
}
`,
			Issues: helper.Issues{},
		},
		{
			Name: "NOT OK - nullable set to true",
			Content: `
variable "test" {
  nullable = true
}
`,
			Issues: helper.Issues{
				{
					Rule:    NewTerraformVariableNullableRule(),
					Message: "nullable should not be set to true (the default is already true)",
					Range: hcl.Range{
						Filename: "variables.tf",
						Start:    hcl.Pos{Line: 3, Column: 3},
						End:      hcl.Pos{Line: 3, Column: 21},
					},
				},
			},
		},
		{
			Name: "NOT OK - boolean var with default = null",
			Content: `
variable "test" {
  type = bool

  default = null       
}
`,
			Issues: helper.Issues{
				{
					Rule:    NewTerraformVariableNullableRule(),
					Message: "boolean variables cannot have default = null",
					Range: hcl.Range{
						Filename: "variables.tf",
						Start:    hcl.Pos{Line: 4, Column: 3},
						End:      hcl.Pos{Line: 4, Column: 21},
					},
				},
			},
		},
		{
			Name: "NOT OK - default = null but has nullable declared",
			Content: `
variable "test" {
  default  = null
  nullable = false
}
`,
			Issues: helper.Issues{
				{
					Rule:    NewTerraformVariableNullableRule(),
					Message: "nullable must not be declared if default = null",
					Range: hcl.Range{
						Filename: "variables.tf",
						Start:    hcl.Pos{Line: 5, Column: 3},
						End:      hcl.Pos{Line: 5, Column: 22},
					},
				},
			},
		},
	}

	rule := NewTerraformVariableNullableRule()
	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runner := helper.TestRunner(t, map[string]string{"variables.tf": tc.Content})
			// Execute rule
			if err := rule.Check(runner); err != nil {
				t.Fatalf("Unexpected error: %s", err)
			}
			helper.AssertIssues(t, tc.Issues, runner.Issues)
		})
	}
}
