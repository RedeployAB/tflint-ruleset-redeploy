package rules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func TestTerraformBlockFormat(t *testing.T) {
	tests := []struct {
		Name    string
		Content string
		Issues  helper.Issues
	}{
		{
			Name:    "OK - attribute then block with blank line",
			Content: readFixture(t, "block_fmt_ok_attr_then_block_blank_line.tf"),
			Issues:  helper.Issues{},
		},
		{
			Name:    "NOT OK - attribute then block with no blank line",
			Content: readFixture(t, "block_fmt_not_ok_attr_then_block_no_blank_line.tf"),
			Issues: helper.Issues{
				{
					Rule:    NewTerraformBlockFormatRule(),
					Message: "Expected exactly one blank line before this block",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 3, Column: 3},
						End:      hcl.Pos{Line: 3, Column: 10},
					},
				},
			},
		},
		{
			Name:    "OK - single block first (no attribute), no blank line after brace",
			Content: readFixture(t, "block_fmt_ok_single_block_first.tf"),
			Issues:  helper.Issues{},
		},
		{
			Name:    "NOT OK - single block first with extra blank line after brace",
			Content: readFixture(t, "block_fmt_not_ok_single_block_first_extra_blank.tf"),
			Issues: helper.Issues{
				{
					Rule:    NewTerraformBlockFormatRule(),
					Message: "Block should appear immediately after opening brace when it's the first item (no blank lines)",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 3, Column: 3},
						End:      hcl.Pos{Line: 3, Column: 10},
					},
				},
			},
		},
		{
			Name:    "OK - multiple blocks each separated by single blank line",
			Content: readFixture(t, "block_fmt_ok_multiple_blocks_single_blank_line.tf"),
			Issues:  helper.Issues{},
		},
		{
			Name:    "NOT OK - multiple blocks no blank line between them",
			Content: readFixture(t, "block_fmt_not_ok_multiple_blocks_no_blank_line.tf"),
			Issues: helper.Issues{
				{
					Rule:    NewTerraformBlockFormatRule(),
					Message: "Expected exactly one blank line before this block",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 6, Column: 3},
						End:      hcl.Pos{Line: 6, Column: 26},
					},
				},
				{
					Rule:    NewTerraformBlockFormatRule(),
					Message: "Expected exactly one blank line before this block",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 10, Column: 3},
						End:      hcl.Pos{Line: 10, Column: 22},
					},
				},
			},
		},
		{
			Name:    "OK - attributes then multiple blocks each separated by single blank line",
			Content: readFixture(t, "block_fmt_ok_attr_multiple_blocks_single_blank_line.tf"),
			Issues:  helper.Issues{},
		},
		{
			Name:    "NOT OK - attribute then block with no blank line, then next block also with no blank line",
			Content: readFixture(t, "block_fmt_not_ok_attr_no_blank_line_multiple_blocks.tf"),
			Issues: helper.Issues{
				{
					Rule:    NewTerraformBlockFormatRule(),
					Message: "Expected exactly one blank line before this block",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 3, Column: 3},
						End:      hcl.Pos{Line: 3, Column: 30},
					},
				},
				{
					Rule:    NewTerraformBlockFormatRule(),
					Message: "Expected exactly one blank line before this block",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 7, Column: 3},
						End:      hcl.Pos{Line: 7, Column: 26},
					},
				},
			},
		},
		{
			Name:    "OK - single data block with first sub-block no blank line",
			Content: readFixture(t, "block_fmt_ok_data_single_block_no_blank_line.tf"),
			Issues:  helper.Issues{},
		},
		{
			Name:    "NOT OK - single data block with extra blank line after brace",
			Content: readFixture(t, "block_fmt_not_ok_data_extra_blank_line_after_brace.tf"),
			Issues: helper.Issues{
				{
					Rule:    NewTerraformBlockFormatRule(),
					Message: "Block should appear immediately after opening brace when it's the first item (no blank lines)",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 3, Column: 3},
						End:      hcl.Pos{Line: 3, Column: 9},
					},
				},
			},
		},
		{
			Name:    "OK - terraform block with two sub-blocks, each separated by one blank line",
			Content: readFixture(t, "block_fmt_ok_terraform_two_subblocks.tf"),
			Issues:  helper.Issues{},
		},
		{
			Name:    "NOT OK - provider block with no blank line between sub-blocks",
			Content: readFixture(t, "block_fmt_not_ok_provider_no_blank_line_between_subblocks.tf"),
			Issues: helper.Issues{
				{
					Rule:    NewTerraformBlockFormatRule(),
					Message: "Expected exactly one blank line before this block",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 3, Column: 3},
						End:      hcl.Pos{Line: 3, Column: 14},
					},
				},
				{
					Rule:    NewTerraformBlockFormatRule(),
					Message: "Expected exactly one blank line before this block",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 4, Column: 3},
						End:      hcl.Pos{Line: 4, Column: 15},
					},
				},
			},
		},
		{
			Name:    "OK - block with preceding comment before sub-block",
			Content: readFixture(t, "block_fmt_ok_comment_before_block.tf"),
			Issues:  helper.Issues{},
		},
	}

	rule := NewTerraformBlockFormatRule()

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runner := helper.TestRunner(t, map[string]string{
				"resource.tf": tc.Content,
			})

			err := rule.Check(runner)
			if err != nil {
				t.Fatalf("Unexpected error occurred: %s", err)
			}

			helper.AssertIssues(t, tc.Issues, runner.Issues)
		})
	}

	t.Run("variable block tests", func(t *testing.T) {
		tests := []struct {
			Name    string
			Content string
			Issues  helper.Issues
		}{
			{
				Name: "OK - variable with single validation block",
				Content: `
variable "example" {
	type = string

	validation {
		// ..
	}
}
`,
				Issues: helper.Issues{},
			},
			{
				Name: "OK - variable with multiple validation blocks",
				Content: `
variable "example" {
	type = string

	validation {
		// ..
	}

	validation {
		// ..
	}
}
`,
				Issues: helper.Issues{},
			},
			{
				Name: "NOT OK - variable with validation block no blank line",
				Content: `
variable "example" {
	type = string
	validation {
		// ..
	}
}
`,
				Issues: helper.Issues{
					{
						Rule:    NewTerraformBlockFormatRule(),
						Message: "Expected exactly one blank line before this block",
						Range: hcl.Range{
							Filename: "resource.tf",
							Start:    hcl.Pos{Line: 4, Column: 2},
							End:      hcl.Pos{Line: 4, Column: 12},
						},
					},
				},
			},
			{
				Name: "NOT OK - variable with multiple validation blocks no blank line",
				Content: `
variable "example" {
	type = string

	validation {
		// ..
	}
	validation {
		// ..
	}
}
`,
				Issues: helper.Issues{
					{
						Rule:    NewTerraformBlockFormatRule(),
						Message: "Expected exactly one blank line before this block",
						Range: hcl.Range{
							Filename: "resource.tf",
							Start:    hcl.Pos{Line: 8, Column: 2},
							End:      hcl.Pos{Line: 8, Column: 12},
						},
					},
				},
			},
		}

		for _, tc := range tests {
			tc := tc // capture range variable
			t.Run(tc.Name, func(t *testing.T) {
				runner := helper.TestRunner(t, map[string]string{
					"resource.tf": tc.Content,
				})
				err := rule.Check(runner)
				if err != nil {
					t.Fatalf("Unexpected error occurred: %s", err)
				}
				helper.AssertIssues(t, tc.Issues, runner.Issues)
			})
		}
	})

	t.Run("output block tests", func(t *testing.T) {
		tests := []struct {
			Name    string
			Content string
			Issues  helper.Issues
		}{
			{
				Name: "OK - blank line before block",
				Content: `
output "example" {
	value = "something"

	precondition {
		// ..
	}
}
`,
				Issues: helper.Issues{},
			},
			{
				Name: "NOT OK - no blank line before block",
				Content: `
output "example" {
	value = "something"
	precondition {
		// ..
	}
}
`,
				Issues: helper.Issues{
					{
						Rule:    NewTerraformBlockFormatRule(),
						Message: "Expected exactly one blank line before this block",
						Range: hcl.Range{
							Filename: "resource.tf",
							Start:    hcl.Pos{Line: 4, Column: 2},
							End:      hcl.Pos{Line: 4, Column: 14},
						},
					},
				},
			},
		}

		for _, tc := range tests {
			tc := tc // capture range variable
			t.Run(tc.Name, func(t *testing.T) {
				runner := helper.TestRunner(t, map[string]string{
					"resource.tf": tc.Content,
				})
				err := rule.Check(runner)
				if err != nil {
					t.Fatalf("Unexpected error occurred: %s", err)
				}
				helper.AssertIssues(t, tc.Issues, runner.Issues)
			})
		}
	})
}
