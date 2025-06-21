package rules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func TestTerraformVariableOrderRule(t *testing.T) {
	tests := []struct {
		Name    string
		Content string
		Issues  helper.Issues
	}{
		{
			Name: "OK - all required in alphabetical order, then optional in alphabetical order",
			Content: `
variable "alpha" {}
variable "beta" {}
variable "delta" {
	default = true
}
variable "gamma" {
	default = "some default"
}
`,
			Issues: helper.Issues{},
		},
		{
			Name: "NOT OK - optional before required",
			Content: `
variable "bar" {
	default = 123
}
variable "foo" {}
`,
			Issues: helper.Issues{
				{
					Rule:    NewTerraformVariableOrderRule(),
					Message: `Out-of-order variable "foo". Required variables must come first in alphabetical order, followed by optional variables in alphabetical order.`,
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 5, Column: 1},
						End:      hcl.Pos{Line: 5, Column: 15},
					},
				},
			},
		},
		{
			Name: "NOT OK - required out of alphabetical order",
			Content: `
variable "zzz" {}
variable "aaa" {}
`,
			Issues: helper.Issues{
				{
					Rule:    NewTerraformVariableOrderRule(),
					Message: `Out-of-order variable "aaa". Required variables must come first in alphabetical order, followed by optional variables in alphabetical order.`,
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 3, Column: 1},
						End:      hcl.Pos{Line: 3, Column: 15},
					},
				},
			},
		},
		{
			Name: "NOT OK - optional out of alphabetical order",
			Content: `
variable "opt_x" {
	default = 1
}
variable "opt_a" {
	default = 2
}
`,
			Issues: helper.Issues{
				{
					Rule:    NewTerraformVariableOrderRule(),
					Message: `Out-of-order variable "opt_a". Required variables must come first in alphabetical order, followed by optional variables in alphabetical order.`,
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 5, Column: 1},
						End:      hcl.Pos{Line: 5, Column: 17},
					},
				},
			},
		},
	}

	rule := NewTerraformVariableOrderRule()
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

func TestTerraformVariableOrderRule_Autofix(t *testing.T) {
	tests := []struct {
		Name         string
		ContentFile  string
		ExpectedFile string
	}{
		{
			Name:         "Autofix - optional before required",
			ContentFile:  "variable_order_autofix_optional_before_required.tf",
			ExpectedFile: "variable_order_autofix_optional_before_required_expected.tf",
		},
		{
			Name:         "Autofix - required out of alphabetical order",
			ContentFile:  "variable_order_autofix_required_out_of_order.tf",
			ExpectedFile: "variable_order_autofix_required_out_of_order_expected.tf",
		},
		{
			Name:         "Autofix - optional out of alphabetical order",
			ContentFile:  "variable_order_autofix_optional_out_of_order.tf",
			ExpectedFile: "variable_order_autofix_optional_out_of_order_expected.tf",
		},
		{
			Name:         "Autofix - complex mix of required and optional",
			ContentFile:  "variable_order_autofix_complex_mix.tf",
			ExpectedFile: "variable_order_autofix_complex_mix_expected.tf",
		},
		{
			Name:         "Autofix - preserve spacing between variables",
			ContentFile:  "variable_order_autofix_preserve_spacing.tf",
			ExpectedFile: "variable_order_autofix_preserve_spacing_expected.tf",
		},
		{
			Name:         "Autofix - preserve single line spacing",
			ContentFile:  "variable_order_autofix_single_line.tf",
			ExpectedFile: "variable_order_autofix_single_line_expected.tf",
		},
		{
			Name:         "Autofix - no space between adjacent variables originally",
			ContentFile:  "variable_order_autofix_no_space.tf",
			ExpectedFile: "variable_order_autofix_no_space_expected.tf",
		},
	}

	rule := NewTerraformVariableOrderRule()
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
