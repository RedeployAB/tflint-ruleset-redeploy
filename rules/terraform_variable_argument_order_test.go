package rules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func TestTerraformVariableArgumentOrderRule(t *testing.T) {
	tests := []struct {
		Name   string
		File   string
		Issues helper.Issues
	}{
		{
			Name:   "OK - only description and type",
			File:   "variable_arg_order_ok_minimal.tf",
			Issues: helper.Issues{},
		},
		{
			Name:   "OK - all attributes in correct order",
			File:   "variable_arg_order_ok_full.tf",
			Issues: helper.Issues{},
		},
		{
			Name: "NOT OK - type after default",
			File: "variable_arg_order_not_ok_type_after_default.tf",
			Issues: helper.Issues{
				{
					Rule:    NewTerraformVariableArgumentOrderRule(),
					Message: "Out-of-order argument 'type'. Expected sequence: description, type, default, ephemeral, sensitive, nullable, validation",
					Range: hcl.Range{
						Filename: "variables.tf",
						Start:    hcl.Pos{Line: 4, Column: 3},
						End:      hcl.Pos{Line: 4, Column: 23},
					},
				},
			},
		},
		{
			Name: "NOT OK - validation block before default",
			File: "variable_arg_order_not_ok_validation_before_default.tf",
			Issues: helper.Issues{
				{
					Rule:    NewTerraformVariableArgumentOrderRule(),
					Message: "Out-of-order argument 'default'. Expected sequence: description, type, default, ephemeral, sensitive, nullable, validation",
					Range: hcl.Range{
						Filename: "variables.tf",
						Start:    hcl.Pos{Line: 10, Column: 3},
						End:      hcl.Pos{Line: 10, Column: 25},
					},
				},
			},
		},
		{
			Name:   "OK - multiple validation blocks after everything else",
			File:   "variable_arg_order_ok_multiple_validation.tf",
			Issues: helper.Issues{},
		},
	}

	rule := NewTerraformVariableArgumentOrderRule()

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			content := readFixture(t, tc.File)
			runner := helper.TestRunner(t, map[string]string{
				"variables.tf": content,
			})

			// Check the rule
			if err := rule.Check(runner); err != nil {
				t.Fatalf("Unexpected error: %s", err)
			}

			helper.AssertIssues(t, tc.Issues, runner.Issues)
		})
	}
}

func TestTerraformVariableArgumentOrderRule_Autofix(t *testing.T) {
	tests := []struct {
		Name         string
		ContentFile  string
		ExpectedFile string
	}{
		{
			Name:         "Autofix - type after default",
			ContentFile:  "variable_arg_order_autofix_type_after_default.tf",
			ExpectedFile: "variable_arg_order_autofix_type_after_default_expected.tf",
		},
		{
			Name:         "Autofix - nullable before sensitive",
			ContentFile:  "variable_arg_order_autofix_nullable_before_sensitive.tf",
			ExpectedFile: "variable_arg_order_autofix_nullable_before_sensitive_expected.tf",
		},
		{
			Name:         "Autofix - validation before default",
			ContentFile:  "variable_arg_order_autofix_validation_before_default.tf",
			ExpectedFile: "variable_arg_order_autofix_validation_before_default_expected.tf",
		},
		{
			Name:         "Autofix - complex out of order",
			ContentFile:  "variable_arg_order_autofix_complex.tf",
			ExpectedFile: "variable_arg_order_autofix_complex_expected.tf",
		},
		{
			Name:         "Autofix - ephemeral in wrong position",
			ContentFile:  "variable_arg_order_autofix_ephemeral_wrong.tf",
			ExpectedFile: "variable_arg_order_autofix_ephemeral_wrong_expected.tf",
		},
		{
			Name:         "Autofix - multiple validation blocks",
			ContentFile:  "variable_arg_order_autofix_multi_validation.tf",
			ExpectedFile: "variable_arg_order_autofix_multi_validation_expected.tf",
		},
		{
			Name:         "Autofix - all attributes with validation",
			ContentFile:  "variable_arg_order_autofix_full_example.tf",
			ExpectedFile: "variable_arg_order_autofix_full_example_expected.tf",
		},
	}

	rule := NewTerraformVariableArgumentOrderRule()

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			content := readFixture(t, tc.ContentFile)
			expected := readFixture(t, tc.ExpectedFile)

			runner := helper.TestRunner(t, map[string]string{
				"variables.tf": content,
			})

			if err := rule.Check(runner); err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			helper.AssertChanges(t, map[string]string{
				"variables.tf": expected,
			}, runner.Changes())
		})
	}
}
