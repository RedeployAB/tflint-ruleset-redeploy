package rules

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// TerraformVariableRedeclaredAsLocalRule checks if a variable name (from a variable block)
// is also declared in locals, emitting an error on the conflicting local.
type TerraformVariableRedeclaredAsLocalRule struct {
	tflint.DefaultRule
}

func NewTerraformVariableRedeclaredAsLocalRule() *TerraformVariableRedeclaredAsLocalRule {
	return &TerraformVariableRedeclaredAsLocalRule{}
}

func (r *TerraformVariableRedeclaredAsLocalRule) Name() string {
	return "terraform_variable_redeclared_as_local"
}

func (r *TerraformVariableRedeclaredAsLocalRule) Enabled() bool {
	return true
}

func (r *TerraformVariableRedeclaredAsLocalRule) Severity() tflint.Severity {
	return tflint.ERROR
}

func (r *TerraformVariableRedeclaredAsLocalRule) Link() string {
	return ""
}

func (r *TerraformVariableRedeclaredAsLocalRule) Check(runner tflint.Runner) error {
	// Gather variable names from 'variable' blocks
	variableNames := make(map[string]bool)

	files, err := runner.GetFiles()
	if err != nil {
		return err
	}

	// First pass: collect all variable names
	for filename, hclFile := range files {
		if hclFile == nil || hclFile.Bytes == nil {
			continue
		}
		syntaxFile, diags := hclsyntax.ParseConfig(hclFile.Bytes, filename, hcl.InitialPos)
		if diags.HasErrors() {
			// Skip parse errors
			continue
		}
		if body, ok := syntaxFile.Body.(*hclsyntax.Body); ok {
			r.collectVariableNames(body, variableNames)
		}
	}

	// Second pass: check for conflicts in locals
	for filename, hclFile := range files {
		if hclFile == nil || hclFile.Bytes == nil {
			continue
		}
		syntaxFile, diags := hclsyntax.ParseConfig(hclFile.Bytes, filename, hcl.InitialPos)
		if diags.HasErrors() {
			continue
		}
		if body, ok := syntaxFile.Body.(*hclsyntax.Body); ok {
			if err := r.checkLocals(body, filename, variableNames, runner); err != nil {
				return err
			}
		}
	}

	return nil
}

// collectVariableNames recursively scans for 'variable' blocks and collects their names.
func (r *TerraformVariableRedeclaredAsLocalRule) collectVariableNames(
	body *hclsyntax.Body,
	variableNames map[string]bool,
) {
	for _, block := range body.Blocks {
		if strings.EqualFold(block.Type, TypeVariable) && len(block.Labels) > 0 {
			nameLabel := block.Labels[0] // variable "<name>"
			variableNames[nameLabel] = true
		}
		// Recurse into nested blocks
		r.collectVariableNames(block.Body, variableNames)
	}
}

// checkLocals scans for 'locals' blocks and checks for name conflicts with variables.
func (r *TerraformVariableRedeclaredAsLocalRule) checkLocals(
	body *hclsyntax.Body,
	filename string,
	variableNames map[string]bool,
	runner tflint.Runner,
) error {
	for _, block := range body.Blocks {
		if strings.EqualFold(block.Type, "locals") {
			// Each attribute in this block is a local variable
			for attrName, attr := range block.Body.Attributes {
				if variableNames[attrName] {
					// Conflict found
					if err := runner.EmitIssue(
						r,
						fmt.Sprintf("Local '%s' conflicts with a variable of the same name", attrName),
						attr.Range(),
					); err != nil {
						return err
					}
				}
			}
		}
		// Recurse into nested blocks
		if err := r.checkLocals(block.Body, filename, variableNames, runner); err != nil {
			return err
		}
	}
	return nil
}
