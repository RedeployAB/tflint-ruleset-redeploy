package rules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func TestTerraformBlockFormat(t *testing.T) {
	tests := []struct {
		Name    string
		Content string
		Issues  helper.Issues
	}{
		// Existing tests for resource
		{
			Name: "OK - attribute then block with blank line",
			Content: `
resource "azurerm_resource_provider_registration" "example" {
  name = "Microsoft.ContainerService"

  feature {
    name       = "AKS-DataPlaneAutoApprove"
    registered = true
  }
}
`,
			Issues: helper.Issues{},
		},
		{
			Name: "NOT OK - attribute then block with no blank line",
			Content: `
resource "azurerm_resource_provider_registration" "example" {
  name = "Microsoft.ContainerService"
  feature {
    name       = "AKS-DataPlaneAutoApprove"
    registered = true
  }
}
`,
			Issues: helper.Issues{
				{
					Rule:    NewTerraformBlockFormatRule(),
					Message: "Expected exactly one blank line before this block",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 4, Column: 3},
						End:      hcl.Pos{Line: 4, Column: 10},
					},
				},
			},
		},
		{
			Name: "OK - single block first (no attribute), no blank line after brace",
			Content: `
resource "azurerm_resource_provider_registration" "example" {
  feature {
    name       = "AKS-DataPlaneAutoApprove"
    registered = true
  }
}
`,
			Issues: helper.Issues{},
		},
		{
			Name: "NOT OK - single block first with extra blank line after brace",
			Content: `
resource "azurerm_resource_provider_registration" "example" {

  feature {
    name       = "AKS-DataPlaneAutoApprove"
    registered = true
  }
}
`,
			Issues: helper.Issues{
				{
					Rule:    NewTerraformBlockFormatRule(),
					Message: "Block should appear immediately after opening brace when it's the first item (no blank lines)",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 4, Column: 3},
						End:      hcl.Pos{Line: 4, Column: 10},
					},
				},
			},
		},
		{
			Name: "OK - multiple blocks each separated by single blank line",
			Content: `
resource "azurerm_firewall_policy_rule_collection_group" "example" {
  application_rule_collection {
    name  = "app_rule_collection1"
    action = "Deny"
  }

  network_rule_collection {
    name  = "network_rule_collection1"
    action = "Deny"
  }

  nat_rule_collection {
    name = "nat_rule_collection1"
    action = "Dnat"
  }
}
`,
			Issues: helper.Issues{},
		},
		{
			Name: "NOT OK - multiple blocks no blank line between them",
			Content: `
resource "azurerm_firewall_policy_rule_collection_group" "example" {
  application_rule_collection {
    name  = "app_rule_collection1"
    action = "Deny"
  }
  network_rule_collection {
    name  = "network_rule_collection1"
    action = "Deny"
  }
  nat_rule_collection {
    name = "nat_rule_collection1"
    action = "Dnat"
  }
}
`,
			Issues: helper.Issues{
				{
					Rule:    NewTerraformBlockFormatRule(),
					Message: "Expected exactly one blank line before this block",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 7, Column: 3},
						End:      hcl.Pos{Line: 7, Column: 26},
					},
				},
				{
					Rule:    NewTerraformBlockFormatRule(),
					Message: "Expected exactly one blank line before this block",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 11, Column: 3},
						End:      hcl.Pos{Line: 11, Column: 20},
					},
				},
			},
		},
		{
			Name: "OK - attributes then multiple blocks each separated by single blank line",
			Content: `
resource "azurerm_firewall_policy_rule_collection_group" "example" {
  name = "test"
  priority = 123

  application_rule_collection {
    name  = "app_rule_collection1"
    action = "Deny"
  }

  network_rule_collection {
    name  = "network_rule_collection1"
    action = "Deny"
  }
}
`,
			Issues: helper.Issues{},
		},
		{
			Name: "NOT OK - attribute then block with no blank line, then next block also with no blank line",
			Content: `
resource "azurerm_firewall_policy_rule_collection_group" "example" {
  name = "test"
  application_rule_collection {
    name = "app_rule_collection1"
    action = "Deny"
  }
  network_rule_collection {
    name = "network_rule_collection1"
    action = "Deny"
  }
}
`,
			Issues: helper.Issues{
				{
					Rule:    NewTerraformBlockFormatRule(),
					Message: "Expected exactly one blank line before this block",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 4, Column: 3},
						End:      hcl.Pos{Line: 4, Column: 30},
					},
				},
				{
					Rule:    NewTerraformBlockFormatRule(),
					Message: "Expected exactly one blank line before this block",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 8, Column: 3},
						End:      hcl.Pos{Line: 8, Column: 24},
					},
				},
			},
		},
		// Additional test cases from the removed function
		{
			Name: "OK - single data block with first sub-block no blank line",
			Content: `
data "aws_ami" "example" {
  filter {
    name   = "xyz"
    values = ["abc"]
  }
}
`,
			Issues: helper.Issues{},
		},
		{
			Name: "NOT OK - single data block with extra blank line after brace",
			Content: `
data "aws_ami" "example" {

  filter {
    name   = "xyz"
    values = ["abc"]
  }
}
`,
			Issues: helper.Issues{
				{
					Rule:    NewTerraformBlockFormatRule(),
					Message: "Block should appear immediately after opening brace when it's the first item (no blank lines)",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 4, Column: 3},
						End:      hcl.Pos{Line: 4, Column: 8},
					},
				},
			},
		},
		{
			Name: "OK - terraform block with two sub-blocks, each separated by one blank line",
			Content: `
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 3.0"
    }
  }

  experiments {}
}
`,
			Issues: helper.Issues{},
		},
		{
			Name: "NOT OK - provider block with no blank line between sub-blocks",
			Content: `
provider "aws" {
  alias = "one"
  assume_role {}
  default_tags {}
}
`,
			Issues: helper.Issues{
				{
					Rule:    NewTerraformBlockFormatRule(),
					Message: "Expected exactly one blank line before this block",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 4, Column: 3},
						End:      hcl.Pos{Line: 4, Column: 14},
					},
				},
			},
		},
	}

	rule := NewTerraformBlockFormatRule()

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
