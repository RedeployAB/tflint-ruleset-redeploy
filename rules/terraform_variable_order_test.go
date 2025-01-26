package rules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func TestTerraformVariableOrderRule(t *testing.T) {
	tests := []struct {
		Name    string
		Content string
		Issues  helper.Issues
	}{
		{
			Name: "OK - all required in alphabetical order, then optional in alphabetical order",
			Content: `
variable "alpha" {}
variable "beta" {}
variable "delta" {
  default = true
}
variable "gamma" {
  default = "some default"
}
`,
			Issues: helper.Issues{},
		},
		{
			Name: "NOT OK - optional before required",
			Content: `
variable "bar" {
  default = 123
}
variable "foo" {}
`,
			Issues: helper.Issues{
				{
					Rule:    NewTerraformVariableOrderRule(),
					Message: `Out-of-order variable "foo". Required variables must come first in alphabetical order, followed by optional variables in alphabetical order.`,
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 5, Column: 1},
						End:      hcl.Pos{Line: 5, Column: 20},
					},
				},
			},
		},
		{
			Name: "NOT OK - required out of alphabetical order",
			Content: `
variable "zzz" {}
variable "aaa" {}
`,
			Issues: helper.Issues{
				{
					Rule:    NewTerraformVariableOrderRule(),
					Message: `Out-of-order variable "aaa". Required variables must come first in alphabetical order, followed by optional variables in alphabetical order.`,
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 3, Column: 1},
						End:      hcl.Pos{Line: 3, Column: 20},
					},
				},
			},
		},
		{
			Name: "NOT OK - optional out of alphabetical order",
			Content: `
variable "opt_x" {
  default = 1
}
variable "opt_a" {
  default = 2
}
`,
			Issues: helper.Issues{
				{
					Rule:    NewTerraformVariableOrderRule(),
					Message: `Out-of-order variable "opt_a". Required variables must come first in alphabetical order, followed by optional variables in alphabetical order.`,
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 5, Column: 1},
						End:      hcl.Pos{Line: 5, Column: 22},
					},
				},
			},
		},
	}

	rule := NewTerraformVariableOrderRule()
	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runner := helper.TestRunner(t, map[string]string{
				"test.tf": tc.Content,
			})
			if err := rule.Check(runner); err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			helper.AssertIssues(t, tc.Issues, runner.Issues)
		})
	}
}
