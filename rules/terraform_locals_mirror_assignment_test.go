package rules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func TestTerraformLocalsMirrorAssignmentRule(t *testing.T) {
	tests := []struct {
		Name    string
		Content string
		Issues  helper.Issues
	}{
		{
			Name:    "NOT OK - local name differs => direct assignment not allowed",
			Content: readFixture(t, "locals_mirror_assignment_not_ok_local_name_differs.tf"),
			Issues: helper.Issues{
				{
					Rule:    NewTerraformLocalsMirrorAssignmentRule(),
					Message: "Local 'bar' is assigned directly from variable 'foo'. This should not be a simple mirror assignment.",
					Range: hcl.Range{
						Filename: "locals.tf",
						Start:    hcl.Pos{Line: 4, Column: 3},
						End:      hcl.Pos{Line: 4, Column: 16},
					},
				},
			},
		},
		{
			Name:    "OK - same local name, but uses an expression => no issues",
			Content: readFixture(t, "locals_mirror_assignment_ok_same_name_expression.tf"),
			Issues:  helper.Issues{},
		},
		{
			Name:    "NOT OK - direct mirror assignment",
			Content: readFixture(t, "locals_mirror_assignment_not_ok_direct_mirror.tf"),
			Issues: helper.Issues{
				{
					Rule:    NewTerraformLocalsMirrorAssignmentRule(),
					Message: "Local 'env' is assigned directly from variable 'env'. This should not be a simple mirror assignment.",
					Range: hcl.Range{
						Filename: "locals.tf",
						Start:    hcl.Pos{Line: 6, Column: 3},
						End:      hcl.Pos{Line: 6, Column: 16},
					},
				},
			},
		},
	}

	rule := NewTerraformLocalsMirrorAssignmentRule()

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runner := helper.TestRunner(t, map[string]string{
				"locals.tf": tc.Content,
			})

			if err := rule.Check(runner); err != nil {
				t.Fatalf("Unexpected error: %s", err)
			}

			helper.AssertIssues(t, tc.Issues, runner.Issues)
		})
	}
}

func TestTerraformLocalsMirrorAssignmentRule_Autofix(t *testing.T) {
	tests := []struct {
		Name     string
		Content  string
		Expected string
	}{
		{
			Name:     "Autofix - remove direct mirror assignment",
			Content:  readFixture(t, "locals_mirror_assignment_autofix_remove_direct.tf"),
			Expected: readFixture(t, "locals_mirror_assignment_autofix_remove_direct_expected.tf"),
		},
		{
			Name:     "Autofix - remove multiple mirror assignments",
			Content:  readFixture(t, "locals_mirror_assignment_autofix_remove_multiple.tf"),
			Expected: readFixture(t, "locals_mirror_assignment_autofix_remove_multiple_expected.tf"),
		},
		{
			Name:     "Autofix - preserve valid expressions",
			Content:  readFixture(t, "locals_mirror_assignment_autofix_preserve_valid.tf"),
			Expected: readFixture(t, "locals_mirror_assignment_autofix_preserve_valid_expected.tf"),
		},
	}

	rule := NewTerraformLocalsMirrorAssignmentRule()

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runner := helper.TestRunner(t, map[string]string{
				"locals.tf": tc.Content,
			})

			if err := rule.Check(runner); err != nil {
				t.Fatalf("Unexpected error: %s", err)
			}

			// Check that we have issues and they are fixable
			if len(runner.Issues) == 0 {
				t.Fatal("Expected issues to be found, but none were found")
			}

			// Apply autofixes by triggering the fix functions
			// The helper runner should automatically apply fixes when EmitIssueWithFix is called
			changes := runner.Changes()

			// Use AssertChanges to verify the fixes were applied correctly
			helper.AssertChanges(t, map[string]string{
				"locals.tf": tc.Expected,
			}, changes)
		})
	}
}
