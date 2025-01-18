package rules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func TestTerraformMetaArguments(t *testing.T) {
	tests := []struct {
		Name     string
		Content  string
		Expected helper.Issues
	}{
		{
			Name: "valid resource order",
			Content: `
resource "aws_instance" "example" {
  count     = 1
  provider  = aws

  lifecycle {}

  depends_on = []
}`,
			Expected: helper.Issues{},
		},
		{
			Name: "invalid resource order",
			Content: `
resource "aws_instance" "example" {
  depends_on = []
  count      = 1
}`,
			Expected: helper.Issues{
				{
					Rule:    NewTerraformMetaArgumentsRule(),
					Message: "Missing or out-of-order meta arguments in resource 'aws_instance example'. Expected sequence: count|for_each -> provider -> lifecycle -> depends_on",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 34},
					},
				},
			},
		},
		{
			Name: "valid module order",
			Content: `
module "example" {
  count = 2

  depends_on = []
}`,
			Expected: helper.Issues{},
		},
		{
			Name: "invalid module order",
			Content: `
module "example" {
  depends_on = []
  count      = 2
}`,
			Expected: helper.Issues{
				{
					Rule:    NewTerraformMetaArgumentsRule(),
					Message: "Missing or out-of-order meta arguments in module 'example'. Expected sequence: count|for_each -> depends_on",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 17},
					},
				},
			},
		},
	}

	rule := NewTerraformMetaArgumentsRule()

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runner := helper.TestRunner(t, map[string]string{
				"resource.tf": tc.Content,
			})

			if err := rule.Check(runner); err != nil {
				t.Fatalf("Unexpected error occurred: %s", err)
			}

			helper.AssertIssues(t, tc.Expected, runner.Issues)
		})
	}
}
