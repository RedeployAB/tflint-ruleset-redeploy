package rules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func TestTerraformMetaArgumentFormat(t *testing.T) {
	tests := []struct {
		Name    string
		Content string
		Issues  helper.Issues
	}{
		{
			Name: "resource with no meta arguments",
			Content: `
resource "aws_instance" "example" {
  # No meta arguments at all
}
`,
			Issues: helper.Issues{},
		},
		{
			Name: "module with no meta arguments",
			Content: `
module "example" {
  # No meta arguments here
}
`,
			Issues: helper.Issues{},
		},
		{
			Name: "resource with top meta argument on last line (no blank line required)",
			Content: `
resource "aws_instance" "example" {
  count = 1
}
`,
			Issues: helper.Issues{},
		},
		{
			Name: "resource missing blank line after top meta argument",
			Content: `
resource "aws_instance" "example" {
  provider = aws
  # next line isn't blank
  name = "test"
}
`,
			Issues: helper.Issues{
				{
					Rule:    NewTerraformMetaArgumentFormatRule(),
					Message: "Expected a blank line after meta-arguments (count/for_each/provider)",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 5, Column: 1},
						End:      hcl.Pos{Line: 5, Column: 1},
					},
				},
			},
		},
		{
			Name: "resource missing blank line before bottom meta argument (depends_on)",
			Content: `
resource "aws_instance" "example" {
  tags = {}
  depends_on = []
}
`,
			Issues: helper.Issues{
				{
					Rule:    NewTerraformMetaArgumentFormatRule(),
					Message: `Expected a blank line before meta-argument 'depends_on'`,
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 4, Column: 1},
						End:      hcl.Pos{Line: 4, Column: 1},
					},
				},
			},
		},
		{
			Name: "module valid formatting with top and bottom meta arguments",
			Content: `
module "example" {
  count = 2

  depends_on = []
}
`,
			Issues: helper.Issues{},
		},
		{
			Name: "module missing blank line before bottom meta argument (depends_on)",
			Content: `
module "example" {
  tags = {}
  depends_on = []
}
`,
			Issues: helper.Issues{
				{
					Rule:    NewTerraformMetaArgumentFormatRule(),
					Message: `Expected a blank line before meta-argument 'depends_on'`,
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 4, Column: 1},
						End:      hcl.Pos{Line: 4, Column: 1},
					},
				},
			},
		},
		{
			Name: "resource partial usage (only provider) - valid formatting",
			Content: `
resource "aws_instance" "example" {
  provider = aws

  # Some setting
  name     = "valid"
}
`,
			Issues: helper.Issues{},
		},
		{
			Name: "resource partial usage (only lifecycle) - missing blank line before",
			Content: `
resource "aws_instance" "example" {
  tags = {
    Something = "xyz"
  }
  lifecycle {}
}
`,
			Issues: helper.Issues{
				{
					Rule:    NewTerraformMetaArgumentFormatRule(),
					Message: `Expected a blank line before meta-argument 'lifecycle'`,
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 6, Column: 1},
						End:      hcl.Pos{Line: 6, Column: 1},
					},
				},
			},
		},
	}

	rule := NewTerraformMetaArgumentFormatRule()

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runner := helper.TestRunner(t, map[string]string{
				"resource.tf": tc.Content,
			})

			if err := rule.Check(runner); err != nil {
				t.Fatalf("Unexpected error occurred: %s", err)
			}

			helper.AssertIssues(t, tc.Issues, runner.Issues)
		})
	}
}
