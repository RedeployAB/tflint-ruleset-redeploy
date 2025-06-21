package rules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func TestTerraformVariableNullableRule(t *testing.T) {
	tests := []struct {
		Name   string
		File   string
		Issues helper.Issues
	}{
		{
			Name:   "OK - nullable = false, bool default = true",
			File:   "variable_nullable_ok_bool_default_true.tf",
			Issues: helper.Issues{},
		},
		{
			Name:   "OK - no default, nullable = false",
			File:   "variable_nullable_ok_no_default.tf",
			Issues: helper.Issues{},
		},
		{
			Name: "NOT OK - nullable set to true",
			File: "variable_nullable_not_ok_nullable_true.tf",
			Issues: helper.Issues{
				{
					Rule:    NewTerraformVariableNullableRule(),
					Message: "nullable should not be set to true (the default is already true)",
					Range: hcl.Range{
						Filename: "variables.tf",
						Start:    hcl.Pos{Line: 2, Column: 3},
						End:      hcl.Pos{Line: 2, Column: 18},
					},
				},
			},
		},
		{
			Name: "NOT OK - boolean var with default = null",
			File: "variable_nullable_not_ok_bool_default_null.tf",
			Issues: helper.Issues{
				{
					Rule:    NewTerraformVariableNullableRule(),
					Message: "boolean variables cannot have default = null",
					Range: hcl.Range{
						Filename: "variables.tf",
						Start:    hcl.Pos{Line: 4, Column: 3},
						End:      hcl.Pos{Line: 4, Column: 17},
					},
				},
			},
		},
		{
			Name: "NOT OK - default = null but has nullable declared",
			File: "variable_nullable_not_ok_default_null_with_nullable.tf",
			Issues: helper.Issues{
				{
					Rule:    NewTerraformVariableNullableRule(),
					Message: "nullable must not be declared if default = null",
					Range: hcl.Range{
						Filename: "variables.tf",
						Start:    hcl.Pos{Line: 3, Column: 3},
						End:      hcl.Pos{Line: 3, Column: 19},
					},
				},
			},
		},
	}

	rule := NewTerraformVariableNullableRule()
	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			content := readFixture(t, tc.File)
			runner := helper.TestRunner(t, map[string]string{"variables.tf": content})
			// Execute rule
			if err := rule.Check(runner); err != nil {
				t.Fatalf("Unexpected error: %s", err)
			}
			helper.AssertIssues(t, tc.Issues, runner.Issues)
		})
	}
}

func TestTerraformVariableNullableRuleAutofix(t *testing.T) {
	tests := []struct {
		Name         string
		ContentFile  string
		ExpectedFile string
	}{
		{
			Name:         "Fix - remove nullable = true",
			ContentFile:  "variable_nullable_autofix_remove_true.tf",
			ExpectedFile: "variable_nullable_autofix_remove_true_expected.tf",
		},
		{
			Name:         "Fix - remove nullable when default = null",
			ContentFile:  "variable_nullable_autofix_remove_with_default_null.tf",
			ExpectedFile: "variable_nullable_autofix_remove_with_default_null_expected.tf",
		},
		{
			Name:         "Fix - remove nullable = true with other attributes",
			ContentFile:  "variable_nullable_autofix_remove_true_with_other_attrs.tf",
			ExpectedFile: "variable_nullable_autofix_remove_true_with_other_attrs_expected.tf",
		},
	}

	rule := NewTerraformVariableNullableRule()
	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			content := readFixture(t, tc.ContentFile)
			runner := helper.TestRunner(t, map[string]string{"variables.tf": content})

			// Execute rule
			if err := rule.Check(runner); err != nil {
				t.Fatalf("Unexpected error: %s", err)
			}

			// Check that we have issues
			if len(runner.Issues) == 0 {
				t.Fatalf("Expected issues but got none")
			}

			// Check the autofix
			changes := runner.Changes()
			expected := readFixture(t, tc.ExpectedFile)
			helper.AssertChanges(t, map[string]string{"variables.tf": expected}, changes)
		})
	}
}
