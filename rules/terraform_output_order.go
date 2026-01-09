package rules

import (
	"fmt"
	"sort"
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
func (r *TerraformOutputOrderRule) Name() string {
	return "terraform_output_order"
}

// Enabled returns whether the rule is enabled by default
func (r *TerraformOutputOrderRule) Enabled() bool {
	return true
}

// Severity returns the severity of the rule
func (r *TerraformOutputOrderRule) Severity() tflint.Severity {
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
	sort.Slice(outputBlocks, func(i, j int) bool {
		return outputBlocks[i].Start < outputBlocks[j].Start
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
		// Get the text content of all output blocks
		type outputBlockWithContent struct {
			outputBlock
			Content string
		}

		var blocksWithContent []outputBlockWithContent
		for _, ob := range outputBlocks {
			text := f.TextAt(ob.Range)
			blocksWithContent = append(blocksWithContent, outputBlockWithContent{
				outputBlock: ob,
				Content:     string(text.Bytes),
			})
		}

		// Sort outputs alphabetically by name
		sort.Slice(blocksWithContent, func(i, j int) bool {
			return blocksWithContent[i].Name < blocksWithContent[j].Name
		})

		// Build the fixed content preserving original spacing
		var fixedContent strings.Builder

		// Create a map to track original spacing between consecutive outputs
		spacingMap := make(map[string]string)
		for i := 1; i < len(outputBlocks); i++ {
			betweenRange := hcl.Range{
				Filename: filename,
				Start:    outputBlocks[i-1].Range.End,
				End:      outputBlocks[i].Range.Start,
			}
			betweenText := f.TextAt(betweenRange)
			key := outputBlocks[i-1].Name + "|||" + outputBlocks[i].Name
			spacingMap[key] = string(betweenText.Bytes)
		}

		for i, ob := range blocksWithContent {
			if i > 0 {
				// Try to find original spacing between these two outputs
				prevName := blocksWithContent[i-1].Name
				currName := ob.Name

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
			fixedContent.WriteString(ob.Content)
		}

		// Replace the entire range from first to last output
		fullRange := hcl.Range{
			Filename: outputBlocks[0].Range.Filename,
			Start:    outputBlocks[0].Range.Start,
			End:      outputBlocks[len(outputBlocks)-1].Range.End,
		}

		return f.ReplaceText(fullRange, fixedContent.String())
	})
}
