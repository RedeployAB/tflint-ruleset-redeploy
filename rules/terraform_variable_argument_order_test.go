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
					Message: "Out-of-order argument 'type'. Expected sequence: description, type, default, sensitive, nullable, validation",
					Range: hcl.Range{
						Filename: "variables.tf",
						Start:    hcl.Pos{Line: 5, Column: 3},
						End:      hcl.Pos{Line: 5, Column: 14},
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
					Message: "Out-of-order argument 'validation'. Expected sequence: description, type, default, sensitive, nullable, validation",
					Range: hcl.Range{
						Filename: "variables.tf",
						Start:    hcl.Pos{Line: 6, Column: 1},
						End:      hcl.Pos{Line: 6, Column: 1},
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
