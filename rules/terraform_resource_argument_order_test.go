package rules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

// TestTerraformResourceArgumentOrder provides coverage for TerraformResourceArgumentOrderRule.
// Each test references a .tf fixture in testdata/, checking whether any issues appear.
func TestTerraformResourceArgumentOrder(t *testing.T) {
	tests := []struct {
		Name     string
		FileName string
		Expected helper.Issues
	}{
		{
			Name:     "Valid - only non-block arguments",
			FileName: "resource_arg_order_ok_no_blocks.tf",
			Expected: helper.Issues{},
		},
		{
			Name:     "Invalid - provider features block is first",
			FileName: "resource_arg_order_invalid_provider_block_first.tf",
			Expected: helper.Issues{
				{
					Rule:    NewTerraformResourceArgumentOrderRule(),
					Message: "Argument 'skip_provider_registration' must not come after a nested block",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 4, Column: 3},
						End:      hcl.Pos{Line: 4, Column: 36},
					},
				},
			},
		},
		{
			Name:     "Valid - provider block after non-block argument",
			FileName: "resource_arg_order_ok_provider_block_after_nonblock.tf",
			Expected: helper.Issues{},
		},
		{
			Name:     "Valid - block comes after normal arguments (azurerm_firewall)",
			FileName: "resource_arg_order_ok_block_after_args.tf",
			Expected: helper.Issues{},
		},
		{
			Name:     "Valid - nested block after normal arguments (azurerm_firewall_application_rule_collection)",
			FileName: "resource_arg_order_ok_nested_block_after_args.tf",
			Expected: helper.Issues{},
		},
		{
			Name:     "Invalid - attribute after block (azurerm_firewall)",
			FileName: "resource_arg_order_invalid_attr_after_block.tf",
			Expected: helper.Issues{
				{
					Rule:    NewTerraformResourceArgumentOrderRule(),
					Message: "Argument 'resource_group_name' must not come after a nested block",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 10, Column: 3},
						End:      hcl.Pos{Line: 10, Column: 29},
					},
				},
				{
					Rule:    NewTerraformResourceArgumentOrderRule(),
					Message: "Argument 'sku_name' must not come after a nested block",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 11, Column: 3},
						End:      hcl.Pos{Line: 11, Column: 36},
					},
				},
				{
					Rule:    NewTerraformResourceArgumentOrderRule(),
					Message: "Argument 'sku_tier' must not come after a nested block",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 12, Column: 3},
						End:      hcl.Pos{Line: 12, Column: 35},
					},
				},
			},
		},
		{
			Name:     "Valid - tags argument is not flagged",
			FileName: "resource_arg_order_ok_tags.tf",
			Expected: helper.Issues{},
		},
		{
			Name:     "Invalid - attribute after nested block in nested block",
			FileName: "resource_arg_order_invalid_nested_attr_after_block.tf",
			Expected: helper.Issues{
				{
					Rule:    NewTerraformResourceArgumentOrderRule(),
					Message: "Argument 'versioning_enabled' must not come after a nested block",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 16, Column: 5},
						End:      hcl.Pos{Line: 16, Column: 31},
					},
				},
			},
		},
	}

	rule := NewTerraformResourceArgumentOrderRule()
	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			content := readFixture(t, tc.FileName)
			runner := helper.TestRunner(t, map[string]string{
				"resource.tf": content,
			})

			if err := rule.Check(runner); err != nil {
				t.Fatalf("Unexpected error occurred: %v", err)
			}
			helper.AssertIssues(t, tc.Expected, runner.Issues)
		})
	}
}
