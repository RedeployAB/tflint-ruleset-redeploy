package rules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func TestTerraformSourceFormat(t *testing.T) {
	tests := []struct {
		Name    string
		Content string
		Issues  helper.Issues
	}{
		{
			Name:    "OK - only source, block ends",
			Content: readFixture(t, "source_format_only_source.tf"),
			Issues:  helper.Issues{},
		},
		{
			Name:    "OK - source plus version, block ends",
			Content: readFixture(t, "source_format_source_version.tf"),
			Issues:  helper.Issues{},
		},
		{
			Name:    "NOT OK - source plus version, extra blank line at end is disallowed",
			Content: readFixture(t, "source_format_extra_blank_line.tf"),
			Issues: helper.Issues{
				{
					Rule:    NewTerraformSourceFormatRule(),
					Message: "Unexpected blank line after 'version' when block ends",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 4, Column: 1},
						End:      hcl.Pos{Line: 4, Column: 1},
					},
				},
			},
		},
		{
			Name:    "NOT OK - source alone with trailing blank line before closing brace",
			Content: readFixture(t, "source_format_source_trailing_blank.tf"),
			Issues: helper.Issues{
				{
					Rule:    NewTerraformSourceFormatRule(),
					Message: "Unexpected blank line after 'source' when block ends",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 3, Column: 1},
						End:      hcl.Pos{Line: 3, Column: 1},
					},
				},
			},
		},
		{
			Name:    "OK - source plus version, more property after blank line",
			Content: readFixture(t, "source_format_source_version_with_property.tf"),
			Issues:  helper.Issues{},
		},
		{
			Name:    "OK - source alone, more property after blank line",
			Content: readFixture(t, "source_format_source_with_property.tf"),
			Issues:  helper.Issues{},
		},
		{
			Name:    "NOT OK - source plus version, property follows with no blank line",
			Content: readFixture(t, "source_format_no_blank_before_property.tf"),
			Issues: helper.Issues{
				{
					Rule:    NewTerraformSourceFormatRule(),
					Message: "Expected a blank line after 'version'",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 4, Column: 1},
						End:      hcl.Pos{Line: 4, Column: 1},
					},
				},
			},
		},
	}

	rule := NewTerraformSourceFormatRule()

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runner := helper.TestRunner(t, map[string]string{
				"resource.tf": tc.Content,
			})

			if err := rule.Check(runner); err != nil {
				t.Fatalf("Unexpected error occurred: %s", err)
			}

			helper.AssertIssues(t, tc.Issues, runner.Issues)
		})
	}
}

func TestTerraformSourceFormat_Autofix(t *testing.T) {
	tests := []struct {
		Name         string
		ContentFile  string
		ExpectedFile string
		HasFix       bool
	}{
		{
			Name:         "Add blank line after source when property follows",
			ContentFile:  "source_format_autofix_add_blank_after_source.tf",
			ExpectedFile: "source_format_autofix_add_blank_after_source_expected.tf",
			HasFix:       true,
		},
		{
			Name:         "Add blank line after version when property follows",
			ContentFile:  "source_format_autofix_add_blank_after_version.tf",
			ExpectedFile: "source_format_autofix_add_blank_after_version_expected.tf",
			HasFix:       true,
		},
		{
			Name:         "Remove blank line after source when block ends",
			ContentFile:  "source_format_autofix_remove_blank_after_source.tf",
			ExpectedFile: "source_format_autofix_remove_blank_after_source_expected.tf",
			HasFix:       true,
		},
		{
			Name:         "Remove blank line after version when block ends",
			ContentFile:  "source_format_autofix_remove_blank_after_version.tf",
			ExpectedFile: "source_format_autofix_remove_blank_after_version_expected.tf",
			HasFix:       true,
		},
		{
			Name:         "Preserve correct format with blank line",
			ContentFile:  "source_format_autofix_preserve_correct.tf",
			ExpectedFile: "source_format_autofix_preserve_correct.tf",
			HasFix:       false,
		},
		{
			Name:         "Preserve correct format without blank line when block ends",
			ContentFile:  "source_format_autofix_preserve_without_blank.tf",
			ExpectedFile: "source_format_autofix_preserve_without_blank.tf",
			HasFix:       false,
		},
		{
			Name:         "Handle comments after source/version",
			ContentFile:  "source_format_autofix_handle_comments.tf",
			ExpectedFile: "source_format_autofix_handle_comments.tf",
			HasFix:       false,
		},
		{
			Name:         "Multiple blank lines after source when block ends",
			ContentFile:  "source_format_autofix_multiple_blanks.tf",
			ExpectedFile: "source_format_autofix_multiple_blanks_expected.tf",
			HasFix:       true,
		},
	}

	rule := NewTerraformSourceFormatRule()

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
