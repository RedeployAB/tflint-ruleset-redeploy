package rules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func TestTerraformSingleBlankLinesRule(t *testing.T) {
	tests := []struct {
		Name    string
		Content string
		Issues  helper.Issues
	}{
		{
			Name: "OK - single blank line only",
			Content: `
resource "random_uuid" "role_assignment" {
  for_each = local.role_assignments

  lifecycle {
    replace_triggered_by = [
      null_resource.role_assignment[each.key]
    ]
  }
}
`,
			Issues: helper.Issues{},
		},
		{
			Name: "OK - no blank lines at all",
			Content: `
resource "random_uuid" "role_assignment" {
  lifecycle {
    replace_triggered_by = [
      null_resource.role_assignment[each.key]
    ]
  }
}
`,
			Issues: helper.Issues{},
		},
		{
			Name: "NOT OK - two consecutive blank lines",
			Content: `
resource "random_uuid" "role_assignment" {
  for_each = local.role_assignments


  lifecycle {
    replace_triggered_by = [
      null_resource.role_assignment[each.key]
    ]
  }
}
`,
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
			runner := helper.TestRunner(t, map[string]string{
				"resource.tf": tc.Content,
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
