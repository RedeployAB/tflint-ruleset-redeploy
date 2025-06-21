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
		Name         string
		ContentFile  string
		ExpectedFile string
		HasFix       bool
	}{
		{
			Name:         "Remove sensitive = false",
			ContentFile:  "variable_sensitive_autofix_remove_false.tf",
			ExpectedFile: "variable_sensitive_autofix_remove_false_expected.tf",
			HasFix:       true,
		},
		{
			Name:         "Remove sensitive = false with extra spaces",
			ContentFile:  "variable_sensitive_autofix_extra_spaces.tf",
			ExpectedFile: "variable_sensitive_autofix_extra_spaces_expected.tf",
			HasFix:       true,
		},
		{
			Name:         "Remove sensitive = false between other attributes",
			ContentFile:  "variable_sensitive_autofix_between_attrs.tf",
			ExpectedFile: "variable_sensitive_autofix_between_attrs_expected.tf",
			HasFix:       true,
		},
		{
			Name:         "Preserve sensitive = true",
			ContentFile:  "variable_sensitive_autofix_preserve_true.tf",
			ExpectedFile: "variable_sensitive_autofix_preserve_true.tf",
			HasFix:       false,
		},
		{
			Name:         "Multiple variables with one needing fix",
			ContentFile:  "variable_sensitive_autofix_multiple.tf",
			ExpectedFile: "variable_sensitive_autofix_multiple_expected.tf",
			HasFix:       true,
		},
	}

	rule := NewTerraformVariableSensitiveRule()

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			content := readFixture(t, tc.ContentFile)
			runner := helper.TestRunner(t, map[string]string{
				"variables.tf": content,
			})

			if err := rule.Check(runner); err != nil {
				t.Fatalf("Unexpected error: %s", err)
			}

			changes := runner.Changes()
			if tc.HasFix {
				expected := readFixture(t, tc.ExpectedFile)
				helper.AssertChanges(t, map[string]string{
					"variables.tf": expected,
				}, changes)
			} else if len(changes) > 0 {
				t.Errorf("Expected no changes, but got: %v", changes)
			}
		})
	}
}
