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
		Name         string
		ContentFile  string
		ExpectedFile string
		HasFix       bool
	}{
		{
			Name:         "Remove sensitive = false",
			ContentFile:  "output_sensitive_autofix_remove_false.tf",
			ExpectedFile: "output_sensitive_autofix_remove_false_expected.tf",
			HasFix:       true,
		},
		{
			Name:         "Remove sensitive = false with extra spaces",
			ContentFile:  "output_sensitive_autofix_extra_spaces.tf",
			ExpectedFile: "output_sensitive_autofix_extra_spaces_expected.tf",
			HasFix:       true,
		},
		{
			Name:         "Remove sensitive = false between other attributes",
			ContentFile:  "output_sensitive_autofix_between_attrs.tf",
			ExpectedFile: "output_sensitive_autofix_between_attrs_expected.tf",
			HasFix:       true,
		},
		{
			Name:         "Preserve sensitive = true",
			ContentFile:  "output_sensitive_autofix_preserve_true.tf",
			ExpectedFile: "output_sensitive_autofix_preserve_true.tf",
			HasFix:       false,
		},
		{
			Name:         "Multiple outputs with one needing fix",
			ContentFile:  "output_sensitive_autofix_multiple.tf",
			ExpectedFile: "output_sensitive_autofix_multiple_expected.tf",
			HasFix:       true,
		},
		{
			Name:         "Output with precondition",
			ContentFile:  "output_sensitive_autofix_with_precondition.tf",
			ExpectedFile: "output_sensitive_autofix_with_precondition_expected.tf",
			HasFix:       true,
		},
	}

	rule := NewTerraformOutputSensitiveRule()

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			content := readFixture(t, tc.ContentFile)
			runner := helper.TestRunner(t, map[string]string{
				"outputs.tf": content,
			})

			if err := rule.Check(runner); err != nil {
				t.Fatalf("Unexpected error: %s", err)
			}

			changes := runner.Changes()
			if tc.HasFix {
				expected := readFixture(t, tc.ExpectedFile)
				helper.AssertChanges(t, map[string]string{
					"outputs.tf": expected,
				}, changes)
			} else if len(changes) > 0 {
				t.Errorf("Expected no changes, but got: %v", changes)
			}
		})
	}
}
