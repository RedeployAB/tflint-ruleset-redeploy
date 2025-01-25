package rules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func TestTerraformVariableSensitiveRule(t *testing.T) {
	tests := []struct {
		Name   string
		File   string
		Issues helper.Issues
	}{
		{
			Name:   "OK - no sensitive declared",
			File:   "variable_sensitive_ok_none.tf",
			Issues: helper.Issues{},
		},
		{
			Name:   "OK - sensitive = true",
			File:   "variable_sensitive_ok_true.tf",
			Issues: helper.Issues{},
		},
		{
			Name: "NOT OK - sensitive = false",
			File: "variable_sensitive_not_ok_false.tf",
			Issues: helper.Issues{
				{
					Rule:    NewTerraformVariableSensitiveRule(),
					Message: "sensitive should not be set to false (omit instead)",
					Range: hcl.Range{
						Filename: "variables.tf",
						Start:    hcl.Pos{Line: 4, Column: 3},
						End:      hcl.Pos{Line: 4, Column: 21},
					},
				},
			},
		},
	}

	rule := NewTerraformVariableSensitiveRule()

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			content := readFixture(t, tc.File)
			runner := helper.TestRunner(t, map[string]string{
				"variables.tf": content,
			})

			if err := rule.Check(runner); err != nil {
				t.Fatalf("Unexpected error: %s", err)
			}

			helper.AssertIssues(t, tc.Issues, runner.Issues)
		})
	}
}
