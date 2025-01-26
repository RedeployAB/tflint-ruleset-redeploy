package rules

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// TerraformBlockOrderRule ensures top-level blocks (if used) appear in the order:
// terraform, provider, data, resource
type TerraformBlockOrderRule struct {
	tflint.DefaultRule
}

// NewTerraformBlockOrderRule creates a new rule instance.
func NewTerraformBlockOrderRule() *TerraformBlockOrderRule {
	return &TerraformBlockOrderRule{}
}

// Name returns the rule name.
func (r *TerraformBlockOrderRule) Name() string {
	return "terraform_block_order"
}

// Enabled returns whether the rule is enabled by default.
func (r *TerraformBlockOrderRule) Enabled() bool {
	return true
}

// Severity returns the severity of the rule.
func (r *TerraformBlockOrderRule) Severity() tflint.Severity {
	return tflint.ERROR
}

// Link returns the rule's reference link.
func (r *TerraformBlockOrderRule) Link() string {
	return ""
}

// Check checks the order of top-level blocks.
func (r *TerraformBlockOrderRule) Check(runner tflint.Runner) error {
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
			if err := r.checkTopLevelBlocks(body, filename, runner); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *TerraformBlockOrderRule) checkTopLevelBlocks(
	body *hclsyntax.Body,
	filename string,
	runner tflint.Runner,
) error {
	// The required order:
	orderMap := map[string]int{
		TypeTerraform: 0,
		TypeProvider:  1,
		TypeData:      2,
		TypeResource:  3,
	}

	// Collect recognized blocks in lexical order
	type blockItem struct {
		Type  string
		Index int
		Range hcl.Range
		Start int
	}

	var blocks []blockItem
	for _, blk := range body.Blocks {
		// Only track the known block types
		lcType := strings.ToLower(blk.Type)
		if idx, ok := orderMap[lcType]; ok {
			blocks = append(blocks, blockItem{
				Type:  lcType,
				Index: idx,
				Range: blk.DefRange(),
				Start: blk.DefRange().Start.Byte,
			})
		}
	}

	// Sort by their appearance in the file
	sort.Slice(blocks, func(i, j int) bool {
		return blocks[i].Start < blocks[j].Start
	})

	// Ensure that the order by `Index` is non-decreasing
	lastIndex := -1
	for _, b := range blocks {
		if b.Index < lastIndex {
			msg := fmt.Sprintf("Out-of-order block '%s'. Expected order: terraform -> provider -> data -> resource", b.Type)
			return runner.EmitIssue(r, msg, b.Range)
		}
		lastIndex = b.Index
	}

	// Do NOT recurse into nested blocks
	return nil
}
