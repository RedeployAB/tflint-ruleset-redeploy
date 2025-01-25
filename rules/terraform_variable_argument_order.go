package rules

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// TerraformVariableArgumentOrderRule enforces argument order for variable blocks:
//   description, type, default, sensitive, nullable, validation
// Any of these may be omitted, but if present, must follow that sequence.
// For validation blocks, multiple occurrences are allowed, but all must appear after the others.
type TerraformVariableArgumentOrderRule struct {
	tflint.DefaultRule
}

func NewTerraformVariableArgumentOrderRule() *TerraformVariableArgumentOrderRule {
	return &TerraformVariableArgumentOrderRule{}
}

func (r *TerraformVariableArgumentOrderRule) Name() string {
	return "terraform_variable_argument_order"
}

func (r *TerraformVariableArgumentOrderRule) Enabled() bool {
	return true
}

func (r *TerraformVariableArgumentOrderRule) Severity() tflint.Severity {
	return tflint.ERROR
}

func (r *TerraformVariableArgumentOrderRule) Link() string {
	return ""
}

func (r *TerraformVariableArgumentOrderRule) Check(runner tflint.Runner) error {
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

func (r *TerraformVariableArgumentOrderRule) processBody(body *hclsyntax.Body, runner tflint.Runner) error {
	for _, block := range body.Blocks {
		// Only examine variable blocks
		if strings.ToLower(block.Type) == "variable" {
			if err := r.checkVariableBlock(block, runner); err != nil {
				return err
			}
		}
		// Recurse into nested blocks
		if err := r.processBody(block.Body, runner); err != nil {
			return err
		}
	}
	return nil
}

func (r *TerraformVariableArgumentOrderRule) checkVariableBlock(
	block *hclsyntax.Block,
	runner tflint.Runner,
) error {
	// Recognized order for attributes/blocks:
	// description(0), type(1), default(2), sensitive(3), nullable(4), validation(5)
	// Any of these may be omitted, but if present, must follow that sequence.
	// For validation blocks, multiple occurrences are allowed, but all must appear after the others.

	// Define the expected order
	orderMap := map[string]int{
		"description": 0,
		"type":        1,
		"default":     2,
		"sensitive":   3,
		"nullable":    4,
		"validation":  5,
	}

	type item struct {
		Name  string
		Index int
		Range hcl.Range
		Start int
		IsBlk bool
	}

	var items []item

	// Gather recognized attributes
	for _, attr := range block.Body.Attributes {
		lcName := strings.ToLower(attr.Name)
		idx, found := orderMap[lcName]
		if found {
			items = append(items, item{
				Name:  lcName,
				Index: idx,
				Range: attr.Range(),
				Start: attr.Range().Start.Byte,
				IsBlk: false,
			})
		}
	}

	// Gather recognized blocks: "validation"
	for _, childBlock := range block.Body.Blocks {
		lcType := strings.ToLower(childBlock.Type)
		if lcType == "validation" {
			items = append(items, item{
				Name:  lcType,
				Index: orderMap[lcType], // 5
				Range: childBlock.DefRange(),
				Start: childBlock.DefRange().Start.Byte,
				IsBlk: true,
			})
		}
	}

	// If no recognized items => no check needed
	if len(items) == 0 {
		return nil
	}

	// Sort items by lexical file order
	sort.Slice(items, func(i, j int) bool {
		return items[i].Start < items[j].Start
	})

	// Track the highest index encountered so far
	lastIndex := -1

	for _, it := range items {
		if it.Index < lastIndex {
			// Out-of-order argument found
			msg := fmt.Sprintf("Out-of-order argument '%s'. Expected sequence: description, type, default, sensitive, nullable, validation", it.Name)
			return runner.EmitIssue(r, msg, it.Range)
		}
		lastIndex = it.Index
	}
	return nil
}
