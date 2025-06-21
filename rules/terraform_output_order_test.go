package rules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func TestTerraformOutputOrderRule(t *testing.T) {
	tests := []struct {
		Name    string
		Content string
		Issues  helper.Issues
	}{
		{
			Name: "OK - single output only",
			Content: `
output "foo" {
	value = "one"
}
`,
			Issues: helper.Issues{},
		},
		{
			Name: "OK - multiple outputs in alphabetical order",
			Content: `
output "alpha" {}
output "beta" {}
output "zzz" {}
`,
			Issues: helper.Issues{},
		},
		{
			Name: "NOT OK - out of alphabetical order",
			Content: `
output "zzz" {}
output "alpha" {}
`,
			Issues: helper.Issues{
				{
					Rule:    NewTerraformOutputOrderRule(),
					Message: `Out-of-order output "alpha". Output blocks must be alphabetically ordered by name.`,
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 3, Column: 1},
						End:      hcl.Pos{Line: 3, Column: 15},
					},
				},
			},
		},
	}

	rule := NewTerraformOutputOrderRule()
	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runner := helper.TestRunner(t, map[string]string{
				"test.tf": tc.Content,
			})
			if err := rule.Check(runner); err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			helper.AssertIssues(t, tc.Issues, runner.Issues)
		})
	}
}

func TestTerraformOutputOrderRule_Autofix(t *testing.T) {
	tests := []struct {
		Name         string
		ContentFile  string
		ExpectedFile string
	}{
		{
			Name:         "Autofix - simple out of order",
			ContentFile:  "output_order_autofix_simple.tf",
			ExpectedFile: "output_order_autofix_simple_expected.tf",
		},
		{
			Name:         "Autofix - multiple outputs out of order",
			ContentFile:  "output_order_autofix_multiple.tf",
			ExpectedFile: "output_order_autofix_multiple_expected.tf",
		},
		{
			Name:         "Autofix - preserve spacing between outputs",
			ContentFile:  "output_order_autofix_preserve_spacing.tf",
			ExpectedFile: "output_order_autofix_preserve_spacing_expected.tf",
		},
		{
			Name:         "Autofix - preserve single line spacing",
			ContentFile:  "output_order_autofix_single_line.tf",
			ExpectedFile: "output_order_autofix_single_line_expected.tf",
		},
		{
			Name:         "Autofix - complex mix with different spacing",
			ContentFile:  "output_order_autofix_complex_mix.tf",
			ExpectedFile: "output_order_autofix_complex_mix_expected.tf",
		},
		{
			Name:         "Autofix - outputs with complex values",
			ContentFile:  "output_order_autofix_complex_values.tf",
			ExpectedFile: "output_order_autofix_complex_values_expected.tf",
		},
	}

	rule := NewTerraformOutputOrderRule()
	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			content := readFixture(t, tc.ContentFile)
			expected := readFixture(t, tc.ExpectedFile)

			runner := helper.TestRunner(t, map[string]string{
				"test.tf": content,
			})

			if err := rule.Check(runner); err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			helper.AssertChanges(t, map[string]string{
				"test.tf": expected,
			}, runner.Changes())
		})
	}
}
