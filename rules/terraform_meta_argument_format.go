package rules

import (
	"fmt"
	"strings"

	"github.com/RedeployAB/tflint-ruleset-redeploy/internal"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// TerraformMetaArgumentFormatRule checks the formatting of meta-arguments in resource and module blocks.
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

func (r *TerraformMetaArgumentFormatRule) processBody(body *hclsyntax.Body, runner tflint.Runner) error {
	for _, blk := range body.Blocks {
		if blk.Type == TypeResource || blk.Type == TypeModule {
			if err := r.checkBlock(blk, runner); err != nil {
				return err
			}
		} else {
			// Recursively process other blocks
			if err := r.processBody(blk.Body, runner); err != nil {
				return err
			}
		}
	}

	return nil
}

// Reverted to older blank-line logic for top/bottom meta-arguments.
// The tests expect "Expected a blank line after meta-arguments" and
// "Expected a blank line before meta-argument 'depends_on/lifecycle'".

func (r *TerraformMetaArgumentFormatRule) checkBlock(block *hclsyntax.Block, runner tflint.Runner) error {
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

	// Track discovered line indices for each meta-argument
	var countForEachIdx, providerIdx, lifecycleIdx, dependsOnIdx int
	countForEachIdx, providerIdx, lifecycleIdx, dependsOnIdx = -1, -1, -1, -1

	// We'll parse lines from block start to end to locate meta-argument lines
	startLine := srcRange.Start.Line - 1
	endLine := srcRange.End.Line - 1
	if endLine >= len(lines) {
		endLine = len(lines) - 1
	}

	for l := startLine; l <= endLine && l < len(lines); l++ {
		text := strings.TrimSpace(lines[l])
		if text == "" {
			continue
		}
		switch {
		case strings.HasPrefix(text, "count ") || strings.HasPrefix(text, "count="):
			if countForEachIdx < 0 {
				countForEachIdx = l
			}
		case strings.HasPrefix(text, "for_each ") || strings.HasPrefix(text, "for_each="):
			if countForEachIdx < 0 {
				countForEachIdx = l
			}
		case strings.HasPrefix(text, "provider ") || strings.HasPrefix(text, "provider="):
			if providerIdx < 0 {
				providerIdx = l
			}
		case strings.HasPrefix(text, "lifecycle "):
			if lifecycleIdx < 0 {
				lifecycleIdx = l
			}
		case strings.HasPrefix(text, "depends_on ") || strings.HasPrefix(text, "depends_on="):
			if dependsOnIdx < 0 {
				dependsOnIdx = l
			}
		}
	}

	// If we have a top meta-argument (count/for_each/provider), check for blank line after it
	topIdx := internal.Max(countForEachIdx, providerIdx)
	if topIdx >= 0 {
		nextLineIdx := topIdx + 1
		for nextLineIdx <= endLine {
			nextLine := strings.TrimSpace(lines[nextLineIdx])
			if nextLine == "" {
				// Found blank line => good
				break
			} else if strings.HasPrefix(nextLine, "//") || strings.HasPrefix(nextLine, "#") {
				// skip comment lines
				nextLineIdx++
				continue
			} else if nextLine == "}" {
				// next is closing brace => no blank line needed
				break
			} else {
				// no blank line => error
				rng := hcl.Range{Filename: srcRange.Filename, Start: hcl.Pos{Line: nextLineIdx + 1, Column: 1}, End: hcl.Pos{Line: nextLineIdx + 1, Column: 1}}
				return runner.EmitIssue(
					r,
					"Expected a blank line after meta-arguments (count/for_each/provider)",
					rng,
				)
			}
			break
		}
	}

	// If we have bottom meta-arguments (depends_on, lifecycle), check for blank line BEFORE them
	for argName, argIdx := range map[string]int{"lifecycle": lifecycleIdx, "depends_on": dependsOnIdx} {
		if argIdx >= 0 {
			prevLineIdx := argIdx - 1
			for prevLineIdx > startLine {
				prevLine := strings.TrimSpace(lines[prevLineIdx])
				if prevLine == "" {
					// good => found blank line
					break
				} else if strings.HasPrefix(prevLine, "//") || strings.HasPrefix(prevLine, "#") {
					prevLineIdx--
					continue
				} else {
					// missing blank line
					rng := hcl.Range{Filename: srcRange.Filename, Start: hcl.Pos{Line: argIdx + 1, Column: 1}, End: hcl.Pos{Line: argIdx + 1, Column: 1}}
					msg := fmt.Sprintf("Expected a blank line before meta-argument '%s'", argName)
					return runner.EmitIssue(r, msg, rng)
				}
				break
			}
		}
	}

	return nil
}
