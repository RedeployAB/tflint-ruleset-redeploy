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
		{
			Name:    "OK - attribute then block with blank line",
			Content: readFixture(t, "block_fmt_ok_attr_then_block_blank_line.tf"),
			Issues:  helper.Issues{},
		},
		{
			Name:    "NOT OK - attribute then block with no blank line",
			Content: readFixture(t, "block_fmt_not_ok_attr_then_block_no_blank_line.tf"),
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
			Name:    "OK - single block first (no attribute), no blank line after brace",
			Content: readFixture(t, "block_fmt_ok_single_block_first.tf"),
			Issues:  helper.Issues{},
		},
		{
			Name:    "NOT OK - single block first with extra blank line after brace",
			Content: readFixture(t, "block_fmt_not_ok_single_block_first_extra_blank.tf"),
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
			Name:    "OK - multiple blocks each separated by single blank line",
			Content: readFixture(t, "block_fmt_ok_multiple_blocks_single_blank_line.tf"),
			Issues:  helper.Issues{},
		},
		{
			Name:    "NOT OK - multiple blocks no blank line between them",
			Content: readFixture(t, "block_fmt_not_ok_multiple_blocks_no_blank_line.tf"),
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
						End:      hcl.Pos{Line: 11, Column: 22},
					},
				},
			},
		},
		{
			Name:    "OK - attributes then multiple blocks each separated by single blank line",
			Content: readFixture(t, "block_fmt_ok_attr_multiple_blocks_single_blank_line.tf"),
			Issues:  helper.Issues{},
		},
		{
			Name:    "NOT OK - attribute then block with no blank line, then next block also with no blank line",
			Content: readFixture(t, "block_fmt_not_ok_attr_no_blank_line_multiple_blocks.tf"),
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
						End:      hcl.Pos{Line: 8, Column: 26},
					},
				},
			},
		},
		{
			Name:    "OK - single data block with first sub-block no blank line",
			Content: readFixture(t, "block_fmt_ok_data_single_block_no_blank_line.tf"),
			Issues:  helper.Issues{},
		},