package rules

import (
	"fmt"
	"path/filepath"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// TerraformConfigBlockFileRule enforces that any "terraform" config block is placed in terraform.tf
type TerraformConfigBlockFileRule struct {
	tflint.DefaultRule
}

func NewTerraformConfigBlockFileRule() *TerraformConfigBlockFileRule {
	return &TerraformConfigBlockFileRule{}
}

func (r *TerraformConfigBlockFileRule) Name() string {
	return "terraform_config_block_file"
}

func (r *TerraformConfigBlockFileRule) Enabled() bool {
	return true
}

func (r *TerraformConfigBlockFileRule) Severity() tflint.Severity {
	return tflint.ERROR
}

func (r *TerraformConfigBlockFileRule) Link() string {
	return GetRuleDocLink(r.Name())
}

func (r *TerraformConfigBlockFileRule) Check(runner tflint.Runner) error {
	files, err := runner.GetFiles()
	if err != nil {
		return err
	}
	for filename, hclFile := range files {
		if hclFile == nil || hclFile.Bytes == nil {
			continue
		}
		// If file is named exactly "terraform.tf", skip checks (allowed).
		if filepath.Base(filename) == FileTerraform {
			continue
		}
		syntaxFile, diags := hclsyntax.ParseConfig(hclFile.Bytes, filename, hcl.InitialPos)
		if diags.HasErrors() {
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

func (r *TerraformConfigBlockFileRule) processBody(
	body *hclsyntax.Body,
	filename string,
	runner tflint.Runner,
) error {
	for _, blk := range body.Blocks {
		// Block types are always lowercase in Terraform
		if blk.Type == TypeTerraform {
			// Found a terraform config block in the wrong file => error
			if err := runner.EmitIssue(
				r,
				fmt.Sprintf(`"terraform" config block must appear in "terraform.tf", not %q`, filepath.Base(filename)),
				blk.DefRange(),
			); err != nil {
				return err
			}
		}
		if err := r.processBody(blk.Body, filename, runner); err != nil {
			return err
		}
	}
	return nil
}
