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
			File:   "variable_nullable_ok1.tf",
			Issues: helper.Issues{},
		},
		{
			Name:   "OK - no default, nullable = false",
			File:   "variable_nullable_ok2.tf",
			Issues: helper.Issues{},
		},
		{
			Name: "NOT OK - nullable set to true",
			File: "variable_nullable_not_ok1.tf",
			Issues: helper.Issues{
				{
					Rule:    NewTerraformVariableNullableRule(),
					Message: "nullable should not be set to true (the default is already true)",
					Range: hcl.Range{
						Filename: "variables.tf",
						Start:    hcl.Pos{Line: 7, Column: 1},
						End:      hcl.Pos{Line: 7, Column: 1},
					},
				},
			},
		},
		{
			Name: "NOT OK - boolean var with default = null",
			File: "variable_nullable_not_ok2.tf",
			Issues: helper.Issues{
				{
					Rule:    NewTerraformVariableNullableRule(),
					Message: "boolean variables cannot have default = null",
					Range: hcl.Range{
						Filename: "variables.tf",
						Start:    hcl.Pos{Line: 6, Column: 1},
						End:      hcl.Pos{Line: 6, Column: 1},
					},
				},
			},
		},
		{
			Name: "NOT OK - default = null but has nullable declared",
			File: "variable_nullable_not_ok3.tf",
			Issues: helper.Issues{
				{
					Rule:    NewTerraformVariableNullableRule(),
					Message: "nullable must not be declared if default = null",
					Range: hcl.Range{
						Filename: "variables.tf",
						Start:    hcl.Pos{Line: 8, Column: 1},
						End:      hcl.Pos{Line: 8, Column: 1},
					},
				},
			},
		},
	}

	rule := NewTerraformVariableNullableRule()
	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			content := readFixture(t, "testdata/"+tc.File)
			runner := helper.TestRunner(t, map[string]string{
				"variables.tf": content,
			})

			// Execute rule
			if err := rule.Check(runner); err != nil {
				t.Fatalf("Unexpected error: %s", err)
			}

			helper.AssertIssues(t, tc.Issues, runner.Issues)
		})
	}
}
