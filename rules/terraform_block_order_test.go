package rules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func TestTerraformBlockOrderRule(t *testing.T) {
	tests := []struct {
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
		{
			Name: "OK - moved block can appear anywhere",
			Content: `
terraform {}

moved {
  from = aws_instance.old
  to   = aws_instance.new
}

resource "aws_instance" "new" {}
`,
			Issues: helper.Issues{},
		},
		{
			Name: "OK - import block can appear anywhere",
			Content: `
terraform {}

resource "aws_instance" "example" {}

import {
  to = aws_instance.example
  id = "i-1234567890abcdef0"
}
`,
			Issues: helper.Issues{},
		},
		{
			Name: "OK - removed block can appear anywhere",
			Content: `
terraform {}

removed {
  from = aws_instance.old

  lifecycle {
    destroy = false
  }
}

resource "aws_instance" "new" {}
`,
			Issues: helper.Issues{},
		},
		{
			Name: "OK - check block can appear anywhere",
			Content: `
terraform {}

check "health_check" {
  data "http" "example" {
    url = "https://example.com"
  }

  assert {
    condition     = data.http.example.status_code == 200
    error_message = "Health check failed"
  }
}

resource "aws_instance" "example" {}
`,
			Issues: helper.Issues{},
		},
		{
			Name: "OK - mixed lifecycle blocks with ordered blocks",
			Content: `
terraform {}

provider "aws" {}

moved {
  from = aws_instance.old
  to   = aws_instance.renamed
}

data "aws_ami" "example" {}

import {
  to = aws_instance.imported
  id = "i-1234567890abcdef0"
}

resource "aws_instance" "renamed" {}
resource "aws_instance" "imported" {}

check "health" {
  assert {
    condition     = true
    error_message = "Always passes"
  }
}
`,
			Issues: helper.Issues{},
		},
	}

	rule := NewTerraformBlockOrderRule()
	for _, tc := range tests {
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
