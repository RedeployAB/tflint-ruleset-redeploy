package rules

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

const (
	blockTypeResource   = "resource"
	blockTypeModule     = "module"
	contentTypeBlock    = "block"
	contentTypeAttr     = "attribute"
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
	type contentItem struct {
		Name     string
		Type     string
		Block    *hclsyntax.Block
		Attr     *hclsyntax.Attribute
		SrcRange hcl.Range
	}

	var items []contentItem
	for _, attr := range body.Attributes {
		items = append(items, contentItem{
			Name:     attr.Name,
			Type:     contentTypeAttr,
			Attr:     attr,
			SrcRange: attr.Range(),
		})
	}
	for _, blk := range body.Blocks {
		items = append(items, contentItem{
			Name:     blk.Type,
			Type:     contentTypeBlock,
			Block:    blk,
			SrcRange: blk.DefRange(),
		})
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].SrcRange.Start.Byte < items[j].SrcRange.Start.Byte
	})

	for _, it := range items {
		if it.Type == contentTypeBlock {
			if it.Block.Type == blockTypeResource || it.Block.Type == blockTypeModule {
				if err := r.checkFormatting(it.Block, runner); err != nil {
					return err
				}
			} else {
				if err := r.processBody(it.Block.Body, runner); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (r *TerraformMetaArgumentFormatRule) checkFormatting(block *hclsyntax.Block, runner tflint.Runner) error {
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

	var countForEachIdx, providerIdx, lifecycleIdx, dependsOnIdx int
	countForEachIdx, providerIdx, lifecycleIdx, dependsOnIdx = -1, -1, -1, -1

	metaArgLines := make(map[string]int)

	startLine := srcRange.Start.Line - 1
	endLine := srcRange.End.Line - 1
	if endLine >= len(lines) {
		endLine = len(lines) - 1
	}

	for l := startLine; l <= endLine && l < len(lines); l++ {
		text := strings.TrimSpace(lines[l])

		switch {
		case strings.HasPrefix(text, "count ") || strings.HasPrefix(text, "count="):
			if _, exists := metaArgLines["count"]; !exists {
				metaArgLines["count"] = l
			}
		case strings.HasPrefix(text, "for_each ") || strings.HasPrefix(text, "for_each="):
			if _, exists := metaArgLines["for_each"]; !exists {
				metaArgLines["for_each"] = l
			}
		case strings.HasPrefix(text, "provider ") || strings.HasPrefix(text, "provider="):
			if _, exists := metaArgLines["provider"]; !exists {
				metaArgLines["provider"] = l
			}
		case strings.HasPrefix(text, "lifecycle "):
			if _, exists := metaArgLines["lifecycle"]; !exists {
				metaArgLines["lifecycle"] = l
			}
		case strings.HasPrefix(text, "depends_on ") || strings.HasPrefix(text, "depends_on="):
			if _, exists := metaArgLines["depends_on"]; !exists {
				metaArgLines["depends_on"] = l
			}
		}
	}

	if idx, ok := metaArgLines["count"]; ok {
		countForEachIdx = idx
	} else if idx, ok := metaArgLines["for_each"]; ok {
		countForEachIdx = idx
	}
	if idx, ok := metaArgLines["provider"]; ok {
		providerIdx = idx
	}

	if idx, ok := metaArgLines["lifecycle"]; ok {
		lifecycleIdx = idx
	}
	if idx, ok := metaArgLines["depends_on"]; ok {
		dependsOnIdx = idx
	}

	topIdx := max(countForEachIdx, providerIdx)
	if topIdx >= 0 {
		nextLineIdx := topIdx + 1
		for nextLineIdx <= endLine {
			nextLine := strings.TrimSpace(lines[nextLineIdx])
			switch {
			case nextLine == "":
				break
			case strings.HasPrefix(nextLine, "//"), strings.HasPrefix(nextLine, "#"):
				nextLineIdx++
			case nextLine == "}":
				break
			default:
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

	for argName, argIdx := range map[string]int{"lifecycle": lifecycleIdx, "depends_on": dependsOnIdx} {
		if argIdx >= 0 {
			prevLineIdx := argIdx - 1
			for prevLineIdx > startLine {
				prevLine := strings.TrimSpace(lines[prevLineIdx])
				switch {
				case prevLine == "":
					break
				case strings.HasPrefix(prevLine, "//"), strings.HasPrefix(prevLine, "#"):
					prevLineIdx--
					continue
				default:
					rng := hcl.Range{Filename: srcRange.Filename, Start: hcl.Pos{Line: argIdx + 1, Column: 1}, End: hcl.Pos{Line: argIdx + 1, Column: 1}}
					errMsg := fmt.Sprintf("Expected a blank line before meta-argument '%s'", argName)
					return runner.EmitIssue(r, errMsg, rng)
				}
				break
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
