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
					Message: "No blank or comment line allowed immediately after '{'",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 1},
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
					Message: "No blank or comment line allowed immediately before '}'",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 6, Column: 1},
						End:      hcl.Pos{Line: 6, Column: 1},
					},
				},
			},
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
