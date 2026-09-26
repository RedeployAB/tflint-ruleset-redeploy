package rules

import (
	"cmp"
	"slices"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// TerraformBlockFormatRule enforces that within top-level Terraform blocks
// (resource, data, terraform, provider, variable, output) and ALL their nested blocks:
//  1. The first block (if any) appears immediately after the opening brace or
//     after exactly one blank line if there are attributes above it.
//  2. Any subsequent blocks in the same block appear after exactly one blank line.
//
// The rule checks formatting for all nested blocks within these top-level constructs,
// including provider-specific blocks like metric_query, ingress, egress, etc.
type TerraformBlockFormatRule struct {
	tflint.DefaultRule
}

// item represents either an attribute or nested block inside our target block
type item struct {
	Type      string
	Range     hcl.Range
	StartLine int
	EndLine   int
}

// countActualBlankLines returns how many groups of actual blank lines exist
// between fromLine (inclusive) and toLine (exclusive). Comment lines are
// NOT counted as blank lines. It returns:
//   - 0 if there are no blank lines
//   - 1 if there is exactly one group of contiguous blank lines
//   - 2 or more if multiple separate groups of blank lines appear
func (*TerraformBlockFormatRule) countActualBlankLines(
	lines []string,
	fromLine, toLine int,
) int {
	if fromLine >= toLine {
		return 0
	}
	blankGroups := 0
	inBlankGroup := false
	for i := fromLine; i < toLine && i < len(lines); i++ {
		s := strings.TrimSpace(lines[i])
		if s == "" {
			// This is an actual blank line
			if !inBlankGroup {
				blankGroups++
				inBlankGroup = true
			}
		} else {
			// This line has content (could be code or comment)
			inBlankGroup = false
		}
	}
	return blankGroups
}

func NewTerraformBlockFormatRule() *TerraformBlockFormatRule {
	return &TerraformBlockFormatRule{}
}

func (*TerraformBlockFormatRule) Name() string {
	return "terraform_block_format"
}

func (*TerraformBlockFormatRule) Enabled() bool {
	return true
}

func (*TerraformBlockFormatRule) Severity() tflint.Severity {
	return tflint.ERROR
}

func (r *TerraformBlockFormatRule) Link() string {
	return GetRuleDocLink(r.Name())
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

		// Pre-split lines for this file to avoid repeated splitting
		lines := strings.Split(string(hclFile.Bytes), "\n")

		if body, ok := syntaxFile.Body.(*hclsyntax.Body); ok {
			// Process top-level blocks with type filtering
			if err := r.processTopLevelBody(body, runner, lines); err != nil {
				return err
			}
		}
	}
	return nil
}

// processTopLevelBody checks the top-level blocks of interest and
// recursively ALL nested blocks inside every top-level block.
func (r *TerraformBlockFormatRule) processTopLevelBody(
	body *hclsyntax.Body,
	runner tflint.Runner,
	lines []string,
) error {
	for _, blk := range body.Blocks {
		// At top level, only check specific block types
		if isBlockTypeOfInterest(blk.Type) {
			if err := r.checkBlock(blk, runner, lines); err != nil {
				return err
			}
		}
		if err := r.processNestedBody(blk.Body, runner, lines); err != nil {
			return err
		}
	}
	return nil
}

// processNestedBody checks ALL blocks in a nested body, recursively.
func (r *TerraformBlockFormatRule) processNestedBody(
	body *hclsyntax.Body,
	runner tflint.Runner,
	lines []string,
) error {
	for _, blk := range body.Blocks {
		if err := r.checkBlock(blk, runner, lines); err != nil {
			return err
		}
		if err := r.processNestedBody(blk.Body, runner, lines); err != nil {
			return err
		}
	}
	return nil
}

func (r *TerraformBlockFormatRule) checkBlock(block *hclsyntax.Block, runner tflint.Runner, lines []string) error {
	// Gather items (attributes/blocks) in lexical order
	items, err := r.collectItems(block)
	if err != nil {
		return err
	}

	return r.checkItemsSpacing(items, block, runner, lines)
}

func (*TerraformBlockFormatRule) collectItems(block *hclsyntax.Block) ([]item, error) {
	var items []item

	for _, attr := range block.Body.Attributes {
		items = append(items, item{
			Type:      TypeAttr,
			Range:     attr.Range(),
			StartLine: attr.Range().Start.Line,
			EndLine:   attr.Range().End.Line,
		})
	}
	for _, childBlk := range block.Body.Blocks {
		blkStart := childBlk.DefRange().Start.Line
		blkEnd := childBlk.Body.Range().End.Line

		items = append(items, item{
			Type:      TypeBlock,
			Range:     childBlk.DefRange(),
			StartLine: blkStart,
			EndLine:   blkEnd,
		})
	}

	slices.SortFunc(items, func(a, b item) int {
		return cmp.Compare(a.StartLine, b.StartLine)
	})

	return items, nil
}

func (r *TerraformBlockFormatRule) checkItemsSpacing(
	items []item,
	block *hclsyntax.Block,
	runner tflint.Runner,
	lines []string,
) error {
	// Use DefRange().Start.Line for the line with the 'resource'/'data'/'provider' etc.
	previousEndLine := block.DefRange().Start.Line
	firstBlock := true
	hasSeenAttributes := false

	for _, it := range items {
		if it.Type == TypeAttr {
			hasSeenAttributes = true
			previousEndLine = it.EndLine
			continue
		}

		// Instead of raw arithmetic, we count ignoring comment-only lines
		linesBetween := r.countActualBlankLines(
			lines,
			previousEndLine, // fromLine (inclusive)
			it.StartLine,    // toLine   (exclusive)
		)
		expectedBlankLines := 1
		if firstBlock && !hasSeenAttributes {
			// No attributes before the first block => expect 0 blank lines
			expectedBlankLines = 0
		}
		if linesBetween != expectedBlankLines {
			if err := r.emitIssue(runner, it.Range, blockSpacingMessage(expectedBlankLines)); err != nil {
				return err
			}
		}
		firstBlock = false
		previousEndLine = it.EndLine
	}

	return nil
}

// blockSpacingMessage returns the issue message for a nested block that is
// not preceded by the expected number of blank lines.
func blockSpacingMessage(expectedBlankLines int) string {
	if expectedBlankLines == 0 {
		return "Block should appear immediately after opening brace when it's the first item (no blank lines)"
	}
	return "Expected exactly one blank line before this block"
}

func (r *TerraformBlockFormatRule) emitIssue(runner tflint.Runner, rng hcl.Range, msg string) error {
	return runner.EmitIssue(r, msg, rng)
}

func isBlockTypeOfInterest(typ string) bool {
	// Block types in Terraform are always lowercase, so direct comparison is safe and faster
	switch typ {
	case TypeResource, TypeData, TypeTerraform, TypeProvider, TypeVariable, TypeOutput:
		return true
	}
	return false
}
