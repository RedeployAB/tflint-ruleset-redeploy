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

// outputBlock represents an output block with its ordering metadata
type outputBlock struct {
	Name     string
	Range    hcl.Range // Full range of the block
	DefRange hcl.Range // Definition range for error reporting
	Start    int
}

// TerraformOutputOrderRule checks that outputs are alphabetically ordered by name
type TerraformOutputOrderRule struct {
	tflint.DefaultRule
}

// NewTerraformOutputOrderRule creates a new rule instance
func NewTerraformOutputOrderRule() *TerraformOutputOrderRule {
	return &TerraformOutputOrderRule{}
}

// Name returns the rule name
func (*TerraformOutputOrderRule) Name() string {
	return "terraform_output_order"
}

// Enabled returns whether the rule is enabled by default
func (*TerraformOutputOrderRule) Enabled() bool {
	return true
}

// Severity returns the severity of the rule
func (*TerraformOutputOrderRule) Severity() tflint.Severity {
	return tflint.ERROR
}

// Link returns the rule's reference link
func (r *TerraformOutputOrderRule) Link() string {
	return GetRuleDocLink(r.Name())
}

// Check checks that output blocks are ordered alphabetically by name
func (r *TerraformOutputOrderRule) Check(runner tflint.Runner) error {
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
			if err := r.processFile(body, filename, runner); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *TerraformOutputOrderRule) processFile(body *hclsyntax.Body, filename string, runner tflint.Runner) error {
	// Collect all output blocks at the top level
	var outputBlocks []outputBlock

	for _, blk := range body.Blocks {
		// Block types are always lowercase in Terraform
		if blk.Type == TypeOutput && len(blk.Labels) > 0 {
			outName := blk.Labels[0]

			// Calculate full range from start of block to end of body
			fullRange := hcl.Range{
				Filename: filename,
				Start:    blk.DefRange().Start,
				End:      blk.Body.Range().End,
			}

			outputBlocks = append(outputBlocks, outputBlock{
				Name:     outName,
				Range:    fullRange,
				DefRange: blk.DefRange(),
				Start:    blk.DefRange().Start.Byte,
			})
		}
	}

	// If no outputs found here, nothing to check
	if len(outputBlocks) == 0 {
		return nil
	}

	// Sort outputs by their starting position
	slices.SortFunc(outputBlocks, func(a, b outputBlock) int {
		return cmp.Compare(a.Start, b.Start)
	})

	// Check if the order is correct
	if isCorrectOutputOrder(outputBlocks) {
		return nil
	}

	// Emit issue with autofix
	return r.emitIssueWithFix(runner, outputBlocks, filename)
}

// isCorrectOutputOrder checks if the output blocks are in alphabetical order
func isCorrectOutputOrder(outputBlocks []outputBlock) bool {
	for i := 1; i < len(outputBlocks); i++ {
		if outputBlocks[i].Name < outputBlocks[i-1].Name {
			return false
		}
	}
	return true
}

// emitIssueWithFix emits an issue with autofix support
func (r *TerraformOutputOrderRule) emitIssueWithFix(
	runner tflint.Runner,
	outputBlocks []outputBlock,
	filename string,
) error {
	// Find the first output that's out of order for the error message and location
	var outOfOrderOutput string
	var outOfOrderRange hcl.Range
	for i := 1; i < len(outputBlocks); i++ {
		if outputBlocks[i].Name < outputBlocks[i-1].Name {
			outOfOrderOutput = outputBlocks[i].Name
			outOfOrderRange = outputBlocks[i].DefRange
			break
		}
	}

	msg := fmt.Sprintf(
		`Out-of-order output %q. Output blocks must be alphabetically ordered by name.`,
		outOfOrderOutput,
	)

	// Use the out-of-order output's range for the issue location
	return runner.EmitIssueWithFix(r, msg, outOfOrderRange, func(f tflint.Fixer) error {
		return fixOutputOrder(f, outputBlocks, filename)
	})
}

// outputBlockWithContent pairs an output block with its source text.
type outputBlockWithContent struct {
	outputBlock
	Content string
}

// outputSpacingKeySeparator joins two output names into a spacing map key.
const outputSpacingKeySeparator = "|||"

// fixOutputOrder rewrites the range spanning all output blocks so that they
// appear alphabetically ordered, preserving the original spacing between
// outputs that were adjacent before the reorder.
func fixOutputOrder(f tflint.Fixer, outputBlocks []outputBlock, filename string) error {
	// Get the text content of all output blocks
	var blocksWithContent []outputBlockWithContent
	for _, ob := range outputBlocks {
		text := f.TextAt(ob.Range)
		blocksWithContent = append(blocksWithContent, outputBlockWithContent{
			outputBlock: ob,
			Content:     string(text.Bytes),
		})
	}

	// Sort outputs alphabetically by name
	slices.SortFunc(blocksWithContent, func(a, b outputBlockWithContent) int {
		return strings.Compare(a.Name, b.Name)
	})

	// Build the fixed content preserving original spacing
	var fixedContent strings.Builder

	spacingMap := outputSpacingMap(f, outputBlocks, filename)

	for i, ob := range blocksWithContent {
		if i > 0 {
			fixedContent.WriteString(lookupOutputSpacing(spacingMap, blocksWithContent[i-1].Name, ob.Name))
		}
		fixedContent.WriteString(ob.Content)
	}

	// Replace the entire range from first to last output
	fullRange := hcl.Range{
		Filename: outputBlocks[0].Range.Filename,
		Start:    outputBlocks[0].Range.Start,
		End:      outputBlocks[len(outputBlocks)-1].Range.End,
	}

	return f.ReplaceText(fullRange, fixedContent.String())
}

// outputSpacingMap records the original text between each pair of
// consecutive outputs, keyed by "<previous>|||<current>".
func outputSpacingMap(f tflint.Fixer, outputBlocks []outputBlock, filename string) map[string]string {
	spacingMap := make(map[string]string)
	for i := 1; i < len(outputBlocks); i++ {
		betweenRange := hcl.Range{
			Filename: filename,
			Start:    outputBlocks[i-1].Range.End,
			End:      outputBlocks[i].Range.Start,
		}
		betweenText := f.TextAt(betweenRange)
		key := outputBlocks[i-1].Name + outputSpacingKeySeparator + outputBlocks[i].Name
		spacingMap[key] = string(betweenText.Bytes)
	}
	return spacingMap
}

// lookupOutputSpacing returns the original spacing between two outputs.
// Both orderings are checked since they might have been reordered; it
// defaults to a double newline if they weren't originally adjacent.
func lookupOutputSpacing(spacingMap map[string]string, prevName, currName string) string {
	if s, ok := spacingMap[prevName+outputSpacingKeySeparator+currName]; ok {
		return s
	}
	if s, ok := spacingMap[currName+outputSpacingKeySeparator+prevName]; ok {
		return s
	}
	return "\n\n"
}
