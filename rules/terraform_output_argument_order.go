package rules

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// TerraformOutputArgumentOrderRule checks that output blocks follow the order:
// description, value, ephemeral, sensitive, precondition, depends_on
type TerraformOutputArgumentOrderRule struct {
	tflint.DefaultRule
}

// outputArgumentItem represents an attribute or block within an output block
type outputArgumentItem struct {
	Name    string
	Index   int
	Range   hcl.Range
	Start   int
	IsBlock bool
}

func NewTerraformOutputArgumentOrderRule() *TerraformOutputArgumentOrderRule {
	return &TerraformOutputArgumentOrderRule{}
}

func (r *TerraformOutputArgumentOrderRule) Name() string {
	return "terraform_output_argument_order"
}

func (r *TerraformOutputArgumentOrderRule) Enabled() bool {
	return true
}

func (r *TerraformOutputArgumentOrderRule) Severity() tflint.Severity {
	return tflint.ERROR
}

func (r *TerraformOutputArgumentOrderRule) Link() string {
	return GetRuleDocLink(r.Name())
}

func (r *TerraformOutputArgumentOrderRule) Check(runner tflint.Runner) error {
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
			// skip parse errors
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

func (r *TerraformOutputArgumentOrderRule) processBody(body *hclsyntax.Body, runner tflint.Runner) error {
	for _, block := range body.Blocks {
		if strings.EqualFold(block.Type, "output") {
			if err := r.checkOutputBlock(block, runner); err != nil {
				return err
			}
		}
		// recurse for nested blocks
		if err := r.processBody(block.Body, runner); err != nil {
			return err
		}
	}
	return nil
}

// checkOutputBlock enforces the order description -> value -> ephemeral -> sensitive -> precondition -> depends_on
func (r *TerraformOutputArgumentOrderRule) checkOutputBlock(
	block *hclsyntax.Block,
	runner tflint.Runner,
) error {
	orderMap := map[string]int{
		"description":  0,
		"value":        1,
		"ephemeral":    2,
		"sensitive":    3,
		"precondition": 4,
		"depends_on":   5,
	}

	var items []outputArgumentItem

	// Gather recognized attributes (attribute names are always lowercase in Terraform)
	for _, attr := range block.Body.Attributes {
		idx, found := orderMap[attr.Name]
		if found && attr.Name != TypePrecondition { // precondition is a block, not an attribute
			items = append(items, outputArgumentItem{
				Name:    attr.Name,
				Index:   idx,
				Range:   attr.Range(),
				Start:   attr.Range().Start.Byte,
				IsBlock: false,
			})
		}
	}

	// Gather precondition blocks (block types are always lowercase in Terraform)
	for _, blk := range block.Body.Blocks {
		if blk.Type == TypePrecondition {
			idx := orderMap[TypePrecondition]
			items = append(items, outputArgumentItem{
				Name:    TypePrecondition,
				Index:   idx,
				Range:   blk.DefRange(),
				Start:   blk.DefRange().Start.Byte,
				IsBlock: true,
			})
		}
	}

	// If nothing recognized, do nothing
	if len(items) == 0 {
		return nil
	}

	// Sort by their position in the file
	sort.Slice(items, func(i, j int) bool {
		return items[i].Start < items[j].Start
	})

	lastIndex := -1
	var outOfOrderItem *outputArgumentItem

	for i := range items {
		if items[i].Index < lastIndex {
			outOfOrderItem = &items[i]
			break
		}
		lastIndex = items[i].Index
	}

	if outOfOrderItem != nil {
		msg := fmt.Sprintf("Out-of-order argument '%s'. Expected sequence: description, value, ephemeral, sensitive, precondition, depends_on", outOfOrderItem.Name)
		return runner.EmitIssueWithFix(r, msg, outOfOrderItem.Range, func(f tflint.Fixer) error {
			return r.fixOutputArgumentOrder(f, block, items)
		})
	}

	return nil
}

// fixOutputArgumentOrder reorders the arguments within an output block
func (r *TerraformOutputArgumentOrderRule) fixOutputArgumentOrder(
	f tflint.Fixer,
	block *hclsyntax.Block,
	items []outputArgumentItem,
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
func (r *TerraformOutputArgumentOrderRule) sortItemsByExpectedOrder(items []outputArgumentItem) []outputArgumentItem {
	orderedItems := make([]outputArgumentItem, len(items))
	copy(orderedItems, items)
	sort.Slice(orderedItems, func(i, j int) bool {
		return orderedItems[i].Index < orderedItems[j].Index
	})
	return orderedItems
}

// isAlreadyOrdered checks if items are already in the correct order
func (r *TerraformOutputArgumentOrderRule) isAlreadyOrdered(items, orderedItems []outputArgumentItem) bool {
	for i := range items {
		if items[i].Name != orderedItems[i].Name {
			return false
		}
	}
	return true
}

// extractItemTexts extracts text content for attributes and blocks
func (r *TerraformOutputArgumentOrderRule) extractItemTexts(
	f tflint.Fixer,
	block *hclsyntax.Block,
	items []outputArgumentItem,
) (map[string]string, map[string]string) {
	attrTexts := make(map[string]string)
	blockTexts := make(map[string]string)

	// Get the text for each attribute (attribute names are always lowercase in Terraform)
	for _, attr := range block.Body.Attributes {
		// Check if this is one of our tracked attributes
		for _, item := range items {
			if item.Name == attr.Name && !item.IsBlock {
				attrRange := attr.Range()
				text := f.TextAt(attrRange)
				attrTexts[attr.Name] = string(text.Bytes)
				break
			}
		}
	}

	// Get the text for precondition blocks (block types are always lowercase in Terraform)
	for _, blk := range block.Body.Blocks {
		if blk.Type == TypePrecondition {
			blockRange := hcl.Range{
				Filename: blk.DefRange().Filename,
				Start:    blk.DefRange().Start,
				End:      blk.Body.Range().End,
			}
			text := f.TextAt(blockRange)
			blockTexts[TypePrecondition] = string(text.Bytes)
		}
	}

	return attrTexts, blockTexts
}

// applyReorderedContent builds and applies the reordered content
func (r *TerraformOutputArgumentOrderRule) applyReorderedContent(
	f tflint.Fixer,
	block *hclsyntax.Block,
	orderedItems []outputArgumentItem,
	attrTexts, blockTexts map[string]string,
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
func (r *TerraformOutputArgumentOrderRule) writeBlockOpening(result *strings.Builder, block *hclsyntax.Block) {
	result.WriteString("output ")
	if len(block.Labels) > 0 {
		result.WriteString(`"`)
		result.WriteString(block.Labels[0])
		result.WriteString(`" `)
	}
	result.WriteString("{\n")
}

// writeOrderedItems writes the items in the correct order with proper formatting
func (r *TerraformOutputArgumentOrderRule) writeOrderedItems(
	result *strings.Builder,
	orderedItems []outputArgumentItem,
	attrTexts, blockTexts map[string]string,
) {
	for i, orderedItem := range orderedItems {
		if i > 0 {
			// Add extra blank line before precondition and depends_on for better formatting
			if orderedItem.Name == TypePrecondition || orderedItem.Name == ArgDependsOn {
				result.WriteString("\n")
			}
			result.WriteString("\n")
		}

		if orderedItem.IsBlock {
			r.writeBlock(result, orderedItem.Name, blockTexts[orderedItem.Name])
		} else {
			r.writeAttribute(result, orderedItem.Name, attrTexts[orderedItem.Name])
		}
	}
}

// writeBlock writes a block with proper indentation
func (r *TerraformOutputArgumentOrderRule) writeBlock(result *strings.Builder, _, text string) {
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
func (r *TerraformOutputArgumentOrderRule) writeAttribute(result *strings.Builder, _, text string) {
	result.WriteString("  ")
	result.WriteString(text)
}
