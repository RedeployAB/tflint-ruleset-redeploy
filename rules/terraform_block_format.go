package rules

import (
	"sort"
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
func (r *TerraformBlockFormatRule) countActualBlankLines(
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

		if body, ok := syntaxFile.Body.(*hclsyntax.Body); ok {
			// Process top-level blocks with type filtering
			if err := r.processBody(body, runner, true); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *TerraformBlockFormatRule) processBody(body *hclsyntax.Body, runner tflint.Runner, isTopLevel bool) error {
	for _, blk := range body.Blocks {
		// At top level, only check specific block types
		// Inside those blocks, check ALL nested blocks
		if !isTopLevel || isBlockTypeOfInterest(blk.Type) {
			if err := r.checkBlock(blk, runner); err != nil {
				return err
			}
		}
		// Recursively process nested blocks - but not at top level anymore
		if err := r.processBody(blk.Body, runner, false); err != nil {
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

	// Gather items (attributes/blocks) in lexical order
	items, err2 := r.collectItems(block)
	if err2 != nil {
		return err2
	}

	return r.checkItemsSpacing(items, block, runner)
}

func (r *TerraformBlockFormatRule) collectItems(block *hclsyntax.Block) ([]item, error) {
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

	sort.Slice(items, func(i, j int) bool {
		return items[i].StartLine < items[j].StartLine
	})

	return items, nil
}

func (r *TerraformBlockFormatRule) checkItemsSpacing(
	items []item,
	block *hclsyntax.Block,
	runner tflint.Runner,
) error {
	files, err := runner.GetFiles()
	if err != nil {
		return err
	}
	hclFile, ok := files[block.Body.Range().Filename]
	if !ok || hclFile.Bytes == nil {
		return nil
	}
	lines := strings.Split(string(hclFile.Bytes), "\n")

	// Helper functions to check spacing logic:
	checkFirstBlockSpacing := func(linesBetween int, rng hcl.Range, hasAttributesBefore bool) error {
		if !hasAttributesBefore {
			// No attributes before this block => expect 0 blank lines
			if linesBetween != 0 {
				return r.emitIssue(runner, rng,
					"Block should appear immediately after opening brace when it's the first item (no blank lines)")
			}
		} else {
			// Has attributes before this block => expect exactly 1 blank line
			if linesBetween != 1 {
				return r.emitIssue(runner, rng, "Expected exactly one blank line before this block")
			}
		}
		return nil
	}

	checkSubsequentBlockSpacing := func(linesBetween int, rng hcl.Range) error {
		// Always expect exactly 1 blank line for subsequent blocks
		if linesBetween != 1 {
			return r.emitIssue(runner, rng, "Expected exactly one blank line before this block")
		}
		return nil
	}

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
		if firstBlock {
			if err2 := checkFirstBlockSpacing(linesBetween, it.Range, hasSeenAttributes); err2 != nil {
				return err2
			}
			firstBlock = false
		} else {
			if err2 := checkSubsequentBlockSpacing(linesBetween, it.Range); err2 != nil {
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
	return t == TypeResource ||
		t == TypeData ||
		t == TypeTerraform ||
		t == TypeProvider ||
		t == TypeVariable ||
		t == TypeOutput
}
