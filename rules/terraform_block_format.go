package rules

import (
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// TerraformBlockFormatRule enforces that within certain block types (resource, data, terraform, provider):
// 1) The first block (if any) appears immediately after the opening brace or
//    after exactly one blank line if there are attributes above it.
// 2) Any subsequent blocks in the same block appear after exactly one blank line.
//
// The same logic applies to nested blocks of nested blocks, recursively.
type TerraformBlockFormatRule struct {
	tflint.DefaultRule
}

func NewTerraformBlockFormatRule() *TerraformBlockFormatRule {
	return &TerraformBlockFormatRule{}
}

func (r *TerraformBlockFormatRule) Name() string {
	return "terraform_block_format"
}

func (r *TerraformBlockFormatRule) Enabled() bool {
	return true
}

func (r *TerraformBlockFormatRule) Severity() tflint.Severity {
	return tflint.ERROR
}

func (r *TerraformBlockFormatRule) Link() string {
	return ""
}

func (r *TerraformBlockFormatRule) Check(runner tflint.Runner) error {
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
			// skip file w/ parse errors
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

func (r *TerraformBlockFormatRule) processBody(body *hclsyntax.Body, runner tflint.Runner) error {
	// Recurse into all blocks
	for _, blk := range body.Blocks {
		// If block.Type is in [resource, data, terraform, provider], check
		if isBlockTypeOfInterest(blk.Type) {
			if err := r.checkBlock(blk, runner); err != nil {
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

// checkBlock ensures that any sub-blocks inside the relevant block type
// are separated by exactly one blank line from the prior item (be it an attribute or block),
// except if it's the very first item, in which case it can appear immediately if it’s first
// or after exactly one blank line if an attribute precedes it.
func (r *TerraformBlockFormatRule) checkBlock(block *hclsyntax.Block, runner tflint.Runner) error {
	files, err := runner.GetFiles()
	if err != nil {
		return err
	}
	hclFile, ok := files[block.Body.Range().Filename]
	if !ok || hclFile.Bytes == nil {
		return nil
	}

	// We'll gather the "items" in lexical order (attributes + blocks).
	type item struct {
		Type  string // "attr" or "block"
		Range hcl.Range
	}
	var items []item

	// Add attributes in the block
	for _, attr := range block.Body.Attributes {
		items = append(items, item{"attr", attr.Range()})
	}
	// Add sub-blocks in the block
	for _, blk2 := range block.Body.Blocks {
		items = append(items, item{"block", blk2.DefRange()})
	}

	// Sort by position
	sort.Slice(items, func(i, j int) bool {
		return items[i].Range.Start.Byte < items[j].Range.Start.Byte
	})

	// We also need the line of the block's opening brace
	openBraceLine := block.Body.Range().Start.Line

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

		linesBetween := blockStartLine - (referenceLine + 1)

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

func (r *TerraformBlockFormatRule) emitIssue(runner tflint.Runner, rng hcl.Range, msg string) error {
	return runner.EmitIssue(r, msg, rng)
}

func isBlockTypeOfInterest(typ string) bool {
	// apply same format rule to resource, data, terraform, provider
	// add more if needed
	t := strings.ToLower(typ)
	return t == "resource" || t == "data" || t == "terraform" || t == "provider"
}
