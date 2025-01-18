package rules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func TestTerraformResourceNameRule(t *testing.T) {
	tests := []struct {
		Name     string
		Content  string
		Expected helper.Issues
	}{
		{
			Name: "valid name (does not repeat type)",
			Content: `
resource "azurerm_load_balancer" "lb" {
	// ...
}
`,
			Expected: helper.Issues{},
		},
		{
			Name: "invalid repeated name",
			Content: `
resource "azurerm_load_balancer" "load_balancer" {
	// ...
}
`,
			Expected: helper.Issues{
				{
					Rule:    NewTerraformResourceNameRule(),
					Message: "Resource name repeats resource type 'load_balancer'",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 55},
					},
				},
			},
		},
	}

	rule := NewTerraformResourceNameRule()

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			runner := helper.TestRunner(t, map[string]string{"resource.tf": test.Content})

			if err := rule.Check(runner); err != nil {
				t.Fatalf("Unexpected error occurred: %s", err)
			}

			helper.AssertIssues(t, test.Expected, runner.Issues)
		})
	}
}
