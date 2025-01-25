package rules

import (
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// TerraformBlockFormatRule enforces that within certain block types (resource, data, terraform, provider):
//  1. The first block (if any) appears immediately after the opening brace or
//     after exactly one blank line if there are attributes above it.
//  2. Any subsequent blocks in the same block appear after exactly one blank line.
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
	for _, blk := range body.Blocks {
		if isBlockTypeOfInterest(blk.Type) {
			if err := r.checkBlock(blk, runner); err != nil {
				return err
			}
		}
		if err := r.processBody(blk.Body, runner); err != nil {
			return err
		}
	}
	return nil
}

func (r *TerraformBlockFormatRule) checkBlock(block *hclsyntax.Block, runner tflint.Runner) error {
	files, err := runner.GetFiles()
	if err != nil {
		return err
	}
	hclFile, ok := files[block.Body.Range().Filename]
	if !ok || hclFile.Bytes == nil {
		return nil
	}

	// First, detect whether this block has attributes at all:
	hasAttributes := len(block.Body.Attributes) > 0

	type item struct {
		Type      string
		Range     hcl.Range
		StartLine int
		EndLine   int
	}
	var items []item

	for _, attr := range block.Body.Attributes {
		items = append(items, item{
			Type:      TypeAttr,
			Range:     attr.Range(),
			StartLine: attr.Range().Start.Line,
			EndLine:   attr.Range().End.Line,
		})
	}
	for _, blk2 := range block.Body.Blocks {
		blkStart := blk2.Body.Range().Start.Line
		blkEnd := blk2.Body.Range().End.Line

		items = append(items, item{
			Type:      TypeBlock,
			Range:     blk2.DefRange(),
			StartLine: blkStart,
			EndLine:   blkEnd,
		})
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].StartLine < items[j].StartLine
	})

	// Use DefRange().Start.Line for the line with the 'resource'/'data'/'provider' etc.
	previousEndLine := block.DefRange().Start.Line

	firstBlock := true

	for _, it := range items {
		if it.Type == TypeAttr {
			previousEndLine = it.EndLine
			continue
		}

		linesBetween := it.StartLine - (previousEndLine + 1)

		// If this is the FIRST block in the parent:
		if firstBlock {
			// If the parent has no attributes at all, then 0 blank lines are allowed
			// (i.e. the block must appear right after the brace). Otherwise, we require exactly one blank line.
			if !hasAttributes {
				// No attributes => expect 0 lines
				if linesBetween != 0 {
					if err2 := r.emitIssue(runner, it.Range,
						"Block should appear immediately after opening brace when it's the first item (no blank lines)"); err2 != nil {
						return err2
					}
				}
			} else {
				// We have attributes => expect exactly 1 blank line
				if linesBetween != 1 {
					if err2 := r.emitIssue(runner, it.Range, "Expected exactly one blank line before this block"); err2 != nil {
						return err2
					}
				}
			}
			firstBlock = false
		} else if linesBetween != 1 {
			if err2 := r.emitIssue(runner, it.Range, "Expected exactly one blank line before this block"); err2 != nil {
				return err2
			}
		}

		previousEndLine = it.EndLine
	}

	return nil
}

func (r *TerraformBlockFormatRule) emitIssue(runner tflint.Runner, rng hcl.Range, msg string) error {
	return runner.EmitIssue(r, msg, rng)
}

func isBlockTypeOfInterest(typ string) bool {
	t := strings.ToLower(typ)
	return t == TypeResource || t == TypeData || t == TypeTerraform || t == TypeProvider
}
