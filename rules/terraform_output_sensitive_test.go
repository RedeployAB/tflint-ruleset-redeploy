package rules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func TestTerraformOutputSensitiveRule(t *testing.T) {
	tests := []struct {
		Name   string
		File   string
		Issues helper.Issues
	}{
		{
			Name:   "OK - no sensitive declared",
			File:   "output_sensitive_ok_none.tf",
			Issues: helper.Issues{},
		},
		{
			Name:   "OK - sensitive = true",
			File:   "output_sensitive_ok_true.tf",
			Issues: helper.Issues{},
		},
		{
			Name: "NOT OK - sensitive = false",
			File: "output_sensitive_not_ok_false.tf",
			Issues: helper.Issues{
				{
					Rule:    NewTerraformOutputSensitiveRule(),
					Message: "sensitive should not be set to false (omit instead)",
					Range: hcl.Range{
						Filename: "outputs.tf",
						Start:    hcl.Pos{Line: 4, Column: 3},
						End:      hcl.Pos{Line: 4, Column: 22},
					},
				},
			},
		},
	}

	rule := NewTerraformOutputSensitiveRule()

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			content := readFixture(t, tc.File)
			runner := helper.TestRunner(t, map[string]string{
				"outputs.tf": content,
			})

			if err := rule.Check(runner); err != nil {
				t.Fatalf("Unexpected error: %s", err)
			}

			helper.AssertIssues(t, tc.Issues, runner.Issues)
		})
	}
}

func TestTerraformOutputSensitiveRule_Autofix(t *testing.T) {
	tests := []struct {
		Name     string
		Content  string
		Expected string
		HasFix   bool
	}{
		{
			Name: "Remove sensitive = false",
			Content: `output "username" {
  description = "An output incorrectly marked sensitive = false."
  value       = "username"
  sensitive   = false
}`,
			Expected: `output "username" {
  description = "An output incorrectly marked sensitive = false."
  value       = "username"
}`,
			HasFix: true,
		},
		{
			Name: "Remove sensitive = false with extra spaces",
			Content: `output "test" {
  value     = "test value"

  sensitive = false
}`,
			Expected: `output "test" {
  value = "test value"

}`,
			HasFix: true,
		},
		{
			Name: "Remove sensitive = false between other attributes",
			Content: `output "test" {
  description = "test output"
  sensitive   = false
  value       = "test value"
}`,
			Expected: `output "test" {
  description = "test output"
  value       = "test value"
}`,
			HasFix: true,
		},
		{
			Name: "Preserve sensitive = true",
			Content: `output "secret" {
  description = "A secret output"
  value       = var.secret_value
  sensitive   = true
}`,
			Expected: `output "secret" {
  description = "A secret output"
  value       = var.secret_value
  sensitive   = true
}`,
			HasFix: false,
		},
		{
			Name: "Multiple outputs with one needing fix",
			Content: `output "public" {
  value     = "public value"
  sensitive = false
}

output "secret" {
  value     = "secret value"
  sensitive = true
}`,
			Expected: `output "public" {
  value = "public value"
}

output "secret" {
  value     = "secret value"
  sensitive = true
}`,
			HasFix: true,
		},
		{
			Name: "Output with precondition",
			Content: `output "test" {
  value     = "test"
  sensitive = false

  precondition {
    condition     = length(var.name) > 0
    error_message = "Name must not be empty"
  }
}`,
			Expected: `output "test" {
  value = "test"

  precondition {
    condition     = length(var.name) > 0
    error_message = "Name must not be empty"
  }
}`,
			HasFix: true,
		},
	}

	rule := NewTerraformOutputSensitiveRule()

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runner := helper.TestRunner(t, map[string]string{
				"outputs.tf": tc.Content,
			})

			if err := rule.Check(runner); err != nil {
				t.Fatalf("Unexpected error: %s", err)
			}

			changes := runner.Changes()
			if tc.HasFix {
				helper.AssertChanges(t, map[string]string{
					"outputs.tf": tc.Expected,
				}, changes)
			} else if len(changes) > 0 {
				t.Errorf("Expected no changes, but got: %v", changes)
			}
		})
	}
}
