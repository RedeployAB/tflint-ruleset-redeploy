package rules

import (
	"fmt"

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

func (*TerraformLocalsMirrorAssignmentRule) Name() string {
	return "terraform_locals_mirror_assignment"
}

func (*TerraformLocalsMirrorAssignmentRule) Enabled() bool {
	return true
}

func (*TerraformLocalsMirrorAssignmentRule) Severity() tflint.Severity {
	return tflint.ERROR
}

func (r *TerraformLocalsMirrorAssignmentRule) Link() string {
	return GetRuleDocLink(r.Name())
}

func (r *TerraformLocalsMirrorAssignmentRule) Check(runner tflint.Runner) error {
	files, err := runner.GetFiles()
	if err != nil {
		return err
	}

	// Parse files once and store parsed bodies
	type parsedFile struct {
		filename string
		body     *hclsyntax.Body
	}
	var parsedFiles []parsedFile

	// Gather variable names from 'variable' blocks
	variableNames := make(map[string]bool)

	// Single pass: parse files and collect variable names
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
			parsedFiles = append(parsedFiles, parsedFile{filename: filename, body: body})
			r.collectVariableNames(body, variableNames)
		}
	}

	// Check for direct assignments in locals using already parsed files
	for _, pf := range parsedFiles {
		if err := r.checkLocals(pf.body, pf.filename, variableNames, runner); err != nil {
			return err
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
		// Block types are always lowercase in Terraform
		if block.Type == TypeVariable && len(block.Labels) > 0 {
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
		// Block types are always lowercase in Terraform
		if block.Type == TypeLocals {
			if err := r.checkLocalsBlock(block, runner); err != nil {
				return err
			}
		}
		// Recurse into nested blocks
		if err := r.checkLocals(block.Body, filename, variableNames, runner); err != nil {
			return err
		}
	}
	return nil
}

// checkLocalsBlock emits an issue, with an autofix that removes the line, for
// every local in the block that is a plain mirror of var.<something>.
func (r *TerraformLocalsMirrorAssignmentRule) checkLocalsBlock(
	block *hclsyntax.Block,
	runner tflint.Runner,
) error {
	// Each attribute in this block is a local variable
	for attrName, attr := range block.Body.Attributes {
		variableName, ok := mirroredVariableName(attr.Expr)
		if !ok {
			continue
		}
		// Emit an issue with autofix for any direct assignment local_name = var.<something>
		err := runner.EmitIssueWithFix(
			r,
			fmt.Sprintf(
				"Local '%s' is assigned directly from variable '%s'. "+
					"This should not be a simple mirror assignment.",
				attrName, variableName,
			),
			attr.Range(),
			func(f tflint.Fixer) error {
				return removeAttributeLine(f, runner, attr.Range())
			},
		)
		if err != nil {
			return err
		}
	}
	return nil
}

// mirroredVariableName reports whether expr is *exactly* var.<name> and, if
// so, returns <name>.
func mirroredVariableName(expr hclsyntax.Expression) (name string, ok bool) {
	scopeExpr, ok := expr.(*hclsyntax.ScopeTraversalExpr)
	if !ok || len(scopeExpr.Traversal) != 2 {
		return "", false
	}
	root, ok := scopeExpr.Traversal[0].(hcl.TraverseRoot)
	if !ok || root.Name != TypeVar {
		return "", false
	}
	second, ok := scopeExpr.Traversal[1].(hcl.TraverseAttr)
	if !ok {
		return "", false
	}
	return second.Name, true
}
