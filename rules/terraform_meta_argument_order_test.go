package rules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func TestTerraformMetaArgumentOrder(t *testing.T) {
	tests := []struct {
		Name     string
		Content  string
		Expected helper.Issues
	}{
		{
			Name:     "valid resource order",
			Content:  readFixture(t, "meta_order_valid_resource.tf"),
			Expected: helper.Issues{},
		},
		{
			Name:     "resource partial usage (count only)",
			Content:  readFixture(t, "meta_order_resource_count_only.tf"),
			Expected: helper.Issues{},
		},
		{
			Name:     "resource with no meta arguments",
			Content:  readFixture(t, "meta_order_resource_no_meta.tf"),
			Expected: helper.Issues{},
		},
		{
			Name:    "invalid resource order",
			Content: readFixture(t, "meta_order_invalid_resource.tf"),
			Expected: helper.Issues{
				{
					Rule:    NewTerraformMetaArgumentOrderRule(),
					Message: "Out-of-order meta argument 'depends_on' in resource 'aws_instance example': must appear after all resource arguments and blocks",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 2, Column: 3},
						End:      hcl.Pos{Line: 2, Column: 18},
					},
				},
			},
		},
		{
			Name:     "valid module order",
			Content:  readFixture(t, "meta_order_valid_module.tf"),
			Expected: helper.Issues{},
		},
		{
			Name:    "invalid module order",
			Content: readFixture(t, "meta_order_invalid_module.tf"),
			Expected: helper.Issues{
				{
					Rule:    NewTerraformMetaArgumentOrderRule(),
					Message: "Out-of-order meta argument 'depends_on' in module 'example': must appear after all module arguments and blocks",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 2, Column: 3},
						End:      hcl.Pos{Line: 2, Column: 18},
					},
				},
			},
		},
		{
			Name:     "valid resource with content and bottom meta-args",
			Content:  readFixture(t, "meta_order_valid_resource_with_content.tf"),
			Expected: helper.Issues{},
		},
		{
			Name:    "lifecycle before content",
			Content: readFixture(t, "meta_order_invalid_lifecycle_before_content.tf"),
			Expected: helper.Issues{
				{
					Rule:    NewTerraformMetaArgumentOrderRule(),
					Message: "Out-of-order meta argument 'lifecycle' in resource 'azurerm_container_app example': must appear after all resource arguments and blocks",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 2, Column: 3},
						End:      hcl.Pos{Line: 2, Column: 12},
					},
				},
			},
		},
		{
			Name:    "depends_on before content",
			Content: readFixture(t, "meta_order_invalid_depends_on_before_content.tf"),
			Expected: helper.Issues{
				{
					Rule:    NewTerraformMetaArgumentOrderRule(),
					Message: "Out-of-order meta argument 'depends_on' in resource 'aws_instance example': must appear after all resource arguments and blocks",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 2, Column: 3},
						End:      hcl.Pos{Line: 2, Column: 18},
					},
				},
			},
		},
		{
			Name:    "module depends_on before content",
			Content: readFixture(t, "meta_order_invalid_module_depends_on_before_content.tf"),
			Expected: helper.Issues{
				{
					Rule:    NewTerraformMetaArgumentOrderRule(),
					Message: "Out-of-order meta argument 'depends_on' in module 'example': must appear after all module arguments and blocks",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 2, Column: 3},
						End:      hcl.Pos{Line: 2, Column: 18},
					},
				},
			},
		},
		{
			Name:    "both bottom meta-args before content (first violation only)",
			Content: readFixture(t, "meta_order_invalid_both_bottom_before_content.tf"),
			Expected: helper.Issues{
				{
					Rule:    NewTerraformMetaArgumentOrderRule(),
					Message: "Out-of-order meta argument 'lifecycle' in resource 'aws_instance example': must appear after all resource arguments and blocks",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 2, Column: 3},
						End:      hcl.Pos{Line: 2, Column: 12},
					},
				},
			},
		},
		{
			Name:    "top meta-arg for_each after content",
			Content: readFixture(t, "meta_order_invalid_top_meta_after_content.tf"),
			Expected: helper.Issues{
				{
					Rule:    NewTerraformMetaArgumentOrderRule(),
					Message: "Out-of-order meta argument 'for_each' in resource 'azurerm_role_assignment blob_contributor': must appear before all resource arguments and blocks",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 6, Column: 3},
						End:      hcl.Pos{Line: 6, Column: 34},
					},
				},
			},
		},
		{
			Name:    "top meta-arg provider after content",
			Content: readFixture(t, "meta_order_invalid_top_provider_after_content.tf"),
			Expected: helper.Issues{
				{
					Rule:    NewTerraformMetaArgumentOrderRule(),
					Message: "Out-of-order meta argument 'provider' in resource 'aws_instance example': must appear before all resource arguments and blocks",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 5, Column: 3},
						End:      hcl.Pos{Line: 5, Column: 22},
					},
				},
			},
		},
		{
			Name:    "top meta-arg count after content (module)",
			Content: readFixture(t, "meta_order_invalid_top_count_after_content_module.tf"),
			Expected: helper.Issues{
				{
					Rule:    NewTerraformMetaArgumentOrderRule(),
					Message: "Out-of-order meta argument 'count' in module 'example': must appear before all module arguments and blocks",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 5, Column: 3},
						End:      hcl.Pos{Line: 5, Column: 12},
					},
				},
			},
		},
		{
			Name:     "valid resource with top content and bottom meta-args",
			Content:  readFixture(t, "meta_order_valid_top_and_bottom.tf"),
			Expected: helper.Issues{},
		},
		{
			Name:     "valid resource with only top meta-args",
			Content:  readFixture(t, "meta_order_valid_top_only.tf"),
			Expected: helper.Issues{},
		},
		{
			Name:    "both top and bottom violations (top reported first)",
			Content: readFixture(t, "meta_order_invalid_top_and_bottom_violations.tf"),
			Expected: helper.Issues{
				{
					Rule:    NewTerraformMetaArgumentOrderRule(),
					Message: "Out-of-order meta argument 'for_each' in resource 'aws_instance example': must appear before all resource arguments and blocks",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 7, Column: 3},
						End:      hcl.Pos{Line: 7, Column: 27},
					},
				},
			},
		},
	}

	rule := NewTerraformMetaArgumentOrderRule()

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runner := helper.TestRunner(t, map[string]string{
				"resource.tf": tc.Content,
			})

			if err := rule.Check(runner); err != nil {
				t.Fatalf("Unexpected error occurred: %s", err)
			}

			helper.AssertIssues(t, tc.Expected, runner.Issues)
		})
	}
}

