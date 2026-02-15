package rules

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

type TerraformArgumentOrderRule struct {
	tflint.DefaultRule
}

// metaOrderItem represents an attribute or block within a resource/module block
type metaOrderItem struct {
	name     string
	startPos int
	isBlock  bool
	isBottom bool
}

func NewTerraformMetaArgumentOrderRule() *TerraformArgumentOrderRule {
	return &TerraformArgumentOrderRule{}
}

func (r *TerraformArgumentOrderRule) Name() string {
	return "terraform_meta_argument_order"
}

func (r *TerraformArgumentOrderRule) Enabled() bool {
	return true
}

func (r *TerraformArgumentOrderRule) Severity() tflint.Severity {
	return tflint.ERROR
}

func (r *TerraformArgumentOrderRule) Link() string {
	return GetRuleDocLink(r.Name())
}

func (r *TerraformArgumentOrderRule) Check(runner tflint.Runner) error {
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
			if err := r.processBody(body, filename, runner); err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *TerraformArgumentOrderRule) processBody(body *hclsyntax.Body, filename string, runner tflint.Runner) error {
	for _, block := range body.Blocks {
		if block.Type == TypeResource || block.Type == TypeModule {
			if err := r.checkBlock(block, runner); err != nil {
				return err
			}
		} else {
			if err := r.processBody(block.Body, filename, runner); err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *TerraformArgumentOrderRule) checkBlock(block *hclsyntax.Block, runner tflint.Runner) error {
	desiredSequence := r.getDesiredSequence(block.Type)
	if desiredSequence == nil {
		return nil
	}
	blockLabels := strings.Join(block.Labels, " ")

	// Check bottom meta-args come after all non-meta content
	foundPositionIssue, err := r.checkBottomMetaArgPositions(block, blockLabels, runner)
	if err != nil {
		return err
	}
	if foundPositionIssue {
		return nil
	}

	metaArgs := r.collectMetaArgumentsInLexOrder(block)
	if len(metaArgs) == 0 {
		return nil
	}
	return r.checkMetaArgSequence(metaArgs, desiredSequence, block, blockLabels, runner)
}

func (r *TerraformArgumentOrderRule) getBottomMetaArgs(blockType string) []string {
	switch blockType {
	case TypeResource:
		return []string{ArgLifecycle, ArgDependsOn}
	case TypeModule:
		return []string{ArgDependsOn}
	default:
		return nil
	}
}

func (r *TerraformArgumentOrderRule) checkBottomMetaArgPositions(block *hclsyntax.Block, blockLabels string, runner tflint.Runner) (bool, error) {
	bottomMetaArgs := r.getBottomMetaArgs(block.Type)
	if len(bottomMetaArgs) == 0 {
		return false, nil
	}

	bottomSet := make(map[string]bool, len(bottomMetaArgs))
	for _, name := range bottomMetaArgs {
		bottomSet[name] = true
	}

	var items []metaOrderItem

	for _, attr := range block.Body.Attributes {
		items = append(items, metaOrderItem{
			name:     attr.Name,
			startPos: attr.Range().Start.Byte,
			isBlock:  false,
			isBottom: bottomSet[attr.Name],
		})
	}
	for _, childBlock := range block.Body.Blocks {
		items = append(items, metaOrderItem{
			name:     childBlock.Type,
			startPos: childBlock.DefRange().Start.Byte,
			isBlock:  true,
			isBottom: bottomSet[childBlock.Type],
		})
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].startPos < items[j].startPos
	})

	// Find the maximum start position of any non-bottom item
	maxNonBottomPos := -1
	for _, it := range items {
		if !it.isBottom && it.startPos > maxNonBottomPos {
			maxNonBottomPos = it.startPos
		}
	}

	if maxNonBottomPos < 0 {
		return false, nil
	}

	// Report the first bottom meta-arg that appears before a non-bottom item
	for _, it := range items {
		if it.isBottom && it.startPos < maxNonBottomPos {
			return true, runner.EmitIssueWithFix(r,
				fmt.Sprintf(
					"Out-of-order meta argument '%s' in %s '%s': must appear after all %s arguments and blocks",
					it.name, block.Type, blockLabels, block.Type,
				),
				metaArgRange(block, it.name),
				func(f tflint.Fixer) error {
					return r.fixBottomMetaArgPositions(f, block, items)
				},
			)
		}
	}

	return false, nil
}

func (r *TerraformArgumentOrderRule) getDesiredSequence(blockType string) []string {
	switch blockType {
	case TypeResource:
		return []string{ArgProvider, ArgCount + "|" + ArgForEach, ArgLifecycle, ArgDependsOn}
	case TypeModule:
		return []string{ArgCount + "|" + ArgForEach, ArgDependsOn}
	default:
		return nil
	}
}

