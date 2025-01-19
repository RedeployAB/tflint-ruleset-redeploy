package rules

import (
	"fmt"
	"sort"

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
			Type:     TypeAttr,
			Attr:     attr,
			SrcRange: attr.Range(),
		})
	}
	for _, blk := range body.Blocks {
		items = append(items, contentItem{
			Name:     blk.Type,
			Type:     TypeBlock,
			Block:    blk,
			SrcRange: blk.DefRange(),
		})
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].SrcRange.Start.Byte < items[j].SrcRange.Start.Byte
	})

	for _, it := range items {
		if it.Type == TypeBlock {
			if it.Block.Type == TypeResource || it.Block.Type == TypeModule {
				if err := r.checkBlock(it.Block, runner); err != nil {
					return err
				}
			} else {
				// Recursively process other blocks
				if err := r.processBody(it.Block.Body, runner); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (r *TerraformMetaArgumentFormatRule) checkBlock(block *hclsyntax.Block, runner tflint.Runner) error {
	// Ensure no blank lines between meta arguments
	metaArgs := []string{ArgCount, ArgForEach, ArgProvider, ArgLifecycle, ArgDependsOn}

	type contentItem struct {
		Name     string
		Type     string
		Line     int
		SrcRange hcl.Range
	}

	var items []contentItem
	for _, attr := range block.Body.Attributes {
		items = append(items, contentItem{
			Name:     attr.Name,
			Type:     TypeAttr,
			Line:     attr.Range().Start.Line,
			SrcRange: attr.Range(),
		})
	}
	for _, blk := range block.Body.Blocks {
		items = append(items, contentItem{
			Name:     blk.Type,
			Type:     TypeBlock,
			Line:     blk.DefRange().Start.Line,
			SrcRange: blk.DefRange(),
		})
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Line < items[j].Line
	})

	var metaArgIndices []int
	for i, item := range items {
		if contains(metaArgs, item.Name) {
			metaArgIndices = append(metaArgIndices, i)
		}
	}

	for i := 0; i < len(metaArgIndices)-1; i++ {
		currentItem := items[metaArgIndices[i]]
		nextItem := items[metaArgIndices[i+1]]

		blankLines := nextItem.Line - currentItem.SrcRange.End.Line - 1
		if blankLines != 0 {
			return runner.EmitIssue(
				r,
				fmt.Sprintf("Meta arguments '%s' and '%s' should have no blank lines between them", currentItem.Name, nextItem.Name),
				nextItem.SrcRange,
			)
		}
	}

	return nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
