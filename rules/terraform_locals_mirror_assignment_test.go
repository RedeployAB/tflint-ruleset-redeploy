package rules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func TestTerraformLocalsMirrorAssignmentRule(t *testing.T) {
	tests := []struct {
		Name   string
		Files  map[string]string
		Issues helper.Issues
	}{
		{
			Name: "NOT OK - local name differs => direct assignment not allowed",
			Files: map[string]string{
				"locals.tf": `
variable "foo" {}

locals {
	bar = var.foo
}
`,
			},
			Issues: helper.Issues{
				{
					Rule:    NewTerraformLocalsMirrorAssignmentRule(),
					Message: "Local 'bar' is assigned directly from variable 'foo'. This should not be a simple mirror assignment.",
					Range: hcl.Range{
						Filename: "locals.tf",
						Start:    hcl.Pos{Line: 5, Column: 2},
						End:      hcl.Pos{Line: 5, Column: 15},
					},
				},
			},
		},
		{
			Name: "OK - same local name, but uses an expression => no issues",
			Files: map[string]string{
				"locals.tf": `
variable "hello" {}

locals {
	hello = lower(var.hello)
}
`,
			},
			Issues: helper.Issues{},
		},
		{
			Name: "NOT OK - direct mirror assignment",
			Files: map[string]string{
				"locals.tf": `
variable "env" {
	default = "dev"
}

locals {
	env = var.env
}
`,
			},
			Issues: helper.Issues{
				{
					Rule:    NewTerraformLocalsMirrorAssignmentRule(),
					Message: "Local 'env' is assigned directly from variable 'env'. This should not be a simple mirror assignment.",
					Range: hcl.Range{
						Filename: "locals.tf",
						Start:    hcl.Pos{Line: 7, Column: 2},
						End:      hcl.Pos{Line: 7, Column: 15},
					},
				},
			},
		},
	}

	rule := NewTerraformLocalsMirrorAssignmentRule()

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runner := helper.TestRunner(t, tc.Files)

			if err := rule.Check(runner); err != nil {
				t.Fatalf("Unexpected error: %s", err)
			}

			helper.AssertIssues(t, tc.Issues, runner.Issues)
		})
	}
}
