package rules

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// TerraformIgnoreChangesAllRule warns against `ignore_changes = all`, which
// silently ignores drift on every attribute. Listing specific attributes keeps
// the intent explicit.
type TerraformIgnoreChangesAllRule struct {
	tflint.DefaultRule
}

func NewTerraformIgnoreChangesAllRule() *TerraformIgnoreChangesAllRule {
	return &TerraformIgnoreChangesAllRule{}
}

func (r *TerraformIgnoreChangesAllRule) Name() string {
	return "terraform_ignore_changes_all"
}

func (r *TerraformIgnoreChangesAllRule) Enabled() bool {
	return true
}

func (r *TerraformIgnoreChangesAllRule) Severity() tflint.Severity {
	return tflint.WARNING
}

func (r *TerraformIgnoreChangesAllRule) Link() string {
	return GetRuleDocLink(r.Name())
}

func (r *TerraformIgnoreChangesAllRule) Check(runner tflint.Runner) error {
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

func (r *TerraformIgnoreChangesAllRule) processBody(body *hclsyntax.Body, runner tflint.Runner) error {
	for _, block := range body.Blocks {
		if block.Type == ArgLifecycle {
			if attr, ok := block.Body.Attributes[ArgIgnoreChanges]; ok && isAllKeyword(attr.Expr) {
				if err := runner.EmitIssue(
					r,
					"Avoid 'ignore_changes = all'; list the specific attributes to ignore instead",
					attr.Range(),
				); err != nil {
					return err
				}
			}
		}
		if err := r.processBody(block.Body, runner); err != nil {
			return err
		}
	}
	return nil
}

// isAllKeyword reports whether expr is the bare `all` keyword, as used in
// `ignore_changes = all`.
func isAllKeyword(expr hclsyntax.Expression) bool {
	scope, ok := expr.(*hclsyntax.ScopeTraversalExpr)
	if !ok || len(scope.Traversal) != 1 {
		return false
	}
	root, ok := scope.Traversal[0].(hcl.TraverseRoot)
	return ok && root.Name == "all"
}
