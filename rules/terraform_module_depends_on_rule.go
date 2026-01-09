package rules

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// TerraformModuleDependsOnRule warns if 'depends_on' is used in a module block.
type TerraformModuleDependsOnRule struct {
	tflint.DefaultRule
}

func NewTerraformModuleDependsOnRule() *TerraformModuleDependsOnRule {
	return &TerraformModuleDependsOnRule{}
}

func (r *TerraformModuleDependsOnRule) Name() string {
	return "terraform_module_depends_on"
}

func (r *TerraformModuleDependsOnRule) Enabled() bool {
	return true
}

func (r *TerraformModuleDependsOnRule) Severity() tflint.Severity {
	return tflint.WARNING
}

func (r *TerraformModuleDependsOnRule) Link() string {
	return GetRuleDocLink(r.Name())
}

func (r *TerraformModuleDependsOnRule) Check(runner tflint.Runner) error {
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

func (r *TerraformModuleDependsOnRule) processBody(body *hclsyntax.Body, runner tflint.Runner) error {
	for _, block := range body.Blocks {
		// Block types are always lowercase in Terraform
		if block.Type == TypeModule {
			// Check if it has an attribute named 'depends_on'
			for _, attr := range block.Body.Attributes {
				// Attribute names are also lowercase in Terraform
				if attr.Name == ArgDependsOn {
					rng := attr.Range()
					if err := runner.EmitIssue(
						r,
						"'depends_on' should not be used for modules",
						rng,
					); err != nil {
						return err
					}
				}
			}
		}
		// Recursively process nested blocks
		if err := r.processBody(block.Body, runner); err != nil {
			return err
		}
	}
	return nil
}
