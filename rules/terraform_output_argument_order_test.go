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
		Name     string
		Content  string
		Expected string
	}{
		{
			Name: "Autofix - sensitive before ephemeral",
			Content: `output "bad_order" {
	description = "some desc"
	value = "some val"
	sensitive = true
	ephemeral = true
}
`,
			Expected: `output "bad_order" {
  description = "some desc"
  value       = "some val"
  ephemeral   = true
  sensitive   = true
}
`,
		},
		{
			Name: "Autofix - value before description",
			Content: `output "test" {
	value = "test value"
	description = "test description"
}
`,
			Expected: `output "test" {
  description = "test description"
  value       = "test value"
}
`,
		},
		{
			Name: "Autofix - complex reordering",
			Content: `output "complex" {
	depends_on = []
	sensitive = true
	value = "value"
	description = "desc"
}
`,
			Expected: `output "complex" {
  description = "desc"
  value       = "value"
  sensitive   = true

  depends_on = []
}
`,
		},
		{
			Name: "Autofix - with blank lines",
			Content: `output "spaced" {
	sensitive = true

	value = "val"

	description = "desc"
}
`,
			Expected: `output "spaced" {
  description = "desc"
  value       = "val"
  sensitive   = true
}
`,
		},
		{
			Name: "Autofix - with precondition block",
			Content: `output "with_precondition" {
	value = var.test
	precondition {
		condition = var.test != ""
		error_message = "Test must not be empty"
	}
	description = "Output with validation"
}
`,
			Expected: `output "with_precondition" {
  description = "Output with validation"
  value       = var.test

  precondition {
    condition     = var.test != ""
    error_message = "Test must not be empty"
  }
}
`,
		},
		{
			Name: "Autofix - with depends_on",
			Content: `output "with_depends" {
	depends_on = [aws_instance.example]
	value = var.instance_id
	description = "Instance ID"
}
`,
			Expected: `output "with_depends" {
  description = "Instance ID"
  value       = var.instance_id

  depends_on = [aws_instance.example]
}
`,
		},
		{
			Name: "Autofix - full reordering with precondition and depends_on",
			Content: `output "full_reorder" {
	depends_on = [aws_instance.example]
	sensitive = true
	precondition {
		condition = var.test != ""
		error_message = "Test required"
	}
	value = var.test
	description = "Full test"
	ephemeral = true
}
`,
			Expected: `output "full_reorder" {
  description = "Full test"
  value       = var.test
  ephemeral   = true
  sensitive   = true

  precondition {
    condition     = var.test != ""
    error_message = "Test required"
  }

  depends_on = [aws_instance.example]
}
`,
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

			helper.AssertChanges(t, map[string]string{
				"outputs.tf": tc.Expected,
			}, runner.Changes())
		})
	}
}
