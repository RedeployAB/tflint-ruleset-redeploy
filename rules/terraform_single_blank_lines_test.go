package rules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"

	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func TestTerraformSingleBlankLinesRule(t *testing.T) {
	tests := []struct {
		Name   string
		File   string
		Issues helper.Issues
	}{
		{
			Name:   "OK - single blank line only",
			File:   "blank_line_ok_single.tf",
			Issues: helper.Issues{},
		},
		{
			Name:   "OK - no blank lines at all",
			File:   "blank_line_ok_none.tf",
			Issues: helper.Issues{},
		},
		{
			Name: "NOT OK - two consecutive blank lines",
			File: "blank_line_not_ok_multiple.tf",
			Issues: helper.Issues{
				{
					Rule:    NewTerraformSingleBlankLinesRule(),
					Message: "More than one consecutive blank line found at lines 3-4",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 3, Column: 1},
						End:      hcl.Pos{Line: 4, Column: 1},
					},
				},
			},
		},
	}

	rule := NewTerraformSingleBlankLinesRule()

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			// readFixture is found in rules/helper_test.go
			content := readFixture(t, tc.File)
			runner := helper.TestRunner(t, map[string]string{
				"resource.tf": content,
			})

			// Perform the check
			if err := rule.Check(runner); err != nil {
				t.Fatalf("Unexpected error occurred: %s", err)
			}

			// Compare actual issues vs. expected
			helper.AssertIssues(t, tc.Issues, runner.Issues)
		})
	}
}

func TestTerraformSingleBlankLinesRule_Autofix(t *testing.T) {
	tests := []struct {
		Name         string
		ContentFile  string
		ExpectedFile string
		HasFix       bool
	}{
		{
			Name:         "Fix two consecutive blank lines",
			ContentFile:  "single_blank_autofix_two_consecutive.tf",
			ExpectedFile: "single_blank_autofix_two_consecutive_expected.tf",
			HasFix:       true,
		},
		{
			Name:         "Fix three consecutive blank lines",
			ContentFile:  "single_blank_autofix_three_consecutive.tf",
			ExpectedFile: "single_blank_autofix_three_consecutive_expected.tf",
			HasFix:       true,
		},
		{
			Name:         "Fix multiple occurrences",
			ContentFile:  "single_blank_autofix_multiple_occurrences.tf",
			ExpectedFile: "single_blank_autofix_multiple_occurrences_expected.tf",
			HasFix:       true,
		},
		{
			Name:         "Preserve single blank lines",
			ContentFile:  "single_blank_autofix_preserve_single.tf",
			ExpectedFile: "single_blank_autofix_preserve_single.tf",
			HasFix:       false,
		},
		{
			Name:         "Fix blank lines at end of file",
			ContentFile:  "single_blank_autofix_end_of_file.tf",
			ExpectedFile: "single_blank_autofix_end_of_file_expected.tf",
			HasFix:       true,
		},
		{
			Name:         "Fix blank lines at start",
			ContentFile:  "single_blank_autofix_start.tf",
			ExpectedFile: "single_blank_autofix_start_expected.tf",
			HasFix:       true,
		},
	}

	rule := NewTerraformSingleBlankLinesRule()

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			content := readFixture(t, tc.ContentFile)
			runner := helper.TestRunner(t, map[string]string{
				"main.tf": content,
			})

			if err := rule.Check(runner); err != nil {
				t.Fatalf("Unexpected error: %s", err)
			}

			changes := runner.Changes()
			if tc.HasFix {
				expected := readFixture(t, tc.ExpectedFile)
				helper.AssertChanges(t, map[string]string{
					"main.tf": expected,
				}, changes)
			} else if len(changes) > 0 {
				t.Errorf("Expected no changes, but got: %v", changes)
			}
		})
	}
}
