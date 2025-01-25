package rules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func TestTerraformMetaArgumentOrder(t *testing.T) {
	tests := []struct {
		Name     string
		Content  string
		Expected helper.Issues
	}{
		{
			Name:     "valid resource order",
			Content:  readFixture(t, "meta_order_valid_resource.tf"),
			Expected: helper.Issues{},
		},
		{
			Name:     "resource partial usage (count only)",
			Content:  readFixture(t, "meta_order_resource_count_only.tf"),
			Expected: helper.Issues{},
		},
		{
			Name:     "resource with no meta arguments",
			Content:  readFixture(t, "meta_order_resource_no_meta.tf"),
			Expected: helper.Issues{},
		},
		{
			Name:    "invalid resource order",
			Content: readFixture(t, "meta_order_invalid_resource.tf"),
			Expected: helper.Issues{
				{
					Rule:    NewTerraformMetaArgumentOrderRule(),
					Message: "Out-of-order meta argument 'depends_on' in resource 'aws_instance example'. Expected sequence: provider -> count|for_each -> lifecycle -> depends_on",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 2, Column: 3},
						End:      hcl.Pos{Line: 2, Column: 18},
					},
				},
			},
		},
		{
			Name:     "valid module order",
			Content:  readFixture(t, "meta_order_valid_module.tf"),
			Expected: helper.Issues{},
		},
		{
			Name:    "invalid module order",
			Content: readFixture(t, "meta_order_invalid_module.tf"),
			Expected: helper.Issues{
				{
					Rule:    NewTerraformMetaArgumentOrderRule(),
					Message: "Out-of-order meta argument 'depends_on' in module 'example'. Expected sequence: count|for_each -> depends_on",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 2, Column: 3},
						End:      hcl.Pos{Line: 2, Column: 18},
					},
				},
			},
		},
	}

	rule := NewTerraformMetaArgumentOrderRule()

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runner := helper.TestRunner(t, map[string]string{
				"resource.tf": tc.Content,
			})

			if err := rule.Check(runner); err != nil {
				t.Fatalf("Unexpected error occurred: %s", err)
			}

			helper.AssertIssues(t, tc.Expected, runner.Issues)
		})
	}
}
