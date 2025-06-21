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
		Name     string
		Content  string
		Expected string
		HasFix   bool
	}{
		{
			Name: "Remove blank line after opening brace",
			Content: `resource "random_id" "example" {

  byte_length = 8
  keepers = {
    env = var.env_name
  }
}`,
			Expected: `resource "random_id" "example" {
  byte_length = 8
  keepers = {
    env = var.env_name
  }
}`,
			HasFix: true,
		},
		{
			Name: "Remove blank line before closing brace",
			Content: `resource "random_id" "example" {
  byte_length = 8
  keepers = {
    env = var.env_name
  }

}`,
			Expected: `resource "random_id" "example" {
  byte_length = 8
  keepers = {
    env = var.env_name
  }
}`,
			HasFix: true,
		},
		{
			Name: "Remove both leading and trailing blank lines in simple block",
			Content: `resource "test" "both" {

  name = "test"
  type = "example"

}`,
			Expected: `resource "test" "both" {
  name = "test"
  type = "example"

}`,
			HasFix: true,
		},
		{
			Name: "Preserve well-formatted blocks",
			Content: `resource "test" "example" {
  name = "test"
  tags = {
    Environment = "test"
  }
}`,
			Expected: `resource "test" "example" {
  name = "test"
  tags = {
    Environment = "test"
  }
}`,
			HasFix: false,
		},
		{
			Name:     "Preserve empty blocks",
			Content:  `resource "null_resource" "empty" {}`,
			Expected: `resource "null_resource" "empty" {}`,
			HasFix:   false,
		},
		{
			Name: "Preserve blocks with comments after opening brace",
			Content: `resource "test" "example" {
  # This is a comment
  name = "test"
}`,
			Expected: `resource "test" "example" {
  # This is a comment
  name = "test"
}`,
			HasFix: false,
		},
		{
			Name: "Remove blank line in nested resource blocks",
			Content: `resource "test" "example" {

  name = "test"

  lifecycle {
    prevent_destroy = true
  }
}`,
			Expected: `resource "test" "example" {
  name = "test"

  lifecycle {
    prevent_destroy = true
  }
}`,
			HasFix: true,
		},
		{
			Name: "Multiple blocks with issues",
			Content: `resource "test" "one" {

  name = "one"
}

module "two" {
  source = "./two"

}`,
			Expected: `resource "test" "one" {
  name = "one"
}

module "two" {
  source = "./two"
}`,
			HasFix: true,
		},
	}

	rule := NewTerraformNoLeadingTrailingBlankLinesRule()

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runner := helper.TestRunner(t, map[string]string{
				"main.tf": tc.Content,
			})

			if err := rule.Check(runner); err != nil {
				t.Fatalf("Unexpected error: %s", err)
			}

			changes := runner.Changes()
			if tc.HasFix {
				helper.AssertChanges(t, map[string]string{
					"main.tf": tc.Expected,
				}, changes)
			} else if len(changes) > 0 {
				t.Errorf("Expected no changes, but got: %v", changes)
			}
		})
	}
}
