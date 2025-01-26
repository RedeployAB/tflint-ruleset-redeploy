package rules

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// TerraformLocalsMirrorAssignmentRule checks if a local variable is assigned directly
// from a variable (var.some_var), emitting an error on the local assignment.
type TerraformLocalsMirrorAssignmentRule struct {
	tflint.DefaultRule
}

func NewTerraformLocalsMirrorAssignmentRule() *TerraformLocalsMirrorAssignmentRule {
	return &TerraformLocalsMirrorAssignmentRule{}
}

func (r *TerraformLocalsMirrorAssignmentRule) Name() string {
	return "terraform_locals_mirror_assignment"
}

func (r *TerraformLocalsMirrorAssignmentRule) Enabled() bool {
	return true
}

func (r *TerraformLocalsMirrorAssignmentRule) Severity() tflint.Severity {
	return tflint.ERROR
}

func (r *TerraformLocalsMirrorAssignmentRule) Link() string {
	return ""
}

func (r *TerraformLocalsMirrorAssignmentRule) Check(runner tflint.Runner) error {
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

	// Second pass: check for direct assignments in locals
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
func (r *TerraformLocalsMirrorAssignmentRule) collectVariableNames(
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

// checkLocals scans for 'locals' blocks and checks for locals assigned directly from variables.
func (r *TerraformLocalsMirrorAssignmentRule) checkLocals(
	body *hclsyntax.Body,
	filename string,
	variableNames map[string]bool,
	runner tflint.Runner,
) error {
	for _, block := range body.Blocks {
		if strings.EqualFold(block.Type, TypeLocals) {
			// Each attribute in this block is a local variable
			for attrName, attr := range block.Body.Attributes {
				// Check if this local is assigned *directly* from var.<something>
				scopeExpr, ok := attr.Expr.(*hclsyntax.ScopeTraversalExpr)
				if ok && len(scopeExpr.Traversal) == 2 {
					if root, ok := scopeExpr.Traversal[0].(hcl.TraverseRoot); ok && root.Name == "var" {
						if second, ok := scopeExpr.Traversal[1].(hcl.TraverseAttr); ok {
							// Emit an issue for any direct assignment local_name = var.<something>
							if err := runner.EmitIssue(
								r,
								fmt.Sprintf(
									"Local '%s' is assigned directly from variable '%s'. This should not be a simple mirror assignment.",
									attrName, second.Name,
								),
								attr.Range(),
							); err != nil {
								return err
							}
						}
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
