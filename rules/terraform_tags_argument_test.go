package rules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func TestTerraformTagsArgumentRule(t *testing.T) {
	tests := []struct {
		Name    string
		Content string
		Issues  helper.Issues
	}{
		{
			Name: "OK usage with tags last, then depends_on, then lifecycle",
			Content: `
resource "aws_nat_gateway" "this" {
  count = 2

  allocation_id = "..."
  subnet_id     = "..."

  tags = {
    Name = "..."
  }

  depends_on = [aws_internet_gateway.this]

  lifecycle {}
}
`,
			Issues: helper.Issues{},
		},
		{
			Name: "OK usage with only tags",
			Content: `
resource "aws_nat_gateway" "this" {
  tags = {
    Name = "..."
  }
}
`,
			Issues: helper.Issues{},
		},
		{
			Name: "OK usage with tags last, no depends_on or lifecycle",
			Content: `
resource "aws_nat_gateway" "this" {
  allocation_id = "..."
  subnet_id     = "..."

  tags = {
    Name = "..."
  }
}
`,
			Issues: helper.Issues{},
		},
		{
			Name: "NOT OK usage - normal arguments after tags",
			Content: `
resource "aws_nat_gateway" "this" {
  count = 2

  tags = {
    Name = "..."
  }

  allocation_id = "..."
}
`,
			Issues: helper.Issues{
				{
					Rule:    NewTerraformTagsArgumentRule(),
					Message: "Argument 'allocation_id' must not come after 'tags'",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 8, Column: 3},
						End:      hcl.Pos{Line: 8, Column: 27},
					},
				},
			},
		},
		{
			Name: "NOT OK usage - block after tags that's not lifecycle",
			Content: `
resource "aws_nat_gateway" "this" {
  tags = {
    Name = "..."
  }

  something_else {}
}
`,
			Issues: helper.Issues{
				{
					Rule:    NewTerraformTagsArgumentRule(),
					Message: "Block 'something_else' must not come after 'tags'",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 6, Column: 3},
						End:      hcl.Pos{Line: 6, Column: 19},
					},
				},
			},
		},
		{
			Name: "NOT OK usage - missing blank line between tags and depends_on",
			Content: `
resource "aws_nat_gateway" "this" {
  tags = {
    Name = "..."
  }
  depends_on = [aws_internet_gateway.this]
}
`,
			Issues: helper.Issues{
				{
					Rule:    NewTerraformTagsArgumentRule(),
					Message: "Expected exactly one blank line between 'tags' and 'depends_on'",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 6, Column: 3},
						End:      hcl.Pos{Line: 6, Column: 53},
					},
				},
			},
		},
		{
			Name: "NOT OK usage - missing blank line between tags and lifecycle",
			Content: `
resource "aws_nat_gateway" "this" {
  tags = {
    Name = "..."
  }
  lifecycle {}
}
`,
			Issues: helper.Issues{
				{
					Rule:    NewTerraformTagsArgumentRule(),
					Message: "Expected exactly one blank line between 'tags' and 'lifecycle'",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 6, Column: 3},
						End:      hcl.Pos{Line: 6, Column: 14},
					},
				},
			},
		},
	}

	rule := NewTerraformTagsArgumentRule()

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
