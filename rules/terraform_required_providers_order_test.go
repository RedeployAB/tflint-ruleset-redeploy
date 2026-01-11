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
	rule := NewTerraformRequiredProvidersOrderRule()
	runner := helper.TestRunner(t, map[string]string{
		"resource.tf": readFixture(t, "required_providers_order_case_insensitive.tf"),
	})

	if err := rule.Check(runner); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// AWS (uppercase) should come before azurerm (case-insensitive)
	helper.AssertIssues(t, helper.Issues{}, runner.Issues)
}

func TestTerraformRequiredProvidersOrderRule_NoRequiredProviders(t *testing.T) {
	// Test with terraform block but no required_providers
	rule := NewTerraformRequiredProvidersOrderRule()
	runner := helper.TestRunner(t, map[string]string{
		"resource.tf": readFixture(t, "required_providers_order_no_required_providers.tf"),
	})

	if err := rule.Check(runner); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	helper.AssertIssues(t, helper.Issues{}, runner.Issues)
}

func TestTerraformRequiredProvidersOrderRule_NoTerraformBlock(t *testing.T) {
	// Test with no terraform block at all
	rule := NewTerraformRequiredProvidersOrderRule()
	runner := helper.TestRunner(t, map[string]string{
		"resource.tf": readFixture(t, "required_providers_order_no_terraform_block.tf"),
	})

	if err := rule.Check(runner); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	helper.AssertIssues(t, helper.Issues{}, runner.Issues)
}

func TestTerraformRequiredProvidersOrderRule_MultipleFiles(t *testing.T) {
	// Test with terraform blocks in multiple files
	rule := NewTerraformRequiredProvidersOrderRule()
	runner := helper.TestRunner(t, map[string]string{
		"terraform.tf": readFixture(t, "required_providers_order_multiple_files_valid.tf"),
		"main.tf":      readFixture(t, "simple_resource.tf"),
	})

	if err := rule.Check(runner); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	helper.AssertIssues(t, helper.Issues{}, runner.Issues)
}

func TestTerraformRequiredProvidersOrderRule_MultipleFilesWithIssue(t *testing.T) {
	// Test with terraform block in one file having ordering issue
	rule := NewTerraformRequiredProvidersOrderRule()
	runner := helper.TestRunner(t, map[string]string{
		"terraform.tf": readFixture(t, "required_providers_order_multiple_files_invalid.tf"),
		"main.tf":      readFixture(t, "simple_resource.tf"),
	})

	if err := rule.Check(runner); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(runner.Issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(runner.Issues))
	}
	if runner.Issues[0].Rule.Name() != "terraform_required_providers_order" {
		t.Errorf("Expected rule terraform_required_providers_order, got %s", runner.Issues[0].Rule.Name())
	}
}

func TestTerraformRequiredProvidersOrderRule_WithComments(t *testing.T) {
	// Test that providers with comments are handled correctly
	rule := NewTerraformRequiredProvidersOrderRule()
	runner := helper.TestRunner(t, map[string]string{
		"resource.tf": readFixture(t, "required_providers_order_with_comments_valid.tf"),
	})

	if err := rule.Check(runner); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should pass - providers are in alphabetical order
	helper.AssertIssues(t, helper.Issues{}, runner.Issues)
}

func TestTerraformRequiredProvidersOrderRule_WithCommentsOutOfOrder(t *testing.T) {
	// Test that out-of-order providers with comments are detected
	rule := NewTerraformRequiredProvidersOrderRule()
	runner := helper.TestRunner(t, map[string]string{
		"resource.tf": readFixture(t, "required_providers_order_with_comments_invalid.tf"),
	})

	if err := rule.Check(runner); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(runner.Issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(runner.Issues))
	}
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
		{
			Name:         "Autofix - reorder providers with blank lines",
			ContentFile:  "required_providers_order_invalid_with_blanks.tf",
			ExpectedFile: "required_providers_order_invalid_with_blanks_expected.tf",
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
