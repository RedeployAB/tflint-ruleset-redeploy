package rules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func TestTerraformIgnoreChangesAllRule(t *testing.T) {
	tests := []struct {
		Name     string
		Content  string
		Expected helper.Issues
	}{
		{
			Name:     "Explicit attribute lists are valid",
			Content:  readFixture(t, "ignore_changes_all_valid.tf"),
			Expected: helper.Issues{},
		},
		{
			Name:    "ignore_changes set to all",
			Content: readFixture(t, "ignore_changes_all_invalid.tf"),
			Expected: helper.Issues{
				{
					Rule:    NewTerraformIgnoreChangesAllRule(),
					Message: "Avoid 'ignore_changes = all'; list the specific attributes to ignore instead",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 3, Column: 5},
						End:      hcl.Pos{Line: 3, Column: 25},
					},
				},
			},
		},
	}

	rule := NewTerraformIgnoreChangesAllRule()

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
