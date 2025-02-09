package rules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func TestTerraformBlockOrderRule(t *testing.T) {
	cases := []struct {
		Name    string
		Content string
		Issues  helper.Issues
	}{
		{
			Name: "OK - all blocks in correct order",
			Content: `
terraform {
	required_version = ">= 1.0.0"
}

provider "aws" {}

data "aws_iam_user" "example" {}

resource "aws_instance" "example" {}
`,
			Issues: helper.Issues{},
		},
		{
			Name: "OK - skipping some blocks",
			Content: `
provider "aws" {}

resource "aws_instance" "example" {}
`,
			Issues: helper.Issues{},
		},
		{
			Name: "NOT OK - 'provider' after 'resource'",
			Content: `
terraform {}

resource "aws_instance" "example" {}

provider "aws" {}
`,
			Issues: helper.Issues{
				{
					Rule:    NewTerraformBlockOrderRule(),
					Message: "Out-of-order block 'provider'. Expected order: terraform -> provider -> data -> resource",
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 6, Column: 1},
						End:      hcl.Pos{Line: 6, Column: 15},
					},
				},
			},
		},
	}

	rule := NewTerraformBlockOrderRule()
	for _, tc := range cases {
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
