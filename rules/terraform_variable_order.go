package rules

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// variableBlock represents a variable block with its ordering metadata.
type variableBlock struct {
	Name       string
	HasDefault bool
	Range      hcl.Range // Full range of the block
	DefRange   hcl.Range // Definition range for error reporting
	Start      int
}

// TerraformVariableOrderRule checks that variables are ordered as:
// 1) Required variables (no default) in alphabetical order
// 2) Optional variables (has default) in alphabetical order
type TerraformVariableOrderRule struct {
	tflint.DefaultRule
}

// NewTerraformVariableOrderRule creates a new rule instance.
func NewTerraformVariableOrderRule() *TerraformVariableOrderRule {
	return &TerraformVariableOrderRule{}
}

// Name returns the rule name.
func (r *TerraformVariableOrderRule) Name() string {
	return "terraform_variable_order"
}

// Enabled returns whether the rule is enabled by default.
func (r *TerraformVariableOrderRule) Enabled() bool {
	return true
}

// Severity returns the severity of the rule.
func (r *TerraformVariableOrderRule) Severity() tflint.Severity {
	return tflint.ERROR
}

// Link returns the rule's reference link.
func (r *TerraformVariableOrderRule) Link() string {
	return GetRuleDocLink(r.Name())
}

// Check checks the order of variable blocks.
func (r *TerraformVariableOrderRule) Check(runner tflint.Runner) error {
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
			// Skip parse errors
			continue
		}
		if body, ok := syntaxFile.Body.(*hclsyntax.Body); ok {
			if err := r.processFile(body, filename, runner); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *TerraformVariableOrderRule) processFile(body *hclsyntax.Body, filename string, runner tflint.Runner) error {
	// Collect all variable blocks at the top level
	var varBlocks []variableBlock

	for _, blk := range body.Blocks {
		// Block types are always lowercase in Terraform
		if blk.Type == TypeVariable && len(blk.Labels) > 0 {
			vName := blk.Labels[0]

			// Check whether "default" attribute is present
			_, hasDefault := blk.Body.Attributes["default"]

			// Calculate full range from start of block to end of body
			fullRange := hcl.Range{
				Filename: filename,
				Start:    blk.DefRange().Start,
				End:      blk.Body.Range().End,
			}

			varBlocks = append(varBlocks, variableBlock{
				Name:       vName,
				HasDefault: hasDefault,
				Range:      fullRange,
				DefRange:   blk.DefRange(),
				Start:      blk.DefRange().Start.Byte,
			})
		}
		// Recurse into nested blocks to maintain backward compatibility
		if err := r.processFile(blk.Body, filename, runner); err != nil {
			return err
		}
	}

	// If no variables found here, nothing to check
	if len(varBlocks) == 0 {
		return nil
	}

	// Sort varBlocks by their starting position
	sort.Slice(varBlocks, func(i, j int) bool {
		return varBlocks[i].Start < varBlocks[j].Start
	})

	// Check if the order is correct
	if isCorrectOrder(varBlocks) {
		return nil
	}

	// Emit issue with autofix
	return r.emitIssueWithFix(runner, varBlocks, filename)
}

// isCorrectOrder checks if the variable blocks are in the correct order
func isCorrectOrder(varBlocks []variableBlock) bool {
	lastRequiredName := ""
	lastOptionalName := ""
	seenOptional := false

	for _, vb := range varBlocks {
		if !vb.HasDefault {
			// Required variable: check if we've seen optional or if out of order
			if seenOptional || (lastRequiredName != "" && vb.Name < lastRequiredName) {
				return false
			}
			lastRequiredName = vb.Name
		} else {
			// Optional variable: check alphabetical order
			if lastOptionalName != "" && vb.Name < lastOptionalName {
				return false
			}
			lastOptionalName = vb.Name
			seenOptional = true
		}
	}
	return true
}

// emitIssueWithFix emits an issue with autofix support
func (r *TerraformVariableOrderRule) emitIssueWithFix(
	runner tflint.Runner,
	varBlocks []variableBlock,
	filename string,
) error {
	// Find the first variable that's out of order for the error message and location
	var outOfOrderVar string
	var outOfOrderRange hcl.Range
	lastRequiredName := ""
	lastOptionalName := ""
	seenOptional := false

	for _, vb := range varBlocks {
		if !vb.HasDefault {
			if seenOptional || (lastRequiredName != "" && vb.Name < lastRequiredName) {
				outOfOrderVar = vb.Name
				outOfOrderRange = vb.DefRange
				break
			}
			lastRequiredName = vb.Name
		} else {
			if lastOptionalName != "" && vb.Name < lastOptionalName {
				outOfOrderVar = vb.Name
				outOfOrderRange = vb.DefRange
				break
			}
			lastOptionalName = vb.Name
			seenOptional = true
		}
	}

	msg := fmt.Sprintf(
		`Out-of-order variable %q. Required variables must come first in alphabetical order, followed by optional variables in alphabetical order.`,
		outOfOrderVar,
	)

	// Use the out-of-order variable's range for the issue location
	return runner.EmitIssueWithFix(r, msg, outOfOrderRange, func(f tflint.Fixer) error {
		// Get the text content of all variable blocks
		type varBlockWithContent struct {
			variableBlock
			Content string
		}

		var blocksWithContent []varBlockWithContent
		for _, vb := range varBlocks {
			text := f.TextAt(vb.Range)
			blocksWithContent = append(blocksWithContent, varBlockWithContent{
				variableBlock: vb,
				Content:       string(text.Bytes),
			})
		}

		// Sort variables: required first (alphabetical), then optional (alphabetical)
		sort.Slice(blocksWithContent, func(i, j int) bool {
			if blocksWithContent[i].HasDefault != blocksWithContent[j].HasDefault {
				return !blocksWithContent[i].HasDefault // required (!HasDefault) comes first
			}
			return blocksWithContent[i].Name < blocksWithContent[j].Name
		})

		// Build the fixed content preserving original spacing
		var fixedContent strings.Builder

		// Create a map to track original spacing between consecutive variables
		spacingMap := make(map[string]string)
		for i := 1; i < len(varBlocks); i++ {
			betweenRange := hcl.Range{
				Filename: filename,
				Start:    varBlocks[i-1].Range.End,
				End:      varBlocks[i].Range.Start,
			}
			betweenText := f.TextAt(betweenRange)
			key := varBlocks[i-1].Name + "|||" + varBlocks[i].Name
			spacingMap[key] = string(betweenText.Bytes)
		}

		for i, vb := range blocksWithContent {
			if i > 0 {
				// Try to find original spacing between these two variables
				prevName := blocksWithContent[i-1].Name
				currName := vb.Name

				// Check both orderings since they might have been reordered
				spacing := ""
				if s, ok := spacingMap[prevName+"|||"+currName]; ok {
					spacing = s
				} else if s, ok := spacingMap[currName+"|||"+prevName]; ok {
					spacing = s
				} else {
					// Default to double newline if they weren't originally adjacent
					spacing = "\n\n"
				}

				fixedContent.WriteString(spacing)
			}
			fixedContent.WriteString(vb.Content)
		}

		// Replace the entire range from first to last variable
		fullRange := hcl.Range{
			Filename: varBlocks[0].Range.Filename,
			Start:    varBlocks[0].Range.Start,
			End:      varBlocks[len(varBlocks)-1].Range.End,
		}

		return f.ReplaceText(fullRange, fixedContent.String())
	})
}
