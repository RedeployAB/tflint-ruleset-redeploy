package rules

import (
	"cmp"
	"fmt"
	"slices"
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
func (*TerraformVariableOrderRule) Name() string {
	return "terraform_variable_order"
}

// Enabled returns whether the rule is enabled by default.
func (*TerraformVariableOrderRule) Enabled() bool {
	return true
}

// Severity returns the severity of the rule.
func (*TerraformVariableOrderRule) Severity() tflint.Severity {
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
			_, hasDefault := blk.Body.Attributes[ArgDefault]

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
	slices.SortFunc(varBlocks, func(a, b variableBlock) int {
		return cmp.Compare(a.Start, b.Start)
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
	_, found := findFirstOutOfOrderVariable(varBlocks)
	return !found
}

// findFirstOutOfOrderVariable returns the first variable block (in file order)
// that violates the ordering, and whether one was found.
func findFirstOutOfOrderVariable(varBlocks []variableBlock) (outOfOrder variableBlock, found bool) {
	lastRequiredName := ""
	lastOptionalName := ""
	seenOptional := false

	for _, vb := range varBlocks {
		if !vb.HasDefault {
			// Required variable: check if we've seen optional or if out of order
			if seenOptional || (lastRequiredName != "" && vb.Name < lastRequiredName) {
				return vb, true
			}
			lastRequiredName = vb.Name
			continue
		}
		// Optional variable: check alphabetical order
		if lastOptionalName != "" && vb.Name < lastOptionalName {
			return vb, true
		}
		lastOptionalName = vb.Name
		seenOptional = true
	}
	return variableBlock{}, false
}

// emitIssueWithFix emits an issue with autofix support
func (r *TerraformVariableOrderRule) emitIssueWithFix(
	runner tflint.Runner,
	varBlocks []variableBlock,
	filename string,
) error {
	// Find the first variable that's out of order for the error message and location
	outOfOrder, _ := findFirstOutOfOrderVariable(varBlocks)

	msg := fmt.Sprintf(
		`Out-of-order variable %q. Required variables must come first in alphabetical order, `+
			`followed by optional variables in alphabetical order.`,
		outOfOrder.Name,
	)

	// Use the out-of-order variable's range for the issue location
	return runner.EmitIssueWithFix(r, msg, outOfOrder.DefRange, func(f tflint.Fixer) error {
		return fixVariableOrder(f, varBlocks, filename)
	})
}

// varBlockWithContent pairs a variable block with its source text.
type varBlockWithContent struct {
	variableBlock
	Content string
}

// compareVariableBlocksForFix orders required variables (no default) before
// optional ones, each group alphabetically by name.
func compareVariableBlocksForFix(a, b varBlockWithContent) int {
	if a.HasDefault != b.HasDefault {
		if a.HasDefault {
			return 1
		}
		return -1 // required (!HasDefault) comes first
	}
	return strings.Compare(a.Name, b.Name)
}

// fixVariableOrder rewrites the variable blocks in the expected order while
// preserving the original spacing between originally adjacent variables.
func fixVariableOrder(f tflint.Fixer, varBlocks []variableBlock, filename string) error {
	// Get the text content of all variable blocks
	blocksWithContent := make([]varBlockWithContent, 0, len(varBlocks))
	for _, vb := range varBlocks {
		text := f.TextAt(vb.Range)
		blocksWithContent = append(blocksWithContent, varBlockWithContent{
			variableBlock: vb,
			Content:       string(text.Bytes),
		})
	}

	// Sort variables: required first (alphabetical), then optional (alphabetical)
	slices.SortFunc(blocksWithContent, compareVariableBlocksForFix)

	spacingMap := buildVariableSpacingMap(f, varBlocks, filename)

	// Build the fixed content preserving original spacing
	var fixedContent strings.Builder
	for i, vb := range blocksWithContent {
		if i > 0 {
			fixedContent.WriteString(lookupVariableSpacing(spacingMap, blocksWithContent[i-1].Name, vb.Name))
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
}

// buildVariableSpacingMap records the original text between each pair of
// consecutive variable blocks, keyed by "prev|||curr".
func buildVariableSpacingMap(f tflint.Fixer, varBlocks []variableBlock, filename string) map[string]string {
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
	return spacingMap
}

// lookupVariableSpacing returns the original spacing between two variables.
// Both orderings are checked since they might have been reordered; it
// defaults to a double newline if they weren't originally adjacent.
func lookupVariableSpacing(spacingMap map[string]string, prevName, currName string) string {
	if s, ok := spacingMap[prevName+"|||"+currName]; ok {
		return s
	}
	if s, ok := spacingMap[currName+"|||"+prevName]; ok {
		return s
	}
	return "\n\n"
}
