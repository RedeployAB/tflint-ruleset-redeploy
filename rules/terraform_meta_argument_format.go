package rules

import (
	"fmt"
	"strings"

	"github.com/RedeployAB/tflint-ruleset-redeploy/internal"
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
	return ""
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

// detectMetaArgName checks whether the given trimmed line starts with
// any recognized meta argument. Returns "" if not found.
func (r *TerraformMetaArgumentFormatRule) detectMetaArgName(line string) string {
	if strings.HasPrefix(line, "count ") || strings.HasPrefix(line, "count=") {
		return ArgCount
	}
	if strings.HasPrefix(line, "for_each ") || strings.HasPrefix(line, "for_each=") {
		return ArgForEach
	}
	if strings.HasPrefix(line, "provider ") || strings.HasPrefix(line, "provider=") {
		return ArgProvider
	}
	if strings.HasPrefix(line, "lifecycle ") {
		return ArgLifecycle
	}
	if strings.HasPrefix(line, "depends_on ") || strings.HasPrefix(line, "depends_on=") {
		return ArgDependsOn
	}
	return ""
}

// gatherMetaArgumentIndices scans lines within the block range to locate
// top/bottom meta-argument line indices.
func (r *TerraformMetaArgumentFormatRule) gatherMetaArgumentIndices(
	lines []string,
	startLine, endLine int,
) (countForEachIdx, providerIdx, lifecycleIdx, dependsOnIdx int) {
	countForEachIdx, providerIdx, lifecycleIdx, dependsOnIdx = -1, -1, -1, -1

	for lineNum := startLine; lineNum <= endLine && lineNum < len(lines); lineNum++ {
		trimmed := strings.TrimSpace(lines[lineNum])
		if trimmed == "" {
			continue
		}
		argName := r.detectMetaArgName(trimmed)
		if argName == "" {
			continue
		}

		switch argName {
		case ArgCount, ArgForEach:
			if countForEachIdx < 0 {
				countForEachIdx = lineNum
			}
		case ArgProvider:
			if providerIdx < 0 {
				providerIdx = lineNum
			}
		case ArgLifecycle:
			if lifecycleIdx < 0 {
				lifecycleIdx = lineNum
			}
		case ArgDependsOn:
			if dependsOnIdx < 0 {
				dependsOnIdx = lineNum
			}
		}
	}
	return countForEachIdx, providerIdx, lifecycleIdx, dependsOnIdx
}

// Helper for checking blank line after top meta-arguments
func (r *TerraformMetaArgumentFormatRule) checkBlankLineAfterTopMetaArgs(
	lines []string,
	topIdx, endLine int,
	srcRange hcl.Range,
	runner tflint.Runner,
) error {
	nextLineIdx := topIdx + 1
	for nextLineIdx <= endLine {
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
	argIdx, startLine int,
	srcRange hcl.Range,
	runner tflint.Runner,
) error {
	prevLineIdx := argIdx - 1
	for prevLineIdx > startLine {
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
				Start:    hcl.Pos{Line: argIdx + 1, Column: 1},
				End:      hcl.Pos{Line: argIdx + 1, Column: 1},
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

// checkBlock checks blank lines after top and before bottom meta-arguments
func (r *TerraformMetaArgumentFormatRule) checkBlock(
	block *hclsyntax.Block,
	runner tflint.Runner,
) error {
	srcRange := block.Body.Range()

	files, err := runner.GetFiles()
	if err != nil {
		return err
	}
	hclFile, ok := files[srcRange.Filename]
	if !ok || hclFile.Bytes == nil {
		return nil
	}

	lines := strings.Split(string(hclFile.Bytes), "\n")

	// Parse lines from block start to end to locate meta-argument lines
	startLine := srcRange.Start.Line - 1
	endLine := srcRange.End.Line - 1
	if endLine >= len(lines) {
		endLine = len(lines) - 1
	}

	countForEachIdx, providerIdx, lifecycleIdx, dependsOnIdx :=
		r.gatherMetaArgumentIndices(lines, startLine, endLine)

	// Blank line after top meta-arguments
	topIdx := internal.Max(countForEachIdx, providerIdx)
	if topIdx >= 0 {
		if err := r.checkBlankLineAfterTopMetaArgs(lines, topIdx, endLine, srcRange, runner); err != nil {
			return err
		}
	}

	// Blank line before bottom meta-arguments
	for argName, argIdx := range map[string]int{"lifecycle": lifecycleIdx, "depends_on": dependsOnIdx} {
		if argIdx >= 0 {
			if err := r.checkBlankLineBeforeBottomMetaArgs(
				lines, argName, argIdx, startLine, srcRange, runner,
			); err != nil {
				return err
			}
		}
	}

	// Done
	return nil
}