// collectMetaArgumentsInLexOrder collects meta-arguments (count, for_each, provider, depends_on, lifecycle)
// in the exact lexical order they appear in the block. This ensures we don’t incorrectly treat
// “depends_on” or others as out-of-order if they actually appear properly in the file.
func (r *TerraformArgumentOrderRule) collectMetaArgumentsInLexOrder(block *hclsyntax.Block) []string {
	type item struct {
		Name     string
		Type     string // "attr" or "block"
		StartIdx int    // position of the line/byte
	}

	var items []item

	for _, attr := range block.Body.Attributes {
		items = append(items, item{
			Name:     attr.Name,
			Type:     TypeAttr,
			StartIdx: attr.Range().Start.Byte,
		})
	}
	for _, childBlock := range block.Body.Blocks {
		items = append(items, item{
			Name:     childBlock.Type,
			Type:     TypeBlock,
			StartIdx: childBlock.DefRange().Start.Byte,
		})
	}

	// sort them by lexical StartIdx so we see them in the actual file order
	sort.Slice(items, func(i, j int) bool {
		return items[i].StartIdx < items[j].StartIdx
	})

	var metaArgs []string
	for _, i := range items {
		// only record recognized meta arguments
		switch i.Type {
		case TypeAttr:
			switch i.Name {
			case ArgCount, ArgForEach, ArgProvider, ArgDependsOn:
				metaArgs = append(metaArgs, i.Name)
			}
		case TypeBlock:
			// for blocks, only "lifecycle" is a meta-argument
			if i.Name == ArgLifecycle {
				metaArgs = append(metaArgs, i.Name)
			}
		}
	}
	return metaArgs
}

// checkCountOrForEach tries to see whether "count|for_each" is out of order,
// comparing foundIdx to lastIndex. Returns updated lastIndex or an error if out-of-order.
func (r *TerraformArgumentOrderRule) checkCountOrForEach(
	foundIdx, forEachIdx, lastIndex int,
	block *hclsyntax.Block, blockLabels string,
	desiredSequence []string, runner tflint.Runner,
) (int, error) {
	if foundIdx < lastIndex {
		// out-of-order
		argName := ArgCount
		if foundIdx == forEachIdx {
			argName = ArgForEach
		}
		return lastIndex, runner.EmitIssue(r,
			fmt.Sprintf(
				"Out-of-order meta argument '%s' in %s '%s'. Expected sequence: %s",
				argName, block.Type, blockLabels,
				strings.Join(desiredSequence, " -> "),
			),
			metaArgRange(block, argName),
		)
	}
	return foundIdx, nil
}

// checkSingleArg ensures that idx >= 0 appears after lastIndex, or else it’s out-of-order.
func (r *TerraformArgumentOrderRule) checkSingleArg(
	want string, idx, lastIndex int,
	block *hclsyntax.Block, blockLabels string,
	desiredSequence []string, runner tflint.Runner,
) (int, error) {
	if idx < lastIndex {
		return lastIndex, runner.EmitIssue(r,
			fmt.Sprintf(
				"Out-of-order meta argument '%s' in %s '%s'. Expected sequence: %s",
				want, block.Type, blockLabels,
				strings.Join(desiredSequence, " -> "),
			),
			metaArgRange(block, want),
		)
	}
	return idx, nil
}

func (r *TerraformArgumentOrderRule) checkMetaArgSequence(
	metaArgs, desiredSequence []string,
	block *hclsyntax.Block,
	blockLabels string,
	runner tflint.Runner,
) error {
	// 1) Build a dictionary mapping metaArg -> index in metaArgs
	metaArgIndices := make(map[string]int)
	for i, arg := range metaArgs {
		if _, already := metaArgIndices[arg]; !already {
			metaArgIndices[arg] = i // store first occurrence only
		}
	}

	// 2) We'll keep track of the highest index found so far
	lastIndex := -1

	// Helper to find the actual index of a meta-argument if it exists
	// returns -1 if not found
	getIndex := func(arg string) int {
		if i, ok := metaArgIndices[arg]; ok {
			return i
		}
		return -1
	}

	for i := 0; i < len(desiredSequence); i++ {
		want := desiredSequence[i]

		// If we are dealing with "count|for_each"
		if want == ArgCount+"|"+ArgForEach {
			countIdx := getIndex(ArgCount)
			forEachIdx := getIndex(ArgForEach)

			// If both are absent, skip
			if countIdx < 0 && forEachIdx < 0 {
				continue
			}
			// If both are present, pick whichever is earlier
			var foundIdx int
			switch {
			case countIdx >= 0 && forEachIdx >= 0:
				if countIdx < forEachIdx {
					foundIdx = countIdx
				} else {
					foundIdx = forEachIdx
				}
			case countIdx >= 0:
				foundIdx = countIdx
			default:
				foundIdx = forEachIdx
			}

			newIndex, err := r.checkCountOrForEach(foundIdx, forEachIdx, lastIndex, block, blockLabels, desiredSequence, runner)
			if err != nil {
				return err
			}
			lastIndex = newIndex
		} else {
			idx := getIndex(want)
			// not present -> skip
			if idx < 0 {
				continue
			}
			newIndex, err := r.checkSingleArg(want, idx, lastIndex, block, blockLabels, desiredSequence, runner)
			if err != nil {
				return err
			}
			lastIndex = newIndex
		}
	}
	return nil
}

