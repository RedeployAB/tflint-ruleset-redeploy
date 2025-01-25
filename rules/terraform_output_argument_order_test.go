package rules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func TestTerraformOutputArgumentOrderRule(t *testing.T) {
	tests := []struct {
		Name   string
		File   string
		Issues helper.Issues
	}{
		{
			Name:   "OK - minimal (only value)",
			File:   "output_arg_order_ok_minimal.tf",
			Issues: helper.Issues{},
		},
		{
			Name:   "OK - all attributes in correct order",
			File:   "output_arg_order_ok_full.tf",
			Issues: helper.Issues{},
		},
		{
			Name: "NOT OK - 'sensitive' comes before 'ephemeral'",
			File: "output_arg_order_not_ok_sens_before_ephemeral.tf",
			Issues: helper.Issues{
				{
					Rule:    NewTerraformOutputArgumentOrderRule(),
					Message: "Out-of-order argument 'sensitive'. Expected sequence: description, value, ephemeral, sensitive, depends_on",
					Range: hcl.Range{
						Filename: "outputs.tf",
						Start:    hcl.Pos{Line: 7, Column: 3},
						End:      hcl.Pos{Line: 7, Column: 17},
					},
				},
			},
		},
	}

	rule := NewTerraformOutputArgumentOrderRule()

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			content := readFixture(t, tc.File)
			runner := helper.TestRunner(t, map[string]string{
				"outputs.tf": content,
			})

			if err := rule.Check(runner); err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			helper.AssertIssues(t, tc.Issues, runner.Issues)
		})
	}
}
