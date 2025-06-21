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
			Name:    "OK - only source, block ends",
			Content: readFixture(t, "source_format_only_source.tf"),
			Issues:  helper.Issues{},
		},
		{
			Name:    "OK - source plus version, block ends",
			Content: readFixture(t, "source_format_source_version.tf"),
			Issues:  helper.Issues{},
		},
		{
			Name:    "NOT OK - source plus version, extra blank line at end is disallowed",
			Content: readFixture(t, "source_format_extra_blank_line.tf"),
			Issues: helper.Issues{
				{
					Rule:    NewTerraformSourceFormatRule(),
					Message: "Unexpected blank line after 'version' when block ends",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 4, Column: 1},
						End:      hcl.Pos{Line: 4, Column: 1},
					},
				},
			},
		},
		{
			Name:    "NOT OK - source alone with trailing blank line before closing brace",
			Content: readFixture(t, "source_format_source_trailing_blank.tf"),
			Issues: helper.Issues{
				{
					Rule:    NewTerraformSourceFormatRule(),
					Message: "Unexpected blank line after 'source' when block ends",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 3, Column: 1},
						End:      hcl.Pos{Line: 3, Column: 1},
					},
				},
			},
		},
		{
			Name:    "OK - source plus version, more property after blank line",
			Content: readFixture(t, "source_format_source_version_with_property.tf"),
			Issues:  helper.Issues{},
		},
		{
			Name:    "OK - source alone, more property after blank line",
			Content: readFixture(t, "source_format_source_with_property.tf"),
			Issues:  helper.Issues{},
		},
		{
			Name:    "NOT OK - source plus version, property follows with no blank line",
			Content: readFixture(t, "source_format_no_blank_before_property.tf"),
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

func TestTerraformSourceFormat_Autofix(t *testing.T) {
	tests := []struct {
		Name     string
		Content  string
		Expected string
		HasFix   bool
	}{
		{
			Name: "Add blank line after source when property follows",
			Content: `module "example" {
  source   = "a source address"
  property = "value"
}`,
			Expected: `module "example" {
  source = "a source address"

  property = "value"
}`,
			HasFix: true,
		},
		{
			Name: "Add blank line after version when property follows",
			Content: `module "example" {
  source   = "a source address"
  version  = "x.x.x"
  property = "value"
}`,
			Expected: `module "example" {
  source  = "a source address"
  version = "x.x.x"

  property = "value"
}`,
			HasFix: true,
		},
		{
			Name: "Remove blank line after source when block ends",
			Content: `module "example" {
  source = "a source address"

}`,
			Expected: `module "example" {
  source = "a source address"
}`,
			HasFix: true,
		},
		{
			Name: "Remove blank line after version when block ends",
			Content: `module "example" {
  source  = "a source address"
  version = "1.0.0"

}`,
			Expected: `module "example" {
  source  = "a source address"
  version = "1.0.0"
}`,
			HasFix: true,
		},
		{
			Name: "Preserve correct format with blank line",
			Content: `module "example" {
  source  = "a source address"
  version = "x.x.x"

  property = "value"
}`,
			Expected: `module "example" {
  source  = "a source address"
  version = "x.x.x"

  property = "value"
}`,
			HasFix: false,
		},
		{
			Name: "Preserve correct format without blank line when block ends",
			Content: `module "example" {
  source = "a source address"
}`,
			Expected: `module "example" {
  source = "a source address"
}`,
			HasFix: false,
		},
		{
			Name: "Handle comments after source/version",
			Content: `module "example" {
  source  = "a source address"
  version = "x.x.x"
  # This is a comment
  property = "value"
}`,
			Expected: `module "example" {
  source  = "a source address"
  version = "x.x.x"
  # This is a comment
  property = "value"
}`,
			HasFix: false,
		},
		{
			Name: "Multiple blank lines after source when block ends",
			Content: `module "example" {
  source = "a source address"


}`,
			Expected: `module "example" {
  source = "a source address"

}`,
			HasFix: true,
		},
	}

	rule := NewTerraformSourceFormatRule()

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runner := helper.TestRunner(t, map[string]string{
				"main.tf": tc.Content,
			})

			if err := rule.Check(runner); err != nil {
				t.Fatalf("Unexpected error: %s", err)
			}

			changes := runner.Changes()
			if tc.HasFix {
				helper.AssertChanges(t, map[string]string{
					"main.tf": tc.Expected,
				}, changes)
			} else if len(changes) > 0 {
				t.Errorf("Expected no changes, but got: %v", changes)
			}
		})
	}
}
