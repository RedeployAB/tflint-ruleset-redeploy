package rules

import (
	"fmt"
	"sort"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

const (
	typeAttr  = "attr"
	typeBlock = "block"
)

// TerraformTagsArgumentRule enforces that if a resource has a "tags" attribute,
// the "tags" must appear *after* all normal resource arguments, but *before*
// depends_on and lifecycle blocks. The user must also ensure that each
// boundary (tags -> depends_on, or tags -> lifecycle) is separated by exactly
// one blank line.
//
// Examples:
//
//	# OK usage
//	resource "aws_nat_gateway" "this" {
//	  count = 2
//
//	  allocation_id = "..."
//	  subnet_id     = "..."
//
//	  tags = {
//	    Name = "..."
//	  }
//
//	  depends_on = [aws_internet_gateway.this]
//
//	  lifecycle {
//	    create_before_destroy = true
//	  }
//	}
//
//	# Not OK usage: tags not last real argument
//	resource "aws_nat_gateway" "this" {
//	  count = 2
//
//	  tags = "..."
//
//	  depends_on = [aws_internet_gateway.this]
//
//	  lifecycle {
//	    create_before_destroy = true
//	  }
//
//	  allocation_id = "..."
//	  subnet_id     = "..."
//	}
type TerraformTagsArgumentRule struct {
	tflint.DefaultRule
}

func NewTerraformTagsArgumentRule() *TerraformTagsArgumentRule {
	return &TerraformTagsArgumentRule{}
}

func (r *TerraformTagsArgumentRule) Name() string {
	return "terraform_tags_argument"
}

func (r *TerraformTagsArgumentRule) Enabled() bool {
	return true
}

func (r *TerraformTagsArgumentRule) Severity() tflint.Severity {
	return tflint.ERROR
}

func (r *TerraformTagsArgumentRule) Link() string {
	return GetRuleDocLink(r.Name())
}

func (r *TerraformTagsArgumentRule) Check(runner tflint.Runner) error {
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
			// skip if parse errors
			continue
		}

		if body, ok := syntaxFile.Body.(*hclsyntax.Body); ok {
			if err := r.processBody(body, filename, runner); err != nil {
				return err
			}
		}
	}

	return nil
}

// processBody recursively processes blocks and checks resource blocks
func (r *TerraformTagsArgumentRule) processBody(body *hclsyntax.Body, filename string, runner tflint.Runner) error {
	for _, block := range body.Blocks {
		if block.Type == "resource" {
			if err := r.checkResourceBlock(block, runner); err != nil {
				return err
			}
		}
		// Recurse
		if err := r.processBody(block.Body, filename, runner); err != nil {
			return err
		}
	}
	return nil
}

func (r *TerraformTagsArgumentRule) checkResourceBlock(
	block *hclsyntax.Block,
	runner tflint.Runner,
) error {
	items := r.collectResourceItems(block)

	tagsIndex := r.findTagsIndex(items)
	if tagsIndex < 0 {
		return nil
	}

	// Step 1: ensure no normal attribute or block after tags
	if err := r.checkItemsAfterTags(items, tagsIndex, runner); err != nil {
		return err
	}
	// Step 2: ensure exactly one blank line
	return r.checkBlankLineAfterTags(items, tagsIndex, runner)
}

type resourceItem struct {
	Name  string
	Type  string
	Range hcl.Range
}

func (r *TerraformTagsArgumentRule) collectResourceItems(block *hclsyntax.Block) []resourceItem {
	var items []resourceItem

	for _, attr := range block.Body.Attributes {
		items = append(items, resourceItem{
			Name:  attr.Name,
			Type:  typeAttr,
			Range: attr.Range(),
		})
	}
	for _, blk := range block.Body.Blocks {
		items = append(items, resourceItem{
			Name:  blk.Type,
			Type:  typeBlock,
			Range: blk.DefRange(),
		})
	}

	// Sort items by position
	sort.Slice(items, func(i, j int) bool {
		return items[i].Range.Start.Byte < items[j].Range.Start.Byte
	})
	return items
}

func (r *TerraformTagsArgumentRule) findTagsIndex(items []resourceItem) int {
	for i, it := range items {
		if it.Type == typeAttr && it.Name == "tags" {
			return i
		}
	}
	return -1
}

func (r *TerraformTagsArgumentRule) checkItemsAfterTags(
	items []resourceItem,
	tagsIndex int,
	runner tflint.Runner,
) error {
	// Step 1: ensure no normal attribute or block after tags, except depends_on or lifecycle
	for i := tagsIndex + 1; i < len(items); i++ {
		switch items[i].Type {
		case typeAttr:
			if items[i].Name != "depends_on" {
				return r.emitIssue(runner, items[i].Range,
					fmt.Sprintf("Argument '%s' must not come after 'tags'", items[i].Name))
			}
		case typeBlock:
			if items[i].Name != "lifecycle" {
				return r.emitIssue(runner, items[i].Range,
					fmt.Sprintf("Block '%s' must not come after 'tags'", items[i].Name))
			}
		}
	}
	return nil
}

func (r *TerraformTagsArgumentRule) checkBlankLineAfterTags(
	items []resourceItem,
	tagsIndex int,
	runner tflint.Runner,
) error {
	// Check single blank line
	tagsLine := items[tagsIndex].Range.End.Line
	var nextItem *resourceItem
	if tagsIndex+1 < len(items) {
		nextItem = &items[tagsIndex+1]
	}

	if nextItem != nil && (nextItem.Name == "depends_on" || nextItem.Name == "lifecycle") {
		linesBetween := nextItem.Range.Start.Line - tagsLine - 1
		if linesBetween != 1 {
			return r.emitIssue(runner, nextItem.Range,
				fmt.Sprintf("Expected exactly one blank line between 'tags' and '%s'", nextItem.Name))
		}
	}
	return nil
}

func (r *TerraformTagsArgumentRule) emitIssue(runner tflint.Runner, rng hcl.Range, msg string) error {
	return runner.EmitIssue(r, msg, rng)
}
