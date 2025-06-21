package rules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func TestTerraformOutputArgumentOrderRule(t *testing.T) {
	tests := []struct {
		Name    string
		Content string
		Issues  helper.Issues
	}{
		{
			Name: "OK - minimal (only value)",
			Content: `
output "min_output" {
	value = "just a test"
}
`,
			Issues: helper.Issues{},
		},
		{
			Name: "OK - all attributes in correct order",
			Content: `
output "full_output" {
	description = "some desc"
	value       = "some val"
	ephemeral   = true
	sensitive   = true

	depends_on = []
}
`,
			Issues: helper.Issues{},
		},
		{
			Name: "OK - with precondition block",
			Content: `
output "with_precondition" {
	description = "Output with validation"
	value       = var.test
	sensitive   = true

	precondition {
		condition     = var.test != ""
		error_message = "Test must not be empty"
	}

	depends_on = [aws_instance.example]
}
`,
			Issues: helper.Issues{},
		},
		{
			Name: "NOT OK - 'sensitive' comes before 'ephemeral'",
			Content: `
output "bad_order" {

	description = "some desc"

	value = "some val"

	sensitive = true
	ephemeral = true
}
`,
			Issues: helper.Issues{
				{
					Rule:    NewTerraformOutputArgumentOrderRule(),
					Message: "Out-of-order argument 'ephemeral'. Expected sequence: description, value, ephemeral, sensitive, precondition, depends_on",
					Range: hcl.Range{
						Filename: "outputs.tf",
						Start:    hcl.Pos{Line: 9, Column: 2},
						End:      hcl.Pos{Line: 9, Column: 18},
					},
				},
			},
		},
		{
			Name: "NOT OK - precondition before value",
			Content: `
output "bad_precondition_order" {
	description = "Test"
	precondition {
		condition     = var.test != ""
		error_message = "Must not be empty"
	}
	value = var.test
}
`,
			Issues: helper.Issues{
				{
					Rule:    NewTerraformOutputArgumentOrderRule(),
					Message: "Out-of-order argument 'value'. Expected sequence: description, value, ephemeral, sensitive, precondition, depends_on",
					Range: hcl.Range{
						Filename: "outputs.tf",
						Start:    hcl.Pos{Line: 8, Column: 2},
						End:      hcl.Pos{Line: 8, Column: 18},
					},
				},
			},
		},
		{
			Name: "NOT OK - depends_on before precondition",
			Content: `
output "bad_depends_order" {
	value      = var.test
	depends_on = [aws_instance.example]
	precondition {
		condition     = var.test != ""
		error_message = "Must not be empty"
	}
}
`,
			Issues: helper.Issues{
				{
					Rule:    NewTerraformOutputArgumentOrderRule(),
					Message: "Out-of-order argument 'precondition'. Expected sequence: description, value, ephemeral, sensitive, precondition, depends_on",
					Range: hcl.Range{
						Filename: "outputs.tf",
						Start:    hcl.Pos{Line: 5, Column: 2},
						End:      hcl.Pos{Line: 5, Column: 14},
					},
				},
			},
		},
	}

	rule := NewTerraformOutputArgumentOrderRule()

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runner := helper.TestRunner(t, map[string]string{
				"outputs.tf": tc.Content,
			})

			if err := rule.Check(runner); err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			helper.AssertIssues(t, tc.Issues, runner.Issues)
		})
	}
}

func TestTerraformOutputArgumentOrderRule_Autofix(t *testing.T) {
	tests := []struct {
		Name         string
		ContentFile  string
		ExpectedFile string
	}{
		{
			Name:         "Autofix - sensitive before ephemeral",
			ContentFile:  "output_arg_order_autofix_sensitive_before_ephemeral.tf",
			ExpectedFile: "output_arg_order_autofix_sensitive_before_ephemeral_expected.tf",
		},
		{
			Name:         "Autofix - value before description",
			ContentFile:  "output_arg_order_autofix_value_before_desc.tf",
			ExpectedFile: "output_arg_order_autofix_value_before_desc_expected.tf",
		},
		{
			Name:         "Autofix - complex reordering",
			ContentFile:  "output_arg_order_autofix_complex.tf",
			ExpectedFile: "output_arg_order_autofix_complex_expected.tf",
		},
		{
			Name:         "Autofix - with blank lines",
			ContentFile:  "output_arg_order_autofix_with_blank_lines.tf",
			ExpectedFile: "output_arg_order_autofix_with_blank_lines_expected.tf",
		},
		{
			Name:         "Autofix - with precondition block",
			ContentFile:  "output_arg_order_autofix_with_precondition.tf",
			ExpectedFile: "output_arg_order_autofix_with_precondition_expected.tf",
		},
		{
			Name:         "Autofix - with depends_on",
			ContentFile:  "output_arg_order_autofix_with_depends.tf",
			ExpectedFile: "output_arg_order_autofix_with_depends_expected.tf",
		},
		{
			Name:         "Autofix - full reordering with precondition and depends_on",
			ContentFile:  "output_arg_order_autofix_full_reorder.tf",
			ExpectedFile: "output_arg_order_autofix_full_reorder_expected.tf",
		},
	}

	rule := NewTerraformOutputArgumentOrderRule()

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			content := readFixture(t, tc.ContentFile)
			expected := readFixture(t, tc.ExpectedFile)

			runner := helper.TestRunner(t, map[string]string{
				"outputs.tf": content,
			})

			if err := rule.Check(runner); err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			helper.AssertChanges(t, map[string]string{
				"outputs.tf": expected,
			}, runner.Changes())
		})
	}
}