func TestTerraformMetaArgumentOrder_Autofix(t *testing.T) {
	tests := []struct {
		Name         string
		ContentFile  string
		ExpectedFile string
	}{
		{
			Name:         "Autofix - lifecycle before content",
			ContentFile:  "meta_order_autofix_lifecycle_before_content.tf",
			ExpectedFile: "meta_order_autofix_lifecycle_before_content_expected.tf",
		},
		{
			Name:         "Autofix - depends_on before content",
			ContentFile:  "meta_order_autofix_depends_on_before_content.tf",
			ExpectedFile: "meta_order_autofix_depends_on_before_content_expected.tf",
		},
		{
			Name:         "Autofix - both bottom meta-args before content",
			ContentFile:  "meta_order_autofix_both_before_content.tf",
			ExpectedFile: "meta_order_autofix_both_before_content_expected.tf",
		},
		{
			Name:         "Autofix - module depends_on before source",
			ContentFile:  "meta_order_autofix_module.tf",
			ExpectedFile: "meta_order_autofix_module_expected.tf",
		},
		{
			Name:         "Autofix - comments before blocks are preserved",
			ContentFile:  "meta_order_autofix_comments_preserved.tf",
			ExpectedFile: "meta_order_autofix_comments_preserved_expected.tf",
		},
		{
			Name:         "Autofix - move for_each to top of resource",
			ContentFile:  "meta_order_autofix_top_for_each.tf",
			ExpectedFile: "meta_order_autofix_top_for_each_expected.tf",
		},
		{
			Name:         "Autofix - move provider to top of resource",
			ContentFile:  "meta_order_autofix_top_provider.tf",
			ExpectedFile: "meta_order_autofix_top_provider_expected.tf",
		},
		{
			Name:         "Autofix - move count to top of module",
			ContentFile:  "meta_order_autofix_top_count_module.tf",
			ExpectedFile: "meta_order_autofix_top_count_module_expected.tf",
		},
		{
			Name:         "Autofix - comments preserved when moving top meta-args",
			ContentFile:  "meta_order_autofix_top_comments_preserved.tf",
			ExpectedFile: "meta_order_autofix_top_comments_preserved_expected.tf",
		},
	}

	rule := NewTerraformMetaArgumentOrderRule()

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			content := readFixture(t, tc.ContentFile)
			expected := readFixture(t, tc.ExpectedFile)

			runner := helper.TestRunner(t, map[string]string{"resource.tf": content})

			if err := rule.Check(runner); err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			helper.AssertChanges(t, map[string]string{"resource.tf": expected}, runner.Changes())
		})
	}
}
