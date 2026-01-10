package rules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func TestTerraformRequiredProvidersOrderRule(t *testing.T) {
	tests := []struct {
		Name    string
		Content string
		Issues  helper.Issues
	}{
		{
			Name:    "Valid alphabetical order",
			Content: readFixture(t, "required_providers_order_valid.tf"),
			Issues:  helper.Issues{},
		},
		{
			Name:    "Invalid order (random before aws)",
			Content: readFixture(t, "required_providers_order_invalid.tf"),
			Issues: helper.Issues{
				{
					Rule:    NewTerraformRequiredProvidersOrderRule(),
					Message: "Provider 'random' is out of alphabetical order. Expected order: aws, azurerm, random",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 3, Column: 5},
						End:      hcl.Pos{Line: 6, Column: 6},
					},
				},
			},
		},
		{
			Name:    "Single provider (no ordering needed)",
			Content: readFixture(t, "required_providers_order_single.tf"),
			Issues:  helper.Issues{},
		},
		{
			Name:    "Empty required_providers block",
			Content: readFixture(t, "required_providers_order_empty.tf"),
			Issues:  helper.Issues{},
		},
	}

	rule := NewTerraformRequiredProvidersOrderRule()

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runner := helper.TestRunner(t, map[string]string{
				"resource.tf": tc.Content,
			})
			if err := rule.Check(runner); err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			helper.AssertIssues(t, tc.Issues, runner.Issues)
		})
	}
}

func TestTerraformRequiredProvidersOrderRule_CaseInsensitive(t *testing.T) {
	// Test that ordering is case-insensitive
	content := `terraform {
  required_providers {
    AWS = {
      source  = "hashicorp/aws"
      version = "~> 4.0"
    }
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 3.0"
    }
  }
}`

	rule := NewTerraformRequiredProvidersOrderRule()
	runner := helper.TestRunner(t, map[string]string{
		"resource.tf": content,
	})

	if err := rule.Check(runner); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// AWS (uppercase) should come before azurerm (case-insensitive)
	helper.AssertIssues(t, helper.Issues{}, runner.Issues)
}

func TestTerraformRequiredProvidersOrderRule_NoRequiredProviders(t *testing.T) {
	// Test with terraform block but no required_providers
	content := `terraform {
  required_version = ">= 1.0"
}`

	rule := NewTerraformRequiredProvidersOrderRule()
	runner := helper.TestRunner(t, map[string]string{
		"resource.tf": content,
	})

	if err := rule.Check(runner); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	helper.AssertIssues(t, helper.Issues{}, runner.Issues)
}

func TestTerraformRequiredProvidersOrderRule_NoTerraformBlock(t *testing.T) {
	// Test with no terraform block at all
	content := `resource "aws_instance" "example" {
  ami           = "ami-12345678"
  instance_type = "t2.micro"
}`

	rule := NewTerraformRequiredProvidersOrderRule()
	runner := helper.TestRunner(t, map[string]string{
		"resource.tf": content,
	})

	if err := rule.Check(runner); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	helper.AssertIssues(t, helper.Issues{}, runner.Issues)
}

func TestTerraformRequiredProvidersOrderRule_Autofix(t *testing.T) {
	tests := []struct {
		Name         string
		ContentFile  string
		ExpectedFile string
	}{
		{
			Name:         "Autofix - reorder providers alphabetically",
			ContentFile:  "required_providers_order_invalid.tf",
			ExpectedFile: "required_providers_order_invalid_expected.tf",
		},
	}

	rule := NewTerraformRequiredProvidersOrderRule()
	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			content := readFixture(t, tc.ContentFile)
			expected := readFixture(t, tc.ExpectedFile)

			runner := helper.TestRunner(t, map[string]string{
				"resource.tf": content,
			})

			if err := rule.Check(runner); err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			helper.AssertChanges(t, map[string]string{
				"resource.tf": expected,
			}, runner.Changes())
		})
	}
}
