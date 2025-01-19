package rules

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// TerraformResourceArgumentOrderRule ensures that for resource/data/provider/terraform blocks:
// Non-block attributes come first, followed by block attributes (sub-blocks).
// This rule is recursive, so it checks nested blocks as well.
// It ignores meta-arguments (count, for_each, depends_on, lifecycle, provider).
// Those are covered by other rules.
type TerraformResourceArgumentOrderRule struct {
	tflint.DefaultRule
}

func NewTerraformResourceArgumentOrderRule() *TerraformResourceArgumentOrderRule {
	return &TerraformResourceArgumentOrderRule{}
}

func (r *TerraformResourceArgumentOrderRule) Name() string {
	return "terraform_resource_argument_order"
}

func (r *TerraformResourceArgumentOrderRule) Enabled() bool {
	return true
}

func (r *TerraformResourceArgumentOrderRule) Severity() tflint.Severity {
	return tflint.ERROR
}

func (r *TerraformResourceArgumentOrderRule) Link() string {
	return ""
}

func (r *TerraformResourceArgumentOrderRule) Check(runner tflint.Runner) error {
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
			if err := r.processBody(body, runner); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *TerraformResourceArgumentOrderRule) processBody(body *hclsyntax.Body, runner tflint.Runner) error {
	for _, block := range body.Blocks {
		// We only care about blocks of type resource, data, provider, or terraform
		blockType := strings.ToLower(block.Type)
		if blockType == TypeResource || blockType == TypeData ||
			blockType == TypeProvider || blockType == TypeTerraform {
			if err := r.checkArgumentOrder(block, runner); err != nil {
				return err
			}
		}
		// Recurse
		if err := r.processBody(block.Body, runner); err != nil {
			return err
		}
	}
	return nil
}

// checkArgumentOrder checks that within this block:
//  - all non-block attributes come first
//  - then all sub-blocks
// We skip meta arguments. Then we apply the same logic inside each sub-block, but
// that recursion is done in processBody -> checkArgumentOrder.
func (r *TerraformResourceArgumentOrderRule) checkArgumentOrder(block *hclsyntax.Block, runner tflint.Runner) error {
	// Collect attributes vs sub-blocks
	var attrs []string
	var blocks []hclsyntax.Block

	for _, attr := range block.Body.Attributes {
		// If it's a recognized meta arg, skip
		if isMetaArg(attr.Name) {
			continue
		}
		attrs = append(attrs, attr.Name)
	}
	for _, child := range block.Body.Blocks {
		// If it's recognized meta block, skip
		if isMetaArg(child.Type) {
			continue
		}
		blocks = append(blocks, *child)
	}

	// We want all attrs first, then all blocks in lexical order.
	// Actually we only want to fail if a block appears, and then another attr *after* that block.
	// We'll gather items in lexical order, ignoring meta arguments.

	type item struct {
		Name  string
		IsBlk bool
		Range hcl.Range
		Idx   int
	}

	var items []item
	// Add each attr in lexical order
	for _, a := range block.Body.Attributes {
		if isMetaArg(a.Name) {
			continue
		}
		items = append(items, item{
			Name:  a.Name,
			IsBlk: false,
			Range: a.Range(),
			Idx:   a.Range().Start.Byte,
		})
	}
	for _, b := range block.Body.Blocks {
		if isMetaArg(b.Type) {
			continue
		}
		items = append(items, item{
			Name:  b.Type,
			IsBlk: true,
			Range: b.DefRange(),
			Idx:   b.DefRange().Start.Byte,
		})
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Idx < items[j].Idx
	})

	var seenBlock bool
	for _, it := range items {
		if it.IsBlk {
			seenBlock = true
		} else {
			// If we've already seen a block, having an attribute after it is invalid
			if seenBlock {
				if err := runner.EmitIssue(
					r,
					fmt.Sprintf("Argument '%s' must not come after a nested block", it.Name),
					it.Range,
				); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func isMetaArg(name string) bool {
	switch strings.ToLower(name) {
	case ArgCount, ArgForEach, ArgProvider, ArgDependsOn, ArgLifecycle:
		return true
	}
	return false
}
