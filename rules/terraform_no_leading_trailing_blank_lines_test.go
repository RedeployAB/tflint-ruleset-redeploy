package rules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func TestTerraformNoLeadingTrailingBlankLinesRule(t *testing.T) {
	tests := []struct {
		Name   string
		File   string
		Issues helper.Issues
	}{
		{
			Name:   "OK - resource with first arg immediately after '{' and last arg immediately before '}'",
			File:   "leading_trailing_blank_ok.tf",
			Issues: helper.Issues{},
		},
		{
			Name: "NOT OK - blank line after opening brace",
			File: "leading_trailing_blank_not_ok_after_brace.tf",
			Issues: helper.Issues{
				{
					Rule:    NewTerraformNoLeadingTrailingBlankLinesRule(),
					Message: "No blank line allowed immediately after '{'",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 3, Column: 1},
					},
				},
			},
		},
		{
			Name: "NOT OK - blank line before closing brace",
			File: "leading_trailing_blank_not_ok_before_brace.tf",
			Issues: helper.Issues{
				{
					Rule:    NewTerraformNoLeadingTrailingBlankLinesRule(),
					Message: "No blank line allowed immediately before '}'",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 6, Column: 1},
						End:      hcl.Pos{Line: 7, Column: 1},
					},
				},
			},
		},
		{
			Name:   "OK - comment line after opening brace",
			File:   "leading_trailing_comment_ok_after_brace.tf",
			Issues: helper.Issues{},
		},
		{
			Name:   "OK - empty block with braces on the same line",
			File:   "leading_trailing_blank_ok_empty_block.tf",
			Issues: helper.Issues{},
		},
	}

	rule := NewTerraformNoLeadingTrailingBlankLinesRule()

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			// readFixture is found in rules/helper_test.go
			content := readFixture(t, tc.File)
			runner := helper.TestRunner(t, map[string]string{
				"resource.tf": content,
			})

			if err := rule.Check(runner); err != nil {
				t.Fatalf("Unexpected error occurred: %s", err)
			}

			helper.AssertIssues(t, tc.Issues, runner.Issues)
		})
	}
}

func TestTerraformNoLeadingTrailingBlankLinesRule_Autofix(t *testing.T) {
	tests := []struct {
		Name         string
		ContentFile  string
		ExpectedFile string
		HasFix       bool
	}{
		{
			Name:         "Remove blank line after opening brace",
			ContentFile:  "no_leading_trailing_autofix_blank_after_open.tf",
			ExpectedFile: "no_leading_trailing_autofix_blank_after_open_expected.tf",
			HasFix:       true,
		},
		{
			Name:         "Remove blank line before closing brace",
			ContentFile:  "no_leading_trailing_autofix_blank_before_close.tf",
			ExpectedFile: "no_leading_trailing_autofix_blank_before_close_expected.tf",
			HasFix:       true,
		},
		{
			Name:         "Remove both leading and trailing blank lines in simple block",
			ContentFile:  "no_leading_trailing_autofix_both.tf",
			ExpectedFile: "no_leading_trailing_autofix_both_expected.tf",
			HasFix:       true,
		},
		{
			Name:         "Preserve well-formatted blocks",
			ContentFile:  "no_leading_trailing_autofix_well_formatted.tf",
			ExpectedFile: "no_leading_trailing_autofix_well_formatted.tf",
			HasFix:       false,
		},
		{
			Name:         "Preserve empty blocks",
			ContentFile:  "no_leading_trailing_autofix_empty_block.tf",
			ExpectedFile: "no_leading_trailing_autofix_empty_block.tf",
			HasFix:       false,
		},
		{
			Name:         "Preserve blocks with comments after opening brace",
			ContentFile:  "no_leading_trailing_autofix_comment_after_brace.tf",
			ExpectedFile: "no_leading_trailing_autofix_comment_after_brace.tf",
			HasFix:       false,
		},
		{
			Name:         "Remove blank line in nested resource blocks",
			ContentFile:  "no_leading_trailing_autofix_nested.tf",
			ExpectedFile: "no_leading_trailing_autofix_nested_expected.tf",
			HasFix:       true,
		},
		{
			Name:         "Multiple blocks with issues",
			ContentFile:  "no_leading_trailing_autofix_multiple.tf",
			ExpectedFile: "no_leading_trailing_autofix_multiple_expected.tf",
			HasFix:       true,
		},
	}

	rule := NewTerraformNoLeadingTrailingBlankLinesRule()

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
