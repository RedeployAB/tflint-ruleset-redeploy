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
						Start:    hcl.Pos{Line: 3, Column: 2},
						End:      hcl.Pos{Line: 3, Column: 17},
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
						Start:    hcl.Pos{Line: 5, Column: 2},
						End:      hcl.Pos{Line: 5, Column: 16},
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
						Start:    hcl.Pos{Line: 4, Column: 2},
						End:      hcl.Pos{Line: 4, Column: 18},
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

func TestTerraformVariableNullableRuleAutofix(t *testing.T) {
	tests := []struct {
		Name     string
		Content  string
		Expected string
	}{
		{
			Name: "Fix - remove nullable = true",
			Content: `
variable "test" {
	description = "test variable"
	nullable    = true
}
`,
			Expected: `
variable "test" {
  description = "test variable"
}
`,
		},
		{
			Name: "Fix - remove nullable when default = null",
			Content: `
variable "test" {
	description = "test variable"
	default     = null
	nullable    = false
}
`,
			Expected: `
variable "test" {
  description = "test variable"
  default     = null
}
`,
		},
		{
			Name: "Fix - remove nullable = true with other attributes",
			Content: `
variable "test" {
	description = "test variable"
	type        = string
	nullable    = true
	validation {
		condition     = length(var.test) > 0
		error_message = "Must not be empty"
	}
}
`,
			Expected: `
variable "test" {
  description = "test variable"
  type        = string
  validation {
    condition     = length(var.test) > 0
    error_message = "Must not be empty"
  }
}
`,
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

			// Check that we have issues
			if len(runner.Issues) == 0 {
				t.Fatalf("Expected issues but got none")
			}

			// Check the autofix
			changes := runner.Changes()
			expected := map[string]string{"variables.tf": tc.Expected}
			helper.AssertChanges(t, expected, changes)
		})
	}
}