// fixBottomMetaArgPositions reorders items so bottom meta-args appear after all other content
func (r *TerraformArgumentOrderRule) fixBottomMetaArgPositions(
	f tflint.Fixer,
	block *hclsyntax.Block,
	items []metaOrderItem,
) error {
	if strings.HasSuffix(block.DefRange().Filename, ".json") {
		return tflint.ErrFixNotSupported
	}

	attrTexts, blockTexts := r.extractMetaOrderItemTexts(f, block, items)

	// Partition: non-bottom (preserve file order), bottom (fixed order)
	var nonBottom, bottom []metaOrderItem
	for _, it := range items {
		if it.isBottom {
			bottom = append(bottom, it)
		} else {
			nonBottom = append(nonBottom, it)
		}
	}

	// Sort bottom by fixed order (lifecycle before depends_on)
	bottomMetaArgs := r.getBottomMetaArgs(block.Type)
	bottomOrder := make(map[string]int, len(bottomMetaArgs))
	for i, name := range bottomMetaArgs {
		bottomOrder[name] = i
	}
	sort.Slice(bottom, func(i, j int) bool {
		return bottomOrder[bottom[i].name] < bottomOrder[bottom[j].name]
	})

	ordered := make([]metaOrderItem, 0, len(items))
	ordered = append(ordered, nonBottom...)
	ordered = append(ordered, bottom...)

	var result strings.Builder
	writeMetaOrderBlockOpening(&result, block)
	writeMetaOrderItems(&result, ordered, attrTexts, blockTexts)
	result.WriteString("\n}")

	fullBlockRange := hcl.Range{
		Filename: block.DefRange().Filename,
		Start:    block.DefRange().Start,
		End:      block.Body.Range().End,
	}

	return f.ReplaceText(fullBlockRange, result.String())
}

// extractMetaOrderItemTexts extracts text for all attributes and blocks in a resource/module block
func (r *TerraformArgumentOrderRule) extractMetaOrderItemTexts(
	f tflint.Fixer,
	block *hclsyntax.Block,
	items []metaOrderItem,
) (map[string]string, map[int]string) {
	attrTexts := make(map[string]string)
	blockTexts := make(map[int]string)

	attrSet := make(map[string]bool)
	blockSet := make(map[int]bool)
	for _, it := range items {
		if it.isBlock {
			blockSet[it.startPos] = true
		} else {
			attrSet[it.name] = true
		}
	}

	for _, attr := range block.Body.Attributes {
		if attrSet[attr.Name] {
			attrTexts[attr.Name] = string(f.TextAt(attr.Range()).Bytes)
		}
	}

	for _, blk := range block.Body.Blocks {
		startPos := blk.DefRange().Start.Byte
		if blockSet[startPos] {
			blockRange := hcl.Range{
				Filename: blk.DefRange().Filename,
				Start:    blk.DefRange().Start,
				End:      blk.Body.Range().End,
			}
			blockTexts[startPos] = string(f.TextAt(blockRange).Bytes)
		}
	}

	return attrTexts, blockTexts
}

func writeMetaOrderBlockOpening(result *strings.Builder, block *hclsyntax.Block) {
	result.WriteString(block.Type)
	for _, label := range block.Labels {
		result.WriteString(` "`)
		result.WriteString(label)
		result.WriteString(`"`)
	}
	result.WriteString(" {\n")
}

func writeMetaOrderItems(
	result *strings.Builder,
	items []metaOrderItem,
	attrTexts map[string]string,
	blockTexts map[int]string,
) {
	for i, item := range items {
		if i > 0 {
			prevIsBlock := items[i-1].isBlock
			needsBlankLine := item.isBlock || prevIsBlock || item.isBottom
			if needsBlankLine {
				result.WriteString("\n")
			}
			result.WriteString("\n")
		}

		if item.isBlock {
			writeMetaOrderBlock(result, blockTexts[item.startPos])
		} else {
			writeMetaOrderAttribute(result, attrTexts[item.name])
		}
	}
}

func writeMetaOrderAttribute(result *strings.Builder, text string) {
	result.WriteString("  ")
	result.WriteString(text)
}

func writeMetaOrderBlock(result *strings.Builder, text string) {
	lines := strings.Split(text, "\n")
	for j, line := range lines {
		if j > 0 {
			result.WriteString("\n")
		}
		if line != "" {
			result.WriteString("  ")
			result.WriteString(line)
		}
	}
}

func metaArgRange(block *hclsyntax.Block, argName string) hcl.Range {
	for _, attr := range block.Body.Attributes {
		if attr.Name == argName {
			return attr.Range()
		}
	}
	for _, childBlock := range block.Body.Blocks {
		if childBlock.Type == argName {
			return childBlock.DefRange()
		}
	}
	return block.DefRange()
}
