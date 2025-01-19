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
//   # OK usage
//   resource "aws_nat_gateway" "this" {
//     count = 2
//
//     allocation_id = "..."
//     subnet_id     = "..."
//
//     tags = {
//       Name = "..."
//     }
//
//     depends_on = [aws_internet_gateway.this]
//
//     lifecycle {
//       create_before_destroy = true
//     }
//   }
//
//   # Not OK usage: tags not last real argument
//   resource "aws_nat_gateway" "this" {
//     count = 2
//
//     tags = "..."
//
//     depends_on = [aws_internet_gateway.this]
//
//     lifecycle {
//       create_before_destroy = true
//     }
//
//     allocation_id = "..."
//     subnet_id     = "..."
//   }
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
	return ""
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

// checkResourceBlock checks if "tags" is present, and if so, ensures
// it is the last normal argument (i.e., after all other non-meta attributes).
// Then "depends_on" and "lifecycle" (if they exist) come after tags,
// each separated by exactly one blank line from tags.
func (r *TerraformTagsArgumentRule) checkResourceBlock(block *hclsyntax.Block, runner tflint.Runner) error {
	// We'll gather the lexical ordering of all content, then locate "tags," "depends_on," "lifecycle"
	type item struct {
		Name  string // attribute or block type
		Type  string // "attr" or "block"
		Range hcl.Range
	}

	var items []item

	for _, attr := range block.Body.Attributes {
		items = append(items, item{
			Name:  attr.Name,
			Type:  typeAttr,
			Range: attr.Range(),
		})
	}
	for _, blk := range block.Body.Blocks {
		items = append(items, item{
			Name:  blk.Type,
			Type:  typeBlock,
			Range: blk.DefRange(),
		})
	}

	// Sort items by position
	sort.Slice(items, func(i, j int) bool {
		return items[i].Range.Start.Byte < items[j].Range.Start.Byte
	})

	// Find if "tags" is present
	tagsIndex := -1
	for i, it := range items {
		if it.Type == typeAttr && it.Name == "tags" {
			tagsIndex = i
			break
		}
	}
	if tagsIndex < 0 {
		// If no tags -> no checks
		return nil
	}

	// If tags is present, ensure all normal attributes come before it
	// but "depends_on" and "lifecycle" must come after it. We'll also
	// check for the single empty line requirement.

	// We'll read the actual file lines for the single-blank-line checks
	files, err := runner.GetFiles()
	if err != nil {
		return err
	}
	hclFile, ok := files[block.Body.Range().Filename]
	if !ok || hclFile.Bytes == nil {
		return nil
	}

	// Step 1: ensure no normal attribute or block after tags, except depends_on or lifecycle
	for i := tagsIndex + 1; i < len(items); i++ {
		if items[i].Type == typeAttr {
			// If it's depends_on -> OK. If it's "tags" again? That's weird. We'll skip. If it's anything else -> NOT OK
			if items[i].Name != "depends_on" {
				return r.emitIssue(runner, items[i].Range, fmt.Sprintf("Argument '%s' must not come after 'tags'", items[i].Name))
			}
		} else if items[i].Type == typeBlock {
			// If it's "lifecycle" -> OK. Otherwise -> not OK
			if items[i].Name != "lifecycle" {
				return r.emitIssue(runner, items[i].Range, fmt.Sprintf("Block '%s' must not come after 'tags'", items[i].Name))
			}
		}
	}

	// Step 2: ensure exactly one blank line separates tags from depends_on or lifecycle if they exist
	// We'll check the lines in lexical order. The line containing tags is items[tagsIndex].Range.End.Line
	tagsLine := items[tagsIndex].Range.End.Line
	// We'll look for the next item if it's depends_on or lifecycle
	if tagsIndex+1 < len(items) {
		next := items[tagsIndex+1]
		if (next.Name == "depends_on" && next.Type == typeAttr) || (next.Name == "lifecycle" && next.Type == typeBlock) {
			// We want exactly one blank line between line-of-tags and line-of-next
			// line-of-tags is tagsLine, line-of-next is next.Range.Start.Line
			linesBetween := next.Range.Start.Line - tagsLine - 1
			if linesBetween != 1 {
				return r.emitIssue(runner, next.Range, fmt.Sprintf("Expected exactly one blank line between 'tags' and '%s'", next.Name))
			}
		}
	}

	return nil
}

func (r *TerraformTagsArgumentRule) emitIssue(runner tflint.Runner, rng hcl.Range, msg string) error {
	return runner.EmitIssue(r, msg, rng)
}
