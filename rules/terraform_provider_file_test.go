package rules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func TestTerraformProviderFileRule(t *testing.T) {
	tests := []struct {
		Name     string
		FileMap  map[string]string
		Expected helper.Issues
	}{
		{
			Name: "Valid - provider in providers.tf",
			FileMap: map[string]string{
				"providers.tf": `provider "aws" {
  region = "us-east-1"
}`,
			},
			Expected: helper.Issues{},
		},
		{
			Name: "Valid - provider in providers.prod.tf",
			FileMap: map[string]string{
				"providers.prod.tf": `provider "aws" {
  region = "us-east-1"
}`,
			},
			Expected: helper.Issues{},
		},
		{
			Name: "Valid - provider in providers.azure.tf",
			FileMap: map[string]string{
				"providers.azure.tf": `provider "azurerm" {
  features {}
}`,
			},
			Expected: helper.Issues{},
		},
		{
			Name: "Invalid - provider in main.tf",
			FileMap: map[string]string{
				"main.tf": `provider "aws" {
  region = "us-east-1"
}`,
			},
			Expected: helper.Issues{
				{
					Rule:    NewTerraformProviderFileRule(),
					Message: `"provider" block must be placed in "providers.tf" or "providers.<area>.tf", not "main.tf"`,
					Range: hcl.Range{
						Filename: "main.tf",
						Start:    hcl.Pos{Line: 1, Column: 1},
						End:      hcl.Pos{Line: 1, Column: 15},
					},
				},
			},
		},
		{
			Name: "Invalid - provider in terraform.tf",
			FileMap: map[string]string{
				"terraform.tf": `terraform {
  required_version = ">= 1.0"
}

provider "aws" {
  region = "us-east-1"
}`,
			},
			Expected: helper.Issues{
				{
					Rule:    NewTerraformProviderFileRule(),
					Message: `"provider" block must be placed in "providers.tf" or "providers.<area>.tf", not "terraform.tf"`,
					Range: hcl.Range{
						Filename: "terraform.tf",
						Start:    hcl.Pos{Line: 5, Column: 1},
						End:      hcl.Pos{Line: 5, Column: 15},
					},
				},
			},
		},
		{
			Name: "Valid - no provider blocks",
			FileMap: map[string]string{
				"main.tf": `resource "aws_instance" "example" {
  ami           = "ami-12345678"
  instance_type = "t2.micro"
}`,
			},
			Expected: helper.Issues{},
		},
		{
			Name: "Multiple providers - one valid, one invalid",
			FileMap: map[string]string{
				"providers.tf": `provider "aws" {
  region = "us-east-1"
}`,
				"main.tf": `provider "azurerm" {
  features {}
}`,
			},
			Expected: helper.Issues{
				{
					Rule:    NewTerraformProviderFileRule(),
					Message: `"provider" block must be placed in "providers.tf" or "providers.<area>.tf", not "main.tf"`,
					Range: hcl.Range{
						Filename: "main.tf",
						Start:    hcl.Pos{Line: 1, Column: 1},
						End:      hcl.Pos{Line: 1, Column: 19},
					},
				},
			},
		},
	}

	rule := NewTerraformProviderFileRule()
	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runner := helper.TestRunner(t, tc.FileMap)
			if err := rule.Check(runner); err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			helper.AssertIssues(t, tc.Expected, runner.Issues)
		})
	}
}
