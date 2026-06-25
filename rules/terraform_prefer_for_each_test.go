package rules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func TestTerraformPreferForEachRule(t *testing.T) {
	tests := []struct {
		Name     string
		Content  string
		Expected helper.Issues
	}{
		{
			Name:     "Toggles, references and for_each are valid",
			Content:  readFixture(t, "prefer_for_each_valid.tf"),
			Expected: helper.Issues{},
		},
		{
			Name:    "count creating multiple instances",
			Content: readFixture(t, "prefer_for_each_invalid.tf"),
			Expected: helper.Issues{
				{
					Rule:    NewTerraformPreferForEachRule(),
					Message: "Use 'for_each' instead of 'count' to create multiple instances",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 2, Column: 11},
						End:      hcl.Pos{Line: 2, Column: 12},
					},
				},
				{
					Rule:    NewTerraformPreferForEachRule(),
					Message: "Use 'for_each' instead of 'count' to create multiple instances",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 6, Column: 11},
						End:      hcl.Pos{Line: 6, Column: 30},
					},
				},
				{
					Rule:    NewTerraformPreferForEachRule(),
					Message: "Use 'for_each' instead of 'count' to create multiple instances",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 10, Column: 11},
						End:      hcl.Pos{Line: 10, Column: 47},
					},
				},
			},
		},
	}

	rule := NewTerraformPreferForEachRule()

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
