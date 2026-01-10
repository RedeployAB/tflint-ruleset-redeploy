package rules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func TestTerraformBlockOrderRule(t *testing.T) {
	tests := []struct {
		Name        string
		FixtureFile string
		Issues      helper.Issues
	}{
		{
			Name:        "OK - all blocks in correct order",
			FixtureFile: "block_order_correct.tf",
			Issues:      helper.Issues{},
		},
		{
			Name:        "OK - skipping some blocks",
			FixtureFile: "block_order_skip_some.tf",
			Issues:      helper.Issues{},
		},
		{
			Name:        "NOT OK - 'provider' after 'resource'",
			FixtureFile: "block_order_provider_after_resource.tf",
			Issues: helper.Issues{
				{
					Rule:    NewTerraformBlockOrderRule(),
					Message: "Out-of-order block 'provider'. Expected order: terraform -> provider -> data -> resource",
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 5, Column: 1},
						End:      hcl.Pos{Line: 5, Column: 15},
					},
				},
			},
		},
		{
			Name:        "OK - moved block can appear anywhere",
			FixtureFile: "block_order_moved.tf",
			Issues:      helper.Issues{},
		},
		{
			Name:        "OK - import block can appear anywhere",
			FixtureFile: "block_order_import.tf",
			Issues:      helper.Issues{},
		},
		{
			Name:        "OK - removed block can appear anywhere",
			FixtureFile: "block_order_removed.tf",
			Issues:      helper.Issues{},
		},
		{
			Name:        "OK - check block can appear anywhere",
			FixtureFile: "block_order_check.tf",
			Issues:      helper.Issues{},
		},
		{
			Name:        "OK - mixed lifecycle blocks with ordered blocks",
			FixtureFile: "block_order_mixed_lifecycle.tf",
			Issues:      helper.Issues{},
		},
	}

	rule := NewTerraformBlockOrderRule()
	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runner := helper.TestRunner(t, map[string]string{
				"test.tf": readFixture(t, tc.FixtureFile),
			})
			if err := rule.Check(runner); err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			helper.AssertIssues(t, tc.Issues, runner.Issues)
		})
	}
}
