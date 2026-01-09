package rules

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// TerraformMetaArgumentFormatRule checks meta-argument formatting in resource and module blocks.
type TerraformMetaArgumentFormatRule struct {
	tflint.DefaultRule
}

func NewTerraformMetaArgumentFormatRule() *TerraformMetaArgumentFormatRule {
	return &TerraformMetaArgumentFormatRule{}
}

func (r *TerraformMetaArgumentFormatRule) Name() string {
	return "terraform_meta_argument_format"
}

func (r *TerraformMetaArgumentFormatRule) Enabled() bool {
	return true
}

func (r *TerraformMetaArgumentFormatRule) Severity() tflint.Severity {
	return tflint.ERROR
}

func (r *TerraformMetaArgumentFormatRule) Link() string {
	return GetRuleDocLink(r.Name())
}

func (r *TerraformMetaArgumentFormatRule) Check(runner tflint.Runner) error {
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

// Helper for checking blank line after top meta-arguments
func (r *TerraformMetaArgumentFormatRule) checkBlankLineAfterTopMetaArgs(
	lines []string,
	topEndLine, endLine int,
	srcRange hcl.Range,
	runner tflint.Runner,
) error {
	nextLineIdx := topEndLine // Lines are 1-based; indices are 0-based
	for nextLineIdx < endLine {
		nextLine := strings.TrimSpace(lines[nextLineIdx])
		switch {
		case nextLine == "":
			// Found blank line => good
			return nil
		case strings.HasPrefix(nextLine, "//"), strings.HasPrefix(nextLine, "#"):
			nextLineIdx++
			continue
		case nextLine == "}":
			// Next is closing brace => no blank line needed
			return nil
		default:
			// No blank line => error
			rng := hcl.Range{
				Filename: srcRange.Filename,
				Start:    hcl.Pos{Line: nextLineIdx + 1, Column: 1},
				End:      hcl.Pos{Line: nextLineIdx + 1, Column: 1},
			}
			return runner.EmitIssue(
				r,
				"Expected a blank line after meta-arguments (count/for_each/provider)",
				rng,
			)
		}
	}
	return nil
}

// Helper for checking blank line before bottom meta-arguments
func (r *TerraformMetaArgumentFormatRule) checkBlankLineBeforeBottomMetaArgs(
	lines []string,
	argName string,
	argStartLine, startLine int,
	srcRange hcl.Range,
	runner tflint.Runner,
) error {
	prevLineIdx := argStartLine - 2 // Move to the line before the argument
	for prevLineIdx >= startLine {
		prevLine := strings.TrimSpace(lines[prevLineIdx])
		switch {
		case prevLine == "":
			// Found blank line => good
			return nil
		case strings.HasPrefix(prevLine, "//"), strings.HasPrefix(prevLine, "#"):
			prevLineIdx--
			continue
		default:
			// Missing blank line
			rng := hcl.Range{
				Filename: srcRange.Filename,
				Start:    hcl.Pos{Line: argStartLine, Column: 1},
				End:      hcl.Pos{Line: argStartLine, Column: 1},
			}
			msg := fmt.Sprintf("Expected a blank line before meta-argument '%s'", argName)
			return runner.EmitIssue(r, msg, rng)
		}
	}
	return nil
}

func (r *TerraformMetaArgumentFormatRule) processBody(body *hclsyntax.Body, runner tflint.Runner) error {
	for _, blk := range body.Blocks {
		if blk.Type == TypeResource || blk.Type == TypeModule {
			if err := r.checkBlock(blk, runner); err != nil {
				return err
			}
		} else {
			// Recursively process nested blocks
			if err := r.processBody(blk.Body, runner); err != nil {
				return err
			}
		}
	}
	return nil
}

// gatherMetaArgEndLines uses the block's attributes/child blocks to compute where
// each meta argument ends (last line). Called in checkBlock() just before we do
// the blank-line checks.
func (r *TerraformMetaArgumentFormatRule) gatherMetaArgEndLines(
	block *hclsyntax.Block,
) (countForEachEndLine, providerEndLine, lifecycleStartLine, dependsOnStartLine int) {
	countForEachEndLine, providerEndLine, lifecycleStartLine, dependsOnStartLine = -1, -1, -1, -1

	// Check each attribute (attribute names are always lowercase in Terraform)
	for _, attr := range block.Body.Attributes {
		switch attr.Name {
		case ArgCount, ArgForEach:
			if attr.Range().End.Line > countForEachEndLine {
				countForEachEndLine = attr.Range().End.Line
			}
		case ArgProvider:
			if attr.Range().End.Line > providerEndLine {
				providerEndLine = attr.Range().End.Line
			}
		case ArgDependsOn:
			if dependsOnStartLine == -1 || attr.Range().Start.Line < dependsOnStartLine {
				dependsOnStartLine = attr.Range().Start.Line
			}
		}
	}

	// Check each child block (e.g. lifecycle) - block types are always lowercase
	for _, child := range block.Body.Blocks {
		if child.Type == ArgLifecycle {
			if lifecycleStartLine == -1 || child.DefRange().Start.Line < lifecycleStartLine {
				lifecycleStartLine = child.DefRange().Start.Line
			}
		}
	}

	return
}

// checkBlock checks blank lines after top and before bottom meta-arguments
func (r *TerraformMetaArgumentFormatRule) checkBlock(
	block *hclsyntax.Block,
	runner tflint.Runner,
) error {
	srcRange := block.Body.Range()

	// Use GetFile for single file lookup instead of GetFiles()
	hclFile, err := runner.GetFile(srcRange.Filename)
	if err != nil {
		return err
	}
	if hclFile == nil || hclFile.Bytes == nil {
		return nil
	}

	lines := strings.Split(string(hclFile.Bytes), "\n")

	startLine := srcRange.Start.Line - 1
	endLine := srcRange.End.Line - 1
	if endLine >= len(lines) {
		endLine = len(lines) - 1
	}

	countForEachEndLine, providerEndLine, lifecycleStartLine, dependsOnStartLine :=
		r.gatherMetaArgEndLines(block)

	// Blank line after top meta-arguments
	topEndLine := Max(countForEachEndLine, providerEndLine)
	if topEndLine >= 0 {
		if err := r.checkBlankLineAfterTopMetaArgs(lines, topEndLine, endLine, srcRange, runner); err != nil {
			return err
		}
	}

	// Blank line before bottom meta-arguments
	for argName, argStartLine := range map[string]int{
		"lifecycle":  lifecycleStartLine,
		"depends_on": dependsOnStartLine,
	} {
		if argStartLine >= 0 {
			if err := r.checkBlankLineBeforeBottomMetaArgs(
				lines, argName, argStartLine, startLine, srcRange, runner,
			); err != nil {
				return err
			}
		}
	}
	// Done
	return nil
}
