package rules

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

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
		// Skip empty or binary files
		if hclFile == nil || hclFile.Bytes == nil {
			continue
		}

		syntaxFile, diags := hclsyntax.ParseConfig(hclFile.Bytes, filename, hcl.InitialPos)
		if diags.HasErrors() {
			// If parse fails, skip this file
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
	type contentItem struct {
		Name     string // Name of the attribute or block
		Type     string // "attribute" or "block"
		Block    *hclsyntax.Block
		Attr     *hclsyntax.Attribute
		SrcRange hcl.Range
	}

	// Collect attributes and blocks
	var items []contentItem
	for _, attr := range body.Attributes {
		items = append(items, contentItem{
			Name:     attr.Name,
			Type:     "attribute",
			Attr:     attr,
			SrcRange: attr.Range(),
		})
	}
	for _, blk := range body.Blocks {
		items = append(items, contentItem{
			Name:     blk.Type,
			Type:     "block",
			Block:    blk,
			SrcRange: blk.DefRange(),
		})
	}

	// Sort items in lexical order
	sort.Slice(items, func(i, j int) bool {
		return items[i].SrcRange.Start.Byte < items[j].SrcRange.Start.Byte
	})

	// Iterate over items
	for _, it := range items {
		if it.Type == "block" {
			if it.Block.Type == "resource" || it.Block.Type == "module" {
				// Check the formatting of meta-arguments
				if err := r.checkFormatting(it.Block, runner); err != nil {
					return err
				}
			} else {
				// Recurse into nested blocks
				if err := r.processBody(it.Block.Body, runner); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (r *TerraformMetaArgumentFormatRule) checkFormatting(block *hclsyntax.Block, runner tflint.Runner) error {
	// Get the block's source range
	srcRange := block.DefRange()

	// Get the file content from runner.GetFiles()
	files, err := runner.GetFiles()
	if err != nil {
		return err
	}
	hclFile, ok := files[srcRange.Filename]
	if !ok || hclFile.Bytes == nil {
		// If the file isn't found or has no content, we can return early or handle as needed.
		return nil
	}

	// Convert file content into lines
	lines := strings.Split(string(hclFile.Bytes), "\n")

	// Initialize indices for meta-arguments
	var countForEachIdx, providerIdx, lifecycleIdx, dependsOnIdx int
	countForEachIdx, providerIdx, lifecycleIdx, dependsOnIdx = -1, -1, -1, -1

	// Map to store line numbers of meta-arguments
	metaArgLines := make(map[string]int)

	// Determine the start and end lines of the block
	startLine := srcRange.Start.Line - 1
	endLine := srcRange.End.Line - 1
	if endLine >= len(lines) {
		endLine = len(lines) - 1
	}

	// Scan each line within the block
	for l := startLine; l <= endLine && l < len(lines); l++ {
		text := strings.TrimSpace(lines[l])

		// Skip empty lines
		if text == "" {
			continue
		}

		// Check for meta-arguments
		if strings.HasPrefix(text, "count ") || strings.HasPrefix(text, "count=") {
			if _, exists := metaArgLines["count"]; !exists {
				metaArgLines["count"] = l
			}
		} else if strings.HasPrefix(text, "for_each ") || strings.HasPrefix(text, "for_each=") {
			if _, exists := metaArgLines["for_each"]; !exists {
				metaArgLines["for_each"] = l
			}
		} else if strings.HasPrefix(text, "provider ") || strings.HasPrefix(text, "provider=") {
			if _, exists := metaArgLines["provider"]; !exists {
				metaArgLines["provider"] = l
			}
		} else if strings.HasPrefix(text, "lifecycle ") {
			if _, exists := metaArgLines["lifecycle"]; !exists {
				metaArgLines["lifecycle"] = l
			}
		} else if strings.HasPrefix(text, "depends_on ") || strings.HasPrefix(text, "depends_on=") {
			if _, exists := metaArgLines["depends_on"]; !exists {
				metaArgLines["depends_on"] = l
			}
		}
	}

	// Get the indices for top meta-arguments
	if idx, ok := metaArgLines["count"]; ok {
		countForEachIdx = idx
	} else if idx, ok := metaArgLines["for_each"]; ok {
		countForEachIdx = idx
	}
	if idx, ok := metaArgLines["provider"]; ok {
		providerIdx = idx
	}

	// Get the indices for bottom meta-arguments
	if idx, ok := metaArgLines["lifecycle"]; ok {
		lifecycleIdx = idx
	}
	if idx, ok := metaArgLines["depends_on"]; ok {
		dependsOnIdx = idx
	}

	// Check for blank line after top meta-arguments
	topIdx := max(countForEachIdx, providerIdx)
	if topIdx >= 0 {
		nextLineIdx := topIdx + 1
		// Skip comments
		for nextLineIdx <= endLine {
			nextLine := strings.TrimSpace(lines[nextLineIdx])
			if nextLine == "" {
				// Blank line found
				break
			} else if strings.HasPrefix(nextLine, "//") || strings.HasPrefix(nextLine, "#") {
				// Skip comment lines
				nextLineIdx++
				continue
			} else {
				rng := hcl.Range{Filename: srcRange.Filename, Start: hcl.Pos{Line: nextLineIdx + 1, Column: 1}, End: hcl.Pos{Line: nextLineIdx + 1, Column: 1}}
				return runner.EmitIssue(
					r,
					"Expected a blank line after meta-arguments (count/for_each/provider)",
					rng,
				)
			}
		}
	}

	// Check for blank line before bottom meta-arguments
	for argName, argIdx := range map[string]int{"lifecycle": lifecycleIdx, "depends_on": dependsOnIdx} {
		if argIdx >= 0 {
			prevLineIdx := argIdx - 1
			// Skip comment lines
			for prevLineIdx > startLine {
				prevLine := strings.TrimSpace(lines[prevLineIdx])
				if prevLine == "" {
					// blank line found
					break
				} else if strings.HasPrefix(prevLine, "//") || strings.HasPrefix(prevLine, "#") {
					prevLineIdx--
					continue
				} else {
					rng := hcl.Range{Filename: srcRange.Filename, Start: hcl.Pos{Line: argIdx + 1, Column: 1}, End: hcl.Pos{Line: argIdx + 1, Column: 1}}
					errMsg := fmt.Sprintf("Expected a blank line before meta-argument '%s'", argName)
					return runner.EmitIssue(r, errMsg, rng)
				}
			}
		}
	}

	return nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
