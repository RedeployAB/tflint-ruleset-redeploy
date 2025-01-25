package rules

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// TerraformOutputArgumentOrderRule checks that output blocks follow the order:
// description, value, ephemeral, sensitive, depends_on
type TerraformOutputArgumentOrderRule struct {
	tflint.DefaultRule
}

func NewTerraformOutputArgumentOrderRule() *TerraformOutputArgumentOrderRule {
	return &TerraformOutputArgumentOrderRule{}
}

func (r *TerraformOutputArgumentOrderRule) Name() string {
	return "terraform_output_argument_order"
}

func (r *TerraformOutputArgumentOrderRule) Enabled() bool {
	return true
}

func (r *TerraformOutputArgumentOrderRule) Severity() tflint.Severity {
	return tflint.ERROR
}

func (r *TerraformOutputArgumentOrderRule) Link() string {
	return ""
}

func (r *TerraformOutputArgumentOrderRule) Check(runner tflint.Runner) error {
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

func (r *TerraformOutputArgumentOrderRule) processBody(body *hclsyntax.Body, runner tflint.Runner) error {
	for _, block := range body.Blocks {
		if strings.EqualFold(block.Type, "output") {
			if err := r.checkOutputBlock(block, runner); err != nil {
				return err
			}
		}
		// recurse for nested blocks
		if err := r.processBody(block.Body, runner); err != nil {
			return err
		}
	}
	return nil
}

// checkOutputBlock enforces the order description -> value -> ephemeral -> sensitive -> depends_on
func (r *TerraformOutputArgumentOrderRule) checkOutputBlock(
	block *hclsyntax.Block,
	runner tflint.Runner,
) error {
	orderMap := map[string]int{
		"description": 0,
		"value":       1,
		"ephemeral":   2,
		"sensitive":   3,
		"depends_on":  4,
	}

	type item struct {
		Name  string
		Index int
		Range hcl.Range
		Start int
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
			})
		}
	}

	// If nothing recognized, do nothing
	if len(items) == 0 {
		return nil
	}

	// Sort by their position in the file
	sort.Slice(items, func(i, j int) bool {
		return items[i].Start < items[j].Start
	})

	lastIndex := -1
	for _, it := range items {
		if it.Index < lastIndex {
			msg := "Out-of-order argument '" + it.Name + "'. Expected sequence: description, value, ephemeral, sensitive, depends_on"
			return runner.EmitIssue(r, msg, it.Range)
		}
		lastIndex = it.Index
	}
	return nil
}
