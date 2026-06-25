package rules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func TestTerraformProviderAliasOrderRule(t *testing.T) {
	tests := []struct {
		Name     string
		Content  string
		Expected helper.Issues
	}{
		{
			Name:     "Alias first and default before aliased is valid",
			Content:  readFixture(t, "provider_alias_order_valid.tf"),
			Expected: helper.Issues{},
		},
		{
			Name:    "Alias not first and aliased before default",
			Content: readFixture(t, "provider_alias_order_invalid.tf"),
			Expected: helper.Issues{
				{
					Rule:    NewTerraformProviderAliasOrderRule(),
					Message: "Provider 'aws': 'alias' must be the first argument in the provider block",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 3, Column: 3},
						End:      hcl.Pos{Line: 3, Column: 8},
					},
				},
				{
					Rule:    NewTerraformProviderAliasOrderRule(),
					Message: "Provider 'google': default (un-aliased) provider must be declared before aliased providers",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 6, Column: 1},
						End:      hcl.Pos{Line: 6, Column: 18},
					},
				},
			},
		},
	}

	rule := NewTerraformProviderAliasOrderRule()

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

// Ordering of default vs aliased provider blocks is only defined within a
// single file. A default in one file and an alias in another must not be
// reported, otherwise the rule would produce false positives.
func TestTerraformProviderAliasOrderRule_CrossFile(t *testing.T) {
	rule := NewTerraformProviderAliasOrderRule()
	runner := helper.TestRunner(t, map[string]string{
		"providers.tf":       readFixture(t, "provider_alias_order_crossfile_default.tf"),
		"providers.extra.tf": readFixture(t, "provider_alias_order_crossfile_aliased.tf"),
	})
	if err := rule.Check(runner); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	helper.AssertIssues(t, helper.Issues{}, runner.Issues)
}
