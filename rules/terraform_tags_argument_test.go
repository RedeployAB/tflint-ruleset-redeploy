package rules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func TestTerraformTagsArgumentRule(t *testing.T) {
	tests := []struct {
		Name    string
		Content string
		Issues  helper.Issues
	}{
		{
			Name:    "OK usage with tags last, then depends_on, then lifecycle",
			Content: readFixture(t, "tags_argument_ok_with_depends_on_and_lifecycle.tf"),
			Issues:  helper.Issues{},
		},
		{
			Name:    "OK usage with only tags",
			Content: readFixture(t, "tags_argument_ok_only_tags.tf"),
			Issues:  helper.Issues{},
		},
		{
			Name:    "OK usage with tags last, no depends_on or lifecycle",
			Content: readFixture(t, "tags_argument_ok_tags_last.tf"),
			Issues:  helper.Issues{},
		},
		{
			Name:    "NOT OK usage - normal arguments after tags",
			Content: readFixture(t, "tags_argument_normal_args_after_tags.tf"),
			Issues: helper.Issues{
				{
					Rule:    NewTerraformTagsArgumentRule(),
					Message: "Argument 'allocation_id' must not come after 'tags'",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 8, Column: 3},
						End:      hcl.Pos{Line: 8, Column: 24},
					},
				},
			},
		},
		{
			Name:    "NOT OK usage - block after tags that's not lifecycle",
			Content: readFixture(t, "tags_argument_block_after_tags.tf"),
			Issues: helper.Issues{
				{
					Rule:    NewTerraformTagsArgumentRule(),
					Message: "Block 'something_else' must not come after 'tags'",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 6, Column: 3},
						End:      hcl.Pos{Line: 6, Column: 17},
					},
				},
			},
		},
		{
			Name:    "NOT OK - missing blank line between tags and depends_on",
			Content: readFixture(t, "tags_argument_missing_blank_between_tags_and_depends_on.tf"),
			Issues: helper.Issues{
				{
					Rule:    NewTerraformTagsArgumentRule(),
					Message: "Expected exactly one blank line between 'tags' and 'depends_on'",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 5, Column: 3},
						End:      hcl.Pos{Line: 5, Column: 43},
					},
				},
			},
		},
		{
			Name:    "NOT OK - missing blank line between tags and lifecycle",
			Content: readFixture(t, "tags_argument_missing_blank_between_tags_and_lifecycle.tf"),
			Issues: helper.Issues{
				{
					Rule:    NewTerraformTagsArgumentRule(),
					Message: "Expected exactly one blank line between 'tags' and 'lifecycle'",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 5, Column: 3},
						End:      hcl.Pos{Line: 5, Column: 12},
					},
				},
			},
		},
	}

	rule := NewTerraformTagsArgumentRule()

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
