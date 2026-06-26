package rules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func TestTerraformRedundantDefaultRule(t *testing.T) {
	sensitiveIssue := &helper.Issue{
		Rule:    NewTerraformRedundantDefaultRule(),
		Message: "sensitive should not be set to false (omit instead)",
		Range: hcl.Range{
			Filename: "resource.tf",
			Start:    hcl.Pos{Line: 4, Column: 3},
			End:      hcl.Pos{Line: 4, Column: 22},
		},
	}
	ephemeralIssue := &helper.Issue{
		Rule:    NewTerraformRedundantDefaultRule(),
		Message: "ephemeral should not be set to false (omit instead)",
		Range: hcl.Range{
			Filename: "resource.tf",
			Start:    hcl.Pos{Line: 10, Column: 3},
			End:      hcl.Pos{Line: 10, Column: 22},
		},
	}
	preventDestroyIssue := &helper.Issue{
		Rule:    NewTerraformRedundantDefaultRule(),
		Message: "prevent_destroy should not be set to false (omit instead)",
		Range: hcl.Range{
			Filename: "resource.tf",
			Start:    hcl.Pos{Line: 15, Column: 5},
			End:      hcl.Pos{Line: 15, Column: 34},
		},
	}
	createBeforeDestroyIssue := &helper.Issue{
		Rule:    NewTerraformRedundantDefaultRule(),
		Message: "create_before_destroy should not be set to false (omit instead)",
		Range: hcl.Range{
			Filename: "resource.tf",
			Start:    hcl.Pos{Line: 16, Column: 5},
			End:      hcl.Pos{Line: 16, Column: 34},
		},
	}

	tests := []struct {
		Name     string
		Content  string
		Config   string
		Expected helper.Issues
	}{
		{
			Name:     "Non-default and non-literal values are valid",
			Content:  readFixture(t, "redundant_default_valid.tf"),
			Expected: helper.Issues{},
		},
		{
			Name:     "Redundant false defaults across contexts",
			Content:  readFixture(t, "redundant_default_invalid.tf"),
			Expected: helper.Issues{sensitiveIssue, ephemeralIssue, preventDestroyIssue, createBeforeDestroyIssue},
		},
		{
			Name:    "create_before_destroy check disabled via config",
			Content: readFixture(t, "redundant_default_invalid.tf"),
			Config: `rule "terraform_redundant_default" {
enabled = true
create_before_destroy = false
}`,
			Expected: helper.Issues{sensitiveIssue, ephemeralIssue, preventDestroyIssue},
		},
	}

	rule := NewTerraformRedundantDefaultRule()

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			files := map[string]string{"resource.tf": tc.Content}
			if tc.Config != "" {
				files[".tflint.hcl"] = tc.Config
			}
			runner := helper.TestRunner(t, files)
			if err := rule.Check(runner); err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			helper.AssertIssues(t, tc.Expected, runner.Issues)
		})
	}
}

func TestTerraformRedundantDefaultRule_Autofix(t *testing.T) {
	tests := []struct {
		Name     string
		Content  string
		Expected string
	}{
		{
			Name:     "Autofix removes redundant false defaults across contexts",
			Content:  readFixture(t, "redundant_default_autofix.tf"),
			Expected: readFixture(t, "redundant_default_autofix_expected.tf"),
		},
	}

	rule := NewTerraformRedundantDefaultRule()

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runner := helper.TestRunner(t, map[string]string{
				"resource.tf": tc.Content,
			})
			if err := rule.Check(runner); err != nil {
				t.Fatalf("Unexpected error: %s", err)
			}
			if len(runner.Issues) == 0 {
				t.Fatal("Expected issues to be found, but none were found")
			}
			helper.AssertChanges(t, map[string]string{
				"resource.tf": tc.Expected,
			}, runner.Changes())
		})
	}
}
