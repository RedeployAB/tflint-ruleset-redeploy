package rules

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// TerraformVariableArgumentOrderRule enforces argument order for variable blocks:
//
//	description, type, default, ephemeral, sensitive, nullable, validation
//
// Any of these may be omitted, but if present, must follow that sequence.
// For validation blocks, multiple occurrences are allowed, but all must appear after the others.
type TerraformVariableArgumentOrderRule struct {
	tflint.DefaultRule
}

// variableArgumentItem represents an attribute or block within a variable block
type variableArgumentItem struct {
	Name  string
	Index int
	Range hcl.Range
	Start int
	IsBlk bool
}

func NewTerraformVariableArgumentOrderRule() *TerraformVariableArgumentOrderRule {
	return &TerraformVariableArgumentOrderRule{}
}

func (r *TerraformVariableArgumentOrderRule) Name() string {
	return "terraform_variable_argument_order"
}

func (r *TerraformVariableArgumentOrderRule) Enabled() bool {
	return true
}

func (r *TerraformVariableArgumentOrderRule) Severity() tflint.Severity {
	return tflint.ERROR
}

func (r *TerraformVariableArgumentOrderRule) Link() string {
	return GetRuleDocLink(r.Name())
}

func (r *TerraformVariableArgumentOrderRule) Check(runner tflint.Runner) error {
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

func (r *TerraformVariableArgumentOrderRule) processBody(body *hclsyntax.Body, runner tflint.Runner) error {
	for _, block := range body.Blocks {
		// Only examine variable blocks (block types are always lowercase in Terraform)
		if block.Type == TypeVariable {
			if err := r.checkVariableBlock(block, runner); err != nil {
				return err
			}
		}
		// Recurse into nested blocks
		if err := r.processBody(block.Body, runner); err != nil {
			return err
		}
	}
	return nil
}

func (r *TerraformVariableArgumentOrderRule) checkVariableBlock(
	block *hclsyntax.Block,
	runner tflint.Runner,
) error {
	// Recognized order for attributes/blocks:
	// description(0), type(1), default(2), ephemeral(3), sensitive(4), nullable(5), validation(6)
	// Any of these may be omitted, but if present, must follow that sequence.
	// For validation blocks, multiple occurrences are allowed, but all must appear after the others.

	// Define the expected order
	orderMap := map[string]int{
		"description": 0,
		"type":        1,
		"default":     2,
		"ephemeral":   3,
		"sensitive":   4,
		"nullable":    5,
		"validation":  6,
	}

	var items []variableArgumentItem

	// Gather recognized attributes (attribute names are always lowercase in Terraform)
	for _, attr := range block.Body.Attributes {
		idx, found := orderMap[attr.Name]
		if found {
			items = append(items, variableArgumentItem{
				Name:  attr.Name,
				Index: idx,
				Range: attr.Range(),
				Start: attr.Range().Start.Byte,
				IsBlk: false,
			})
		}
	}

	// Gather recognized blocks: "validation" (block types are always lowercase in Terraform)
	for _, childBlock := range block.Body.Blocks {
		if childBlock.Type == TypeValidation {
			items = append(items, variableArgumentItem{
				Name:  childBlock.Type,
				Index: orderMap[childBlock.Type], // 6
				Range: childBlock.DefRange(),
				Start: childBlock.DefRange().Start.Byte,
				IsBlk: true,
			})
		}
	}

	// If no recognized items => no check needed
	if len(items) == 0 {
		return nil
	}

	// Sort items by lexical file order
	sort.Slice(items, func(i, j int) bool {
		return items[i].Start < items[j].Start
	})

	// Track the highest index encountered so far
	lastIndex := -1
	var outOfOrderItem *variableArgumentItem

	for i := range items {
		if items[i].Index < lastIndex {
			// Out-of-order argument found
			outOfOrderItem = &items[i]
			break
		}
		lastIndex = items[i].Index
	}

	if outOfOrderItem != nil {
		msg := fmt.Sprintf(
			"Out-of-order argument '%s'. Expected sequence: description, type, default, ephemeral, sensitive, nullable, validation",
			outOfOrderItem.Name,
		)
		return runner.EmitIssueWithFix(r, msg, outOfOrderItem.Range, func(f tflint.Fixer) error {
			return r.fixVariableArgumentOrder(f, block, items)
		})
	}
	return nil
}

// fixVariableArgumentOrder reorders the arguments within a variable block
func (r *TerraformVariableArgumentOrderRule) fixVariableArgumentOrder(
	f tflint.Fixer,
	block *hclsyntax.Block,
	items []variableArgumentItem,
) error {
	// For JSON files, we can't reliably preserve formatting
	if strings.HasSuffix(block.DefRange().Filename, ".json") {
		return tflint.ErrFixNotSupported
	}

	// Sort items by their expected order
	orderedItems := r.sortItemsByExpectedOrder(items)

	// Check if already in correct order
	if r.isAlreadyOrdered(items, orderedItems) {
		return nil
	}

	// Extract text content for all items
	attrTexts, blockTexts := r.extractItemTexts(f, block, items)

	// Build and apply the reordered content
	return r.applyReorderedContent(f, block, orderedItems, attrTexts, blockTexts)
}

// sortItemsByExpectedOrder sorts items by their expected order index
func (r *TerraformVariableArgumentOrderRule) sortItemsByExpectedOrder(items []variableArgumentItem) []variableArgumentItem {
	orderedItems := make([]variableArgumentItem, len(items))
	copy(orderedItems, items)
	sort.Slice(orderedItems, func(i, j int) bool {
		if orderedItems[i].Index != orderedItems[j].Index {
			return orderedItems[i].Index < orderedItems[j].Index
		}
		// For same index (multiple validation blocks), keep original order
		return orderedItems[i].Start < orderedItems[j].Start
	})
	return orderedItems
}

// isAlreadyOrdered checks if items are already in the correct order
func (r *TerraformVariableArgumentOrderRule) isAlreadyOrdered(items, orderedItems []variableArgumentItem) bool {
	for i := range items {
		if items[i].Name != orderedItems[i].Name {
			return false
		}
	}
	return true
}

// extractItemTexts extracts text content for attributes and blocks
func (r *TerraformVariableArgumentOrderRule) extractItemTexts(
	f tflint.Fixer,
	block *hclsyntax.Block,
	items []variableArgumentItem,
) (map[string]string, map[int]string) {
	attrTexts := make(map[string]string)
	blockTexts := make(map[int]string) // Use start position as key for validation blocks

	// Extract attribute texts
	r.extractAttributeTexts(f, block, items, attrTexts)

	// Extract validation block texts
	r.extractValidationBlockTexts(f, block, items, blockTexts)

	return attrTexts, blockTexts
}

// extractAttributeTexts extracts text for all attributes
func (r *TerraformVariableArgumentOrderRule) extractAttributeTexts(
	f tflint.Fixer,
	block *hclsyntax.Block,
	items []variableArgumentItem,
	attrTexts map[string]string,
) {
	for _, attr := range block.Body.Attributes {
		// Check if this is one of our tracked attributes (attribute names are always lowercase in Terraform)
		for _, item := range items {
			if item.Name == attr.Name && !item.IsBlk {
				attrRange := attr.Range()
				text := f.TextAt(attrRange)
				attrTexts[attr.Name] = string(text.Bytes)
				break
			}
		}
	}
}

// extractValidationBlockTexts extracts text for validation blocks
func (r *TerraformVariableArgumentOrderRule) extractValidationBlockTexts(
	f tflint.Fixer,
	block *hclsyntax.Block,
	items []variableArgumentItem,
	blockTexts map[int]string,
) {
	for _, blk := range block.Body.Blocks {
		// Block types are always lowercase in Terraform
		if blk.Type == TypeValidation {
			// Find the corresponding item by start position
			for _, item := range items {
				if item.IsBlk && item.Start == blk.DefRange().Start.Byte {
					blockRange := hcl.Range{
						Filename: blk.DefRange().Filename,
						Start:    blk.DefRange().Start,
						End:      blk.Body.Range().End,
					}
					text := f.TextAt(blockRange)
					blockTexts[item.Start] = string(text.Bytes)
					break
				}
			}
		}
	}
}

// applyReorderedContent builds and applies the reordered content
func (r *TerraformVariableArgumentOrderRule) applyReorderedContent(
	f tflint.Fixer,
	block *hclsyntax.Block,
	orderedItems []variableArgumentItem,
	attrTexts map[string]string,
	blockTexts map[int]string,
) error {
	var result strings.Builder

	// Write the opening line
	r.writeBlockOpening(&result, block)

	// Write attributes and blocks in the correct order
	r.writeOrderedItems(&result, orderedItems, attrTexts, blockTexts)

	result.WriteString("\n}")

	// Replace the entire block
	fullBlockRange := hcl.Range{
		Filename: block.DefRange().Filename,
		Start:    block.DefRange().Start,
		End:      block.Body.Range().End,
	}

	return f.ReplaceText(fullBlockRange, result.String())
}

// writeBlockOpening writes the opening line of the block
func (r *TerraformVariableArgumentOrderRule) writeBlockOpening(result *strings.Builder, block *hclsyntax.Block) {
	result.WriteString("variable ")
	if len(block.Labels) > 0 {
		result.WriteString(`"`)
		result.WriteString(block.Labels[0])
		result.WriteString(`" `)
	}
	result.WriteString("{\n")
}

// writeOrderedItems writes the items in the correct order with proper formatting
func (r *TerraformVariableArgumentOrderRule) writeOrderedItems(
	result *strings.Builder,
	orderedItems []variableArgumentItem,
	attrTexts map[string]string,
	blockTexts map[int]string,
) {
	for i, orderedItem := range orderedItems {
		if i > 0 {
			// Add extra blank line before validation blocks
			if orderedItem.IsBlk && (i == 0 || !orderedItems[i-1].IsBlk) {
				result.WriteString("\n")
			}
			result.WriteString("\n")
		}

		if orderedItem.IsBlk {
			r.writeBlock(result, blockTexts[orderedItem.Start])
		} else {
			r.writeAttribute(result, attrTexts[orderedItem.Name])
		}
	}
}

// writeBlock writes a block with proper indentation
func (r *TerraformVariableArgumentOrderRule) writeBlock(result *strings.Builder, text string) {
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

// writeAttribute writes an attribute with proper indentation
func (r *TerraformVariableArgumentOrderRule) writeAttribute(result *strings.Builder, text string) {
	result.WriteString("  ")
	result.WriteString(text)
}
