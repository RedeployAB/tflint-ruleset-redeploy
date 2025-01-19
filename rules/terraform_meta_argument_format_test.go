package rules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func TestTerraformMetaArgumentFormat(t *testing.T) {
	tests := []struct {
		Name    string
		Content string
		Issues  helper.Issues
	}{
		{
			Name:    "resource with no meta arguments",
			Content: readFixture(t, "meta_fmt_no_meta_args.tf"),
			Issues:  helper.Issues{},
		},
		{
			Name:    "module with no meta arguments",
			Content: readFixture(t, "meta_fmt_module_no_meta.tf"),
			Issues:  helper.Issues{},
		},
		{
			Name:    "resource with top meta argument on last line (no blank line required)",
			Content: readFixture(t, "meta_fmt_top_meta_last_line.tf"),
			Issues:  helper.Issues{},
		},
		{
			Name:    "resource missing blank line after top meta argument",
			Content: readFixture(t, "meta_fmt_missing_blank_after_top.tf"),
			Issues: helper.Issues{
				{
					Rule:    NewTerraformMetaArgumentFormatRule(),
					Message: "Expected a blank line after meta-arguments (count/for_each/provider)",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 5, Column: 1},
						End:      hcl.Pos{Line: 5, Column: 1},
					},
				},
			},
		},
		{
			Name:    "resource missing blank line before bottom meta argument (depends_on)",
			Content: readFixture(t, "meta_fmt_missing_blank_before_depends_on.tf"),
			Issues: helper.Issues{
				{
					Rule:    NewTerraformMetaArgumentFormatRule(),
					Message: `Expected a blank line before meta-argument 'depends_on'`,
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 4, Column: 1},
						End:      hcl.Pos{Line: 4, Column: 1},
					},
				},
			},
		},
		{
			Name:    "module valid formatting with top and bottom meta arguments",
			Content: readFixture(t, "meta_fmt_module_valid.tf"),
			Issues:  helper.Issues{},
		},
		{
			Name:    "module missing blank line before bottom meta argument (depends_on)",
			Content: readFixture(t, "meta_fmt_module_missing_blank_before_depends_on.tf"),
			Issues: helper.Issues{
				{
					Rule:    NewTerraformMetaArgumentFormatRule(),
					Message: `Expected a blank line before meta-argument 'depends_on'`,
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 4, Column: 1},
						End:      hcl.Pos{Line: 4, Column: 1},
					},
				},
			},
		},
		{
			Name:    "resource partial usage (only provider) - valid formatting",
			Content: readFixture(t, "meta_fmt_only_provider_valid.tf"),
			Issues:  helper.Issues{},
		},
		{
			Name:    "resource partial usage (only lifecycle) - missing blank line before",
			Content: readFixture(t, "meta_fmt_missing_blank_before_lifecycle.tf"),
			Issues: helper.Issues{
				{
					Rule:    NewTerraformMetaArgumentFormatRule(),
					Message: `Expected a blank line before meta-argument 'lifecycle'`,
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 6, Column: 1},
						End:      hcl.Pos{Line: 6, Column: 1},
					},
				},
			},
		},
	}

	rule := NewTerraformMetaArgumentFormatRule()

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
