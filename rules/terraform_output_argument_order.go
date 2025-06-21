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
	Name     string
	Index    int
	Range    hcl.Range
	Start    int
	IsBlock  bool
	FullText string // Store the full text for blocks
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

	// Gather recognized attributes
	for _, attr := range block.Body.Attributes {
		lcName := strings.ToLower(attr.Name)
		idx, found := orderMap[lcName]
		if found && lcName != "precondition" { // precondition is a block, not an attribute
			items = append(items, outputArgumentItem{
				Name:    lcName,
				Index:   idx,
				Range:   attr.Range(),
				Start:   attr.Range().Start.Byte,
				IsBlock: false,
			})
		}
	}

	// Gather precondition blocks
	for _, blk := range block.Body.Blocks {
		if strings.ToLower(blk.Type) == "precondition" {
			idx := orderMap["precondition"]
			items = append(items, outputArgumentItem{
				Name:    "precondition",
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
	orderedItems := make([]outputArgumentItem, len(items))
	copy(orderedItems, items)
	sort.Slice(orderedItems, func(i, j int) bool {
		return orderedItems[i].Index < orderedItems[j].Index
	})

	// Check if already in correct order
	alreadyOrdered := true
	for i := range items {
		if items[i].Name != orderedItems[i].Name {
			alreadyOrdered = false
			break
		}
	}
	if alreadyOrdered {
		return nil
	}

	// Create a map to store attribute text content
	attrTexts := make(map[string]string)
	blockTexts := make(map[string]string)

	// Get the text for each attribute
	for _, attr := range block.Body.Attributes {
		lcName := strings.ToLower(attr.Name)
		// Check if this is one of our tracked attributes
		for _, item := range items {
			if item.Name == lcName && !item.IsBlock {
				// Get the range from the start of the attribute name to the end of the value
				// This includes the entire line like "description = "some desc""
				attrRange := attr.Range()
				text := f.TextAt(attrRange)
				attrTexts[lcName] = string(text.Bytes)
				break
			}
		}
	}

	// Get the text for precondition blocks
	for _, blk := range block.Body.Blocks {
		if strings.ToLower(blk.Type) == "precondition" {
			// Get the full block range
			blockRange := hcl.Range{
				Filename: blk.DefRange().Filename,
				Start:    blk.DefRange().Start,
				End:      blk.Body.Range().End,
			}
			text := f.TextAt(blockRange)
			blockTexts["precondition"] = string(text.Bytes)
		}
	}

	// Build the reordered content
	var result strings.Builder

	// Write the opening line
	result.WriteString("output ")
	if len(block.Labels) > 0 {
		result.WriteString(`"`)
		result.WriteString(block.Labels[0])
		result.WriteString(`" `)
	}
	result.WriteString("{\n")

	// Write attributes and blocks in the correct order
	for i, orderedItem := range orderedItems {
		if i > 0 {
			// Add extra blank line before precondition and depends_on for better formatting
			if orderedItem.Name == "precondition" || orderedItem.Name == "depends_on" {
				// Always add blank line before these special items
				result.WriteString("\n")
			}
			result.WriteString("\n")
		}

		if orderedItem.IsBlock {
			// Add proper indentation for blocks
			lines := strings.Split(blockTexts[orderedItem.Name], "\n")
			for j, line := range lines {
				if j > 0 {
					result.WriteString("\n")
				}
				if line != "" {
					result.WriteString("  ")
					result.WriteString(line)
				}
			}
		} else {
			// Add proper indentation for attributes
			result.WriteString("  ")
			// Add the attribute text
			if text, ok := attrTexts[orderedItem.Name]; ok {
				result.WriteString(text)
			}
		}
	}

	result.WriteString("\n}")

	// Replace the entire block
	fullBlockRange := hcl.Range{
		Filename: block.DefRange().Filename,
		Start:    block.DefRange().Start,
		End:      block.Body.Range().End,
	}

	return f.ReplaceText(fullBlockRange, result.String())
}
