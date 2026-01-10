package rules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func TestTerraformProviderFileRule(t *testing.T) {
	awsProvider := readFixture(t, "provider_file_aws_provider.tf")
	azurermProvider := readFixture(t, "provider_file_azurerm_provider.tf")
	terraformWithProvider := readFixture(t, "provider_file_terraform_with_provider.tf")
	simpleResource := readFixture(t, "simple_resource.tf")

	tests := []struct {
		Name     string
		FileMap  map[string]string
		Expected helper.Issues
	}{
		{
			Name: "Valid - provider in providers.tf",
			FileMap: map[string]string{
				"providers.tf": awsProvider,
			},
			Expected: helper.Issues{},
		},
		{
			Name: "Valid - provider in providers.prod.tf",
			FileMap: map[string]string{
				"providers.prod.tf": awsProvider,
			},
			Expected: helper.Issues{},
		},
		{
			Name: "Valid - provider in providers.azure.tf",
			FileMap: map[string]string{
				"providers.azure.tf": azurermProvider,
			},
			Expected: helper.Issues{},
		},
		{
			Name: "Invalid - provider in main.tf",
			FileMap: map[string]string{
				"main.tf": awsProvider,
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
				"terraform.tf": terraformWithProvider,
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
				"main.tf": simpleResource,
			},
			Expected: helper.Issues{},
		},
		{
			Name: "Multiple providers - one valid, one invalid",
			FileMap: map[string]string{
				"providers.tf": awsProvider,
				"main.tf":      azurermProvider,
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

func TestTerraformProviderFileRule_MultipleProvidersInWrongFile(t *testing.T) {
	// Test that multiple provider blocks in wrong file are all flagged
	rule := NewTerraformProviderFileRule()
	runner := helper.TestRunner(t, map[string]string{
		"main.tf": readFixture(t, "provider_file_multiple_wrong.tf"),
	})

	if err := rule.Check(runner); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(runner.Issues) != 3 {
		t.Fatalf("Expected 3 issues, got %d", len(runner.Issues))
	}
}

func TestTerraformProviderFileRule_AliasedProvider(t *testing.T) {
	// Test that aliased providers are also caught
	rule := NewTerraformProviderFileRule()
	runner := helper.TestRunner(t, map[string]string{
		"main.tf": readFixture(t, "provider_file_aliased_wrong.tf"),
	})

	if err := rule.Check(runner); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(runner.Issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(runner.Issues))
	}
}

func TestTerraformProviderFileRule_ValidAliasedProvider(t *testing.T) {
	// Test that aliased providers in correct file pass
	rule := NewTerraformProviderFileRule()
	runner := helper.TestRunner(t, map[string]string{
		"providers.tf": readFixture(t, "provider_file_aliased_valid.tf"),
	})

	if err := rule.Check(runner); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	helper.AssertIssues(t, helper.Issues{}, runner.Issues)
}

func TestTerraformProviderFileRule_ProviderInSubdirectory(t *testing.T) {
	// Test that providers in subdirectory wrong files are caught
	rule := NewTerraformProviderFileRule()
	runner := helper.TestRunner(t, map[string]string{
		"modules/submodule/main.tf": readFixture(t, "provider_file_simple_valid.tf"),
	})

	if err := rule.Check(runner); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(runner.Issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(runner.Issues))
	}
}

func TestTerraformProviderFileRule_ProviderInSubdirectoryValid(t *testing.T) {
	// Test that providers in subdirectory providers.tf pass
	rule := NewTerraformProviderFileRule()
	runner := helper.TestRunner(t, map[string]string{
		"modules/submodule/providers.tf": readFixture(t, "provider_file_simple_valid.tf"),
	})

	if err := rule.Check(runner); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	helper.AssertIssues(t, helper.Issues{}, runner.Issues)
}

func TestTerraformProviderFileRule_ProviderWithComplexConfig(t *testing.T) {
	// Test provider with complex configuration
	rule := NewTerraformProviderFileRule()
	runner := helper.TestRunner(t, map[string]string{
		"providers.tf": readFixture(t, "provider_file_complex_valid.tf"),
	})

	if err := rule.Check(runner); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	helper.AssertIssues(t, helper.Issues{}, runner.Issues)
}
