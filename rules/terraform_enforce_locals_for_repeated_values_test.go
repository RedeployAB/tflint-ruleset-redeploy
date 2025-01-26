package rules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func TestTerraformEnforceLocalsForRepeatedValuesRule_DefaultThreshold(t *testing.T) {
	tests := []struct {
		Name    string
		Content string
		Issues  helper.Issues
	}{
		{
			Name: "OK - repeated only 2 times, threshold=3",
			Content: `
resource "fake_resource" "example" {
	name  = "myvalue"
	other = "something"
}

resource "another_resource" "stuff" {
	name = "myvalue"
}
`,
			Issues: helper.Issues{}, // only repeated 2 times => OK
		},
		{
			Name: "NOT OK - repeated 3 times, threshold=3",
			Content: `
resource "fake_resource" "example" {
	name  = "myvalue"
	other = "myvalue"
}

resource "another_resource" "stuff" {
	name = "myvalue"
}
`,
			Issues: helper.Issues{
				// We expect issues for all 3 occurrences
				{
					Rule:    NewTerraformEnforceLocalsForRepeatedValuesRule(),
					Message: `Value "myvalue" repeated 3 times. Consider a local variable.`,
					Range: hcl.Range{
						Filename: "main.tf",
						Start:    hcl.Pos{Line: 3, Column: 2},
						End:      hcl.Pos{Line: 3, Column: 19},
					},
				},
				{
					Rule:    NewTerraformEnforceLocalsForRepeatedValuesRule(),
					Message: `Value "myvalue" repeated 3 times. Consider a local variable.`,
					Range: hcl.Range{
						Filename: "main.tf",
						Start:    hcl.Pos{Line: 4, Column: 2},
						End:      hcl.Pos{Line: 4, Column: 19},
					},
				},
				{
					Rule:    NewTerraformEnforceLocalsForRepeatedValuesRule(),
					Message: `Value "myvalue" repeated 3 times. Consider a local variable.`,
					Range: hcl.Range{
						Filename: "main.tf",
						Start:    hcl.Pos{Line: 8, Column: 2},
						End:      hcl.Pos{Line: 8, Column: 18},
					},
				},
			},
		},
	}

	rule := NewTerraformEnforceLocalsForRepeatedValuesRule()

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runner := helper.TestRunner(t, map[string]string{
				"main.tf": tc.Content,
			})

			if err := rule.Check(runner); err != nil {
				t.Fatalf("Unexpected error: %s", err)
			}

			helper.AssertIssues(t, tc.Issues, runner.Issues)
		})
	}
}

func TestTerraformEnforceLocalsForRepeatedValuesRule_ConfigThreshold(t *testing.T) {
	// We'll configure threshold=2
	content := `
resource "fake_resource" "example" {
	name  = "repeated"
	other = "repeated"
}
`
	// That is repeated 2 times
	// => we expect issues if threshold=2

	rule := NewTerraformEnforceLocalsForRepeatedValuesRule()

	// Provide a .tflint.hcl file so the rule can read threshold=2 from it
	runner := helper.TestRunner(t, map[string]string{
		".tflint.hcl": `
rule "terraform_enforce_locals_for_repeated_values" {
	enabled = true
	threshold = 2
}
`,
		"main.tf": content,
	})

	// Make sure we actually load the .tflint.hcl
	if err := runner.LoadConfig(); err != nil {
		t.Fatalf("Unexpected error loading config: %s", err)
	}
	// Let the rule see the loaded config
	if err := rule.Configure(runner); err != nil {
		t.Fatalf("Unexpected error configuring rule: %s", err)
	}

	if err := rule.Check(runner); err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	helper.AssertIssues(t, helper.Issues{
		{
			Rule:    rule,
			Message: `Value "repeated" repeated 2 times. Consider a local variable.`,
			Range: hcl.Range{
				Filename: "main.tf",
				Start:    hcl.Pos{Line: 3, Column: 10},
				End:      hcl.Pos{Line: 3, Column: 20},
			},
		},
		{
			Rule:    rule,
			Message: `Value "repeated" repeated 2 times. Consider a local variable.`,
			Range: hcl.Range{
				Filename: "main.tf",
				Start:    hcl.Pos{Line: 4, Column: 10},
				End:      hcl.Pos{Line: 4, Column: 20},
			},
		},
	}, runner.Issues)
}
