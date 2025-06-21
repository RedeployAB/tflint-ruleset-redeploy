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
		Name     string
		Content  string
		Expected string
		HasFix   bool
	}{
		{
			Name: "Fix two consecutive blank lines",
			Content: `resource "random_uuid" "test" {
  for_each = local.test


  lifecycle {
    prevent_destroy = true
  }
}`,
			Expected: `resource "random_uuid" "test" {
  for_each = local.test

  lifecycle {
    prevent_destroy = true
  }
}`,
			HasFix: true,
		},
		{
			Name: "Fix three consecutive blank lines",
			Content: `variable "test" {
  type = string



  default = "value"
}`,
			Expected: `variable "test" {
  type = string

  default = "value"
}`,
			HasFix: true,
		},
		{
			Name: "Fix multiple occurrences",
			Content: `locals {
  a = 1


  b = 2



  c = 3
}`,
			Expected: `locals {
  a = 1

  b = 2

  c = 3
}`,
			HasFix: true,
		},
		{
			Name: "Preserve single blank lines",
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
			Name: "Fix blank lines at end of file",
			Content: `output "test" {
  value = "test"
}


`,
			Expected: `output "test" {
  value = "test"
}

`,
			HasFix: true,
		},
		{
			Name: "Fix blank lines at start",
			Content: `

module "test" {
  source = "./test"
}`,
			Expected: `
module "test" {
  source = "./test"
}`,
			HasFix: true,
		},
	}

	rule := NewTerraformSingleBlankLinesRule()

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
