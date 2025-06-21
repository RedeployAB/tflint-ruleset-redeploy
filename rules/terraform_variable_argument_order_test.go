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
		Name     string
		Content  string
		Expected string
	}{
		{
			Name: "Autofix - type after default",
			Content: `variable "fail_type_after_default" {
  description = "Out-of-order example"
  default     = "some_value"
  type        = string
}
`,
			Expected: `variable "fail_type_after_default" {
  description = "Out-of-order example"
  type        = string
  default     = "some_value"
}
`,
		},
		{
			Name: "Autofix - nullable before sensitive",
			Content: `variable "test" {
  nullable = false
  sensitive = true
}
`,
			Expected: `variable "test" {
  sensitive = true
  nullable  = false
}
`,
		},
		{
			Name: "Autofix - validation before default",
			Content: `variable "test" {
  description = "Test variable"
  type        = string
  validation {
    condition     = length(var.test) > 0
    error_message = "Must not be empty"
  }
  default = "value"
}
`,
			Expected: `variable "test" {
  description = "Test variable"
  type        = string
  default     = "value"

  validation {
    condition     = length(var.test) > 0
    error_message = "Must not be empty"
  }
}
`,
		},
		{
			Name: "Autofix - complex out of order",
			Content: `variable "complex" {
  nullable = false
  description = "Complex example"
  sensitive = true
  type = list(string)
  default = ["a", "b"]
}
`,
			Expected: `variable "complex" {
  description = "Complex example"
  type        = list(string)
  default     = ["a", "b"]
  sensitive   = true
  nullable    = false
}
`,
		},
		{
			Name: "Autofix - ephemeral in wrong position",
			Content: `variable "ephemeral_test" {
  description = "Test"
  type = string
  sensitive = true
  ephemeral = true
  default = "value"
}
`,
			Expected: `variable "ephemeral_test" {
  description = "Test"
  type        = string
  default     = "value"
  ephemeral   = true
  sensitive   = true
}
`,
		},
		{
			Name: "Autofix - multiple validation blocks",
			Content: `variable "multi_validation" {
  type = string
  validation {
    condition = true
    error_message = "First"
  }
  default = "value"
  validation {
    condition = true
    error_message = "Second"
  }
}
`,
			Expected: `variable "multi_validation" {
  type    = string
  default = "value"

  validation {
    condition     = true
    error_message = "First"
  }
  validation {
    condition     = true
    error_message = "Second"
  }
}
`,
		},
		{
			Name: "Autofix - all attributes with validation",
			Content: `variable "full_example" {
  nullable = false
  validation {
    condition = var.full_example != ""
    error_message = "Cannot be empty"
  }
  type = string
  sensitive = true
  description = "Full example"
  default = "test"
  ephemeral = true
}
`,
			Expected: `variable "full_example" {
  description = "Full example"
  type        = string
  default     = "test"
  ephemeral   = true
  sensitive   = true
  nullable    = false

  validation {
    condition     = var.full_example != ""
    error_message = "Cannot be empty"
  }
}
`,
		},
	}

	rule := NewTerraformVariableArgumentOrderRule()

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runner := helper.TestRunner(t, map[string]string{
				"variables.tf": tc.Content,
			})

			if err := rule.Check(runner); err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			helper.AssertChanges(t, map[string]string{
				"variables.tf": tc.Expected,
			}, runner.Changes())
		})
	}
}
