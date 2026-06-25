package rules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func TestTerraformLifecycleRedundantDefaultRule(t *testing.T) {
	tests := []struct {
		Name     string
		Content  string
		Expected helper.Issues
	}{
		{
			Name:     "Non-default values and non-literals are valid",
			Content:  readFixture(t, "lifecycle_redundant_default_valid.tf"),
			Expected: helper.Issues{},
		},
		{
			Name:    "Redundant false defaults",
			Content: readFixture(t, "lifecycle_redundant_default_invalid.tf"),
			Expected: helper.Issues{
				{
					Rule:    NewTerraformLifecycleRedundantDefaultRule(),
					Message: "prevent_destroy should not be set to false (omit instead)",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 3, Column: 5},
						End:      hcl.Pos{Line: 3, Column: 34},
					},
				},
				{
					Rule:    NewTerraformLifecycleRedundantDefaultRule(),
					Message: "create_before_destroy should not be set to false (omit instead)",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 4, Column: 5},
						End:      hcl.Pos{Line: 4, Column: 34},
					},
				},
			},
		},
	}

	rule := NewTerraformLifecycleRedundantDefaultRule()

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
