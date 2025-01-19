package rules

import (
	"fmt"
	"sort"
	"strings"

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