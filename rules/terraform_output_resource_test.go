package rules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func TestTerraformOutputResourceRule(t *testing.T) {
	tests := []struct {
		Name    string
		Content string
		Issues  helper.Issues
	}{
		{
			Name: "OK - references a single attribute",
			Content: `
resource "aws_instance" "example" {}

output "out_ok" {
  value = aws_instance.example.id
}
`,
			Issues: helper.Issues{},
		},
		{
			Name: "NOT OK - references entire resource",
			Content: `
resource "aws_instance" "example" {}

output "out_bad" {
  value = aws_instance.example
}
`,
			Issues: helper.Issues{
				{
					Rule:    NewTerraformOutputResourceRule(),
					Message: "Output is referencing the entire resource or data, rather than a specific attribute. This can cause schema issues.",
					Range: hcl.Range{
						Filename: "main.tf",
						Start:    hcl.Pos{Line: 5, Column: 9},
						End:      hcl.Pos{Line: 5, Column: 34},
					},
				},
			},
		},
		{
			Name: "NOT OK - references entire data resource",
			Content: `
data "aws_caller_identity" "current" {}

output "caller" {
  value = data.aws_caller_identity.current
}
`,
			Issues: helper.Issues{
				{
					Rule:    NewTerraformOutputResourceRule(),
					Message: "Output is referencing the entire resource or data, rather than a specific attribute. This can cause schema issues.",
					Range: hcl.Range{
						Filename: "main.tf",
						Start:    hcl.Pos{Line: 5, Column: 9},
						End:      hcl.Pos{Line: 5, Column: 47},
					},
				},
			},
		},
		{
			Name: "OK - ternary with variable check referencing resource attribute",
			Content: `
variable "aks_identity_type" {}

resource "azurerm_user_assigned_identity" "aks" {
  name = "dummy"
}

output "some_output" {
  value = var.aks_identity_type == "UserAssigned" ? azurerm_user_assigned_identity.aks[0].client_id : null
}
`,
			Issues: helper.Issues{},
		},
	}

	rule := NewTerraformOutputResourceRule()
	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runner := helper.TestRunner(t, map[string]string{
				"main.tf": tc.Content,
			})
			if err := rule.Check(runner); err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			helper.AssertIssues(t, tc.Issues, runner.Issues)
		})
	}
}
