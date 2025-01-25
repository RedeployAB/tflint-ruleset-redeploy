package rules

import (
	hcl "github.com/hashicorp/hcl/v2"
	"path/filepath"
	"testing"

	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func TestTerraformSingleBlankLinesRule(t *testing.T) {
	tests := []struct {
		Name    string
		File    string
		Issues  helper.Issues
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
					Message: "More than one consecutive blank line found at line 6",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 6, Column: 1},
						End:      hcl.Pos{Line: 6, Column: 1},
					},
				},
			},
		},
	}

	rule := NewTerraformSingleBlankLinesRule()

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			// readFixture is found in rules/helper_test.go
			content := readFixture(t, filepath.Join("testdata", tc.File))
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
