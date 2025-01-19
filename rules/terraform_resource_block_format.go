package rules

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// TerraformResourceBlockFormatRule enforces that within a resource block:
// 1) The first block (if any) appears immediately after the opening brace or
//    after exactly one blank line if there are attributes above it.
// 2) Any subsequent blocks in the same resource appear after exactly one blank line.
//
// The same logic applies to nested blocks of nested blocks, recursively.
//
// Examples:
// OK:
//   resource "example" "foo" {
//     name = "test"
//
//     block1 { ... }
//   }
//
// NOT OK:
//   resource "example" "foo" {
//     name = "test"
//     block1 { ... }   # missing blank line
//   }
type TerraformResourceBlockFormatRule struct {
	tflint.DefaultRule
}

func NewTerraformResourceBlockFormatRule() *TerraformResourceBlockFormatRule {
	return &TerraformResourceBlockFormatRule{}
}

func (r *TerraformResourceBlockFormatRule) Name() string {
	return "terraform_resource_block_format"
}

func (r *TerraformResourceBlockFormatRule) Enabled() bool {
	return true
}

func (r *TerraformResourceBlockFormatRule) Severity() tflint.Severity {
	return tflint.ERROR
}

func (r *TerraformResourceBlockFormatRule) Link() string {
	return ""
}

func (r *TerraformResourceBlockFormatRule) Check(runner tflint.Runner) error {
	files, err := runner.GetFiles()
	if err != nil {
		return err
	}
	for filename, hclFile := range files {
		if hclFile == nil || hclFile.Bytes == nil {
			continue
		}

		syntaxFile, diags := hclsyntax.ParseConfig(hclFile.Bytes, filename, hcl.InitialPos)
		if diags.HasErrors() {
			// skip file with parse errors
			continue
		}

		if body, ok := syntaxFile.Body.(*hclsyntax.Body); ok {
			if err := r.processBody(body, runner); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *TerraformResourceBlockFormatRule) processBody(body *hclsyntax.Body, runner tflint.Runner) error {
	// Recurse into all blocks
	for _, blk := range body.Blocks {
		// If it's resource, check the block format inside
		if blk.Type == "resource" {
			if err := r.checkResourceBlock(blk, runner); err != nil {
				return err
			}
		}
		// Recurse deeper
		if err := r.processBody(blk.Body, runner); err != nil {
			return err
		}
	}
	return nil
}

// checkResourceBlock ensures that any blocks inside "resource" (like nested blocks)
// are separated by exactly one blank line from the prior item (be it an attribute or block),
// except if it's the very first item, in which case it can appear immediately if it’s first
// or after exactly one blank line if an attribute precedes it.
func (r *TerraformResourceBlockFormatRule) checkResourceBlock(resourceBlock *hclsyntax.Block, runner tflint.Runner) error {
	files, err := runner.GetFiles()
	if err != nil {
		return err
	}
	hclFile, ok := files[resourceBlock.Body.Range().Filename]
	if !ok || hclFile.Bytes == nil {
		return nil
	}
	lines := strings.Split(string(hclFile.Bytes), "\n")

	// We'll gather the "items" in lexical order (attributes + blocks).
	type item struct {
		Type  string // "attr" or "block"
		Range hcl.Range
	}
	var items []item

	// Add attributes
	for _, attr := range resourceBlock.Body.Attributes {
		items = append(items, item{"attr", attr.Range()})
	}
	// Add blocks
	for _, blk := range resourceBlock.Body.Blocks {
		items = append(items, item{"block", blk.DefRange()})
	}

	// Sort by position
	sort.Slice(items, func(i, j int) bool {
		return items[i].Range.Start.Byte < items[j].Range.Start.Byte
	})

	// We also need the line of the resource's opening brace, so we can treat that
	// as the "previous item" if there's nothing else preceding the first block.
	openBraceLine := resourceBlock.DefRange().End.Line
	// resourceBlock.DefRange() ends at the close of "resource <type> <name>"
	// Typically the brace is on the same line or the next column
	// but we assume the next line is the block body.
	//
	// We can approximate: if the brace is on the same line, openBraceLine
	// is the correct line. We'll use that as our reference line.

	var prevLine int
	noPriorItems := true

	for _, it := range items {
		if it.Type != "block" {
			// If it's an attribute, just track the end line
			prevLine = it.Range.End.Line
			noPriorItems = false
			continue
		}

		// It's a block. We want to see how many blank lines are between
		// the prior item (or the opening brace line if none) and this block's start.
		blockStartLine := it.Range.Start.Line

		referenceLine := openBraceLine
		if !noPriorItems {
			referenceLine = prevLine
		}

		linesBetween := blockStartLine - referenceLine - 1

		if noPriorItems {
			// If it's the first item, linesBetween must be 0 for no extra blank lines
			if linesBetween != 0 {
				return r.emitIssue(runner, it.Range,
					"Block should appear immediately after opening brace when it's the first item (no blank lines)")
			}
		} else {
			// Not the first item -> we expect exactly 1 blank line
			if linesBetween != 1 {
				return r.emitIssue(runner, it.Range,
					"Expected exactly one blank line before this block")
			}
		}

		// Set prevLine to this block's end
		prevLine = it.Range.End.Line
		noPriorItems = false
	}

	return nil
}

func (r *TerraformResourceBlockFormatRule) emitIssue(runner tflint.Runner, rng hcl.Range, msg string) error {
	return runner.EmitIssue(r, msg, rng)
}
