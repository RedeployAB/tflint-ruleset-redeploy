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

func TestTerraformLocalsMirrorAssignmentRule_Autofix(t *testing.T) {
	tests := []struct {
		Name     string
		Files    map[string]string
		Expected map[string]string
	}{
		{
			Name: "Autofix - remove direct mirror assignment",
			Files: map[string]string{
				"locals.tf": `variable "foo" {}

locals {
	bar = var.foo
}
`,
			},
			Expected: map[string]string{
				"locals.tf": `variable "foo" {}

locals {
}
`,
			},
		},
		{
			Name: "Autofix - remove multiple mirror assignments",
			Files: map[string]string{
				"locals.tf": `variable "env" {
	default = "dev"
}
variable "region" {
	default = "us-east-1"
}

locals {
	environment = var.env
	aws_region  = var.region
	computed    = lower(var.env)
}
`,
			},
			Expected: map[string]string{
				"locals.tf": `variable "env" {
  default = "dev"
}
variable "region" {
  default = "us-east-1"
}

locals {
  computed = lower(var.env)
}
`,
			},
		},
		{
			Name: "Autofix - preserve valid expressions",
			Files: map[string]string{
				"locals.tf": `variable "hello" {}

locals {
	hello      = lower(var.hello)
	direct_bad = var.hello
}
`,
			},
			Expected: map[string]string{
				"locals.tf": `variable "hello" {}

locals {
  hello = lower(var.hello)
}
`,
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

			// Check that we have issues and they are fixable
			if len(runner.Issues) == 0 {
				t.Fatal("Expected issues to be found, but none were found")
			}

			// Apply autofixes by triggering the fix functions
			// The helper runner should automatically apply fixes when EmitIssueWithFix is called
			changes := runner.Changes()
			
			// Use AssertChanges to verify the fixes were applied correctly
			helper.AssertChanges(t, tc.Expected, changes)
		})
	}
}
