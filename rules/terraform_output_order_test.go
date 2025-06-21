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
		Name     string
		Content  string
		Expected string
	}{
		{
			Name: "Autofix - simple out of order",
			Content: `output "zzz" {}
output "alpha" {}
`,
			Expected: `output "alpha" {}
output "zzz" {}
`,
		},
		{
			Name: "Autofix - multiple outputs out of order",
			Content: `output "charlie" {
	value = "c"
}
output "beta" {
	value = "b"
}
output "alpha" {
	value = "a"
}
`,
			Expected: `output "alpha" {
  value = "a"
}
output "beta" {
  value = "b"
}
output "charlie" {
  value = "c"
}
`,
		},
		{
			Name: "Autofix - preserve spacing between outputs",
			Content: `output "beta" {}


output "alpha" {}
`,
			Expected: `output "alpha" {}


output "beta" {}
`,
		},
		{
			Name: "Autofix - preserve single line spacing",
			Content: `output "beta" {}
output "alpha" {}
`,
			Expected: `output "alpha" {}
output "beta" {}
`,
		},
		{
			Name: "Autofix - complex mix with different spacing",
			Content: `output "delta" {
	value = "d"
}

output "beta" {}

output "echo" {
	value = "e"
}

output "alpha" {}

output "charlie" {
	value = "c"
}
`,
			Expected: `output "alpha" {}

output "beta" {}

output "charlie" {
  value = "c"
}

output "delta" {
  value = "d"
}

output "echo" {
  value = "e"
}
`,
		},
		{
			Name: "Autofix - outputs with complex values",
			Content: `output "second" {
	value = {
		a = 1
		b = 2
	}
}

output "first" {
	value = [
		"one",
		"two"
	]
}
`,
			Expected: `output "first" {
  value = [
    "one",
    "two"
  ]
}

output "second" {
  value = {
    a = 1
    b = 2
  }
}
`,
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

			helper.AssertChanges(t, map[string]string{
				"test.tf": tc.Expected,
			}, runner.Changes())
		})
	}
}
