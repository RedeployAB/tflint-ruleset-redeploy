package rules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func TestTerraformSingleTernaryPerLineRule(t *testing.T) {
	tests := []struct {
		Name     string
		Content  string
		Expected helper.Issues
	}{
		{
			Name:     "Single ternary per line is valid",
			Content:  readFixture(t, "single_ternary_per_line_valid.tf"),
			Expected: helper.Issues{},
		},
		{
			Name:    "Nested and chained ternaries on a single line",
			Content: readFixture(t, "single_ternary_per_line_invalid.tf"),
			Expected: helper.Issues{
				{
					Rule:    NewTerraformSingleTernaryPerLineRule(),
					Message: "Line contains 2 ternary operations; use local values to keep at most one ternary per line",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 3, Column: 23},
						End:      hcl.Pos{Line: 3, Column: 110},
					},
				},
				{
					Rule:    NewTerraformSingleTernaryPerLineRule(),
					Message: "Line contains 2 ternary operations; use local values to keep at most one ternary per line",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 6, Column: 13},
						End:      hcl.Pos{Line: 6, Column: 82},
					},
				},
			},
		},
		{
			Name:    "Chained and independent ternaries on a line",
			Content: readFixture(t, "single_ternary_per_line_edge.tf"),
			Expected: helper.Issues{
				{
					Rule:    NewTerraformSingleTernaryPerLineRule(),
					Message: "Line contains 3 ternary operations; use local values to keep at most one ternary per line",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 3, Column: 12},
						End:      hcl.Pos{Line: 3, Column: 49},
					},
				},
				{
					Rule:    NewTerraformSingleTernaryPerLineRule(),
					Message: "Line contains 2 ternary operations; use local values to keep at most one ternary per line",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 6, Column: 11},
						End:      hcl.Pos{Line: 6, Column: 24},
					},
				},
			},
		},
	}

	rule := NewTerraformSingleTernaryPerLineRule()

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runner := helper.TestRunner(t, map[string]string{
				"resource.tf": tc.Content,
			})
			if err := rule.Check(runner); err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			helper.AssertIssues(t, tc.Expected, runner.Issues)
		})
	}
}
