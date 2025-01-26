package rules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func TestTerraformRepeatedValuesRule_DefaultThreshold(t *testing.T) {
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
					Rule:    NewTerraformRepeatedValuesRule(),
					Message: `Value "myvalue" repeated 3 times. Consider a local variable.`,
					Range: hcl.Range{
						Filename: "main.tf",
						Start:    hcl.Pos{Line: 3, Column: 2},
						End:      hcl.Pos{Line: 3, Column: 19},
					},
				},
				{
					Rule:    NewTerraformRepeatedValuesRule(),
					Message: `Value "myvalue" repeated 3 times. Consider a local variable.`,
					Range: hcl.Range{
						Filename: "main.tf",
						Start:    hcl.Pos{Line: 4, Column: 2},
						End:      hcl.Pos{Line: 4, Column: 19},
					},
				},
				{
					Rule:    NewTerraformRepeatedValuesRule(),
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

	rule := NewTerraformRepeatedValuesRule()

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

func TestTerraformRepeatedValuesRule_ConfigThreshold(t *testing.T) {
	// We'll run multiple sub-tests with different thresholds

	t.Run("Repeated 2 times with threshold=2", func(t *testing.T) {
		content := `
resource "fake_resource" "example" {
  name  = "repeated"
  other = "repeated"
}
`
		// That is repeated 2 times
		// => we expect issues if threshold=2

		rule := NewTerraformRepeatedValuesRule()

		// Provide a .tflint.hcl file so the rule can read threshold=2 from it
		runner := helper.TestRunner(t, map[string]string{
			".tflint.hcl": `
rule "terraform_repeated_values" {
  enabled = true
  threshold = 2
}
`,
			"main.tf": content,
		})

		if err := rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}

		helper.AssertIssues(t, helper.Issues{
			{
				Rule:    rule,
				Message: `Value "repeated" repeated 2 times. Consider a local variable.`,
				Range: hcl.Range{
					Filename: "main.tf",
					Start:    hcl.Pos{Line: 3, Column: 3},
					End:      hcl.Pos{Line: 3, Column: 21},
				},
			},
			{
				Rule:    rule,
				Message: `Value "repeated" repeated 2 times. Consider a local variable.`,
				Range: hcl.Range{
					Filename: "main.tf",
					Start:    hcl.Pos{Line: 4, Column: 3},
					End:      hcl.Pos{Line: 4, Column: 21},
				},
			},
		}, runner.Issues)
	})

	t.Run("3 repeats with threshold=4 => no issues", func(t *testing.T) {
		content := `
resource "fake_resource" "one" {
  value = "hello"
}

resource "fake_resource" "two" {
  value = "hello"
}

resource "fake_resource" "three" {
  value = "hello"
}
`
		rule := NewTerraformRepeatedValuesRule()
		runner := helper.TestRunner(t, map[string]string{
			".tflint.hcl": `
rule "terraform_repeated_values" {
  enabled = true
  threshold = 4
}
`,
			"main.tf": content,
		})

		if err := rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		// We expect zero issues because "hello" repeats only 3 times
		helper.AssertIssues(t, helper.Issues{}, runner.Issues)
	})

	t.Run("1 repeat with threshold=1 => single usage is flagged", func(t *testing.T) {
		content := `
resource "fake_resource" "example" {
  name = "solo"
}
`
		rule := NewTerraformRepeatedValuesRule()
		runner := helper.TestRunner(t, map[string]string{
			".tflint.hcl": `
rule "terraform_repeated_values" {
  enabled = true
  threshold = 1
}
`,
			"main.tf": content,
		})

		if err := rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		// "solo" is used once, which meets (>=1) => 1 issue
		helper.AssertIssues(t, helper.Issues{
			{
				Rule:    rule,
				Message: `Value "solo" repeated 1 times. Consider a local variable.`,
				Range: hcl.Range{
					Filename: "main.tf",
					Start:    hcl.Pos{Line: 3, Column: 3},
					End:      hcl.Pos{Line: 3, Column: 16},
				},
			},
		}, runner.Issues)
	})
}
