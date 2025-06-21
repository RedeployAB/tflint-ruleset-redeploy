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
						Start:    hcl.Pos{Line: 5, Column: 2},
						End:      hcl.Pos{Line: 5, Column: 19},
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

func TestTerraformVariableSensitiveRule_Autofix(t *testing.T) {
	tests := []struct {
		Name     string
		Content  string
		Expected string
		HasFix   bool
	}{
		{
			Name: "Remove sensitive = false",
			Content: `variable "test" {
	description = "sensitive false"
	sensitive = false
}`,
			Expected: `variable "test" {
  description = "sensitive false"
}`,
			HasFix: true,
		},
		{
			Name: "Remove sensitive = false with extra spaces",
			Content: `variable "test" {
	description = "sensitive false"

	sensitive   =   false
}`,
			Expected: `variable "test" {
  description = "sensitive false"

}`,
			HasFix: true,
		},
		{
			Name: "Remove sensitive = false between other attributes",
			Content: `variable "test" {
	description = "sensitive false"
	sensitive = false
	type = string
}`,
			Expected: `variable "test" {
  description = "sensitive false"
  type        = string
}`,
			HasFix: true,
		},
		{
			Name: "Preserve sensitive = true",
			Content: `variable "test" {
	description = "sensitive true"
	sensitive = true
}`,
			Expected: `variable "test" {
	description = "sensitive true"
	sensitive = true
}`,
			HasFix: false,
		},
		{
			Name: "Multiple variables with one needing fix",
			Content: `variable "test1" {
	description = "sensitive false"
	sensitive = false
}

variable "test2" {
	description = "sensitive true"
	sensitive = true
}`,
			Expected: `variable "test1" {
  description = "sensitive false"
}

variable "test2" {
  description = "sensitive true"
  sensitive   = true
}`,
			HasFix: true,
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

			changes := runner.Changes()
			if tc.HasFix {
				helper.AssertChanges(t, map[string]string{
					"variables.tf": tc.Expected,
				}, changes)
			} else if len(changes) > 0 {
				t.Errorf("Expected no changes, but got: %v", changes)
			}
		})
	}
}
