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
						Start:    hcl.Pos{Line: 5, Column: 2},
						End:      hcl.Pos{Line: 5, Column: 30},
					},
				},
			},
		},