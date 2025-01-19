package rules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func TestTerraformSourceFormat(t *testing.T) {
	tests := []struct {
		Name    string
		Content string
		Issues  helper.Issues
	}{
		{
			Name: "OK - only source, block ends",
			Content: `
module "example" {
  source = "a source address"
}
`,
			Issues: helper.Issues{},
		},
		{
			Name: "OK - source plus version, block ends",
			Content: `
module "example" {
  source  = "a source address"
  version = "x.x.x"
}
`,
			Issues: helper.Issues{},
		},
		{
			Name: "NOT OK - source plus version, extra blank line at end is disallowed",
			Content: `
module "example" {
  source  = "a source address"
  version = "x.x.x"

}
`,
			// The rule as described suggests we do NOT want that trailing blank line
			// However, your examples showed that a blank line after version is only needed
			// if more arguments follow. But here, the block ends. So let's produce an issue:
			Issues: helper.Issues{
				{
					Rule:    NewTerraformSourceFormatRule(),
					Message: "Expected a blank line after 'version'",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 5, Column: 1},
						End:      hcl.Pos{Line: 5, Column: 1},
					},
				},
			},
		},
		{
			Name: "NOT OK - source alone with trailing blank line before closing brace",
			Content: `
module "example" {
  source = "a source address"

}
`,
			Issues: helper.Issues{
				{
					Rule:    NewTerraformSourceFormatRule(),
					Message: "Expected a blank line after 'source'",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 4, Column: 1},
						End:      hcl.Pos{Line: 4, Column: 1},
					},
				},
			},
		},
		{
			Name: "OK - source plus version, more property after blank line",
			Content: `
module "example" {
  source  = "a source address"
  version = "x.x.x"

  property = "value"
}
`,
			Issues: helper.Issues{},
		},
		{
			Name: "OK - source alone, more property after blank line",
			Content: `
module "example" {
  source = "a source address"

  property = "value"
}
`,
			Issues: helper.Issues{},
		},
		{
			Name: "NOT OK - source plus version, property follows with no blank line",
			Content: `
module "example" {
  source  = "a source address"
  version = "x.x.x"
  property = "value"
}
`,
			Issues: helper.Issues{
				{
					Rule:    NewTerraformSourceFormatRule(),
					Message: "Expected a blank line after 'version'",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 4, Column: 1},
						End:      hcl.Pos{Line: 4, Column: 1},
					},
				},
			},
		},
	}

	rule := NewTerraformSourceFormatRule()

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
