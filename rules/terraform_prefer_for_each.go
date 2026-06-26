package rules

import (
	"math/big"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/zclconf/go-cty/cty"
)

// TerraformPreferForEachRule discourages using `count` to create multiple
// near-identical instances, recommending `for_each` instead. To avoid false
// positives it only flags `count` expressions that unambiguously create more
// than one instance (a literal >= 2 or a length() call, including within a
// ternary branch). Conditional 0/1 toggles and bare references are left alone.
type TerraformPreferForEachRule struct {
	tflint.DefaultRule
}

func NewTerraformPreferForEachRule() *TerraformPreferForEachRule {
	return &TerraformPreferForEachRule{}
}

func (r *TerraformPreferForEachRule) Name() string {
	return "terraform_prefer_for_each"
}

func (r *TerraformPreferForEachRule) Enabled() bool {
	return true
}

func (r *TerraformPreferForEachRule) Severity() tflint.Severity {
	return tflint.WARNING
}

func (r *TerraformPreferForEachRule) Link() string {
	return GetRuleDocLink(r.Name())
}

func (r *TerraformPreferForEachRule) Check(runner tflint.Runner) error {
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
		body, ok := syntaxFile.Body.(*hclsyntax.Body)
		if !ok {
			continue
		}
		if err := r.checkBlocks(body, runner); err != nil {
			return err
		}
	}
	return nil
}

func (r *TerraformPreferForEachRule) checkBlocks(body *hclsyntax.Body, runner tflint.Runner) error {
	for _, block := range body.Blocks {
		switch block.Type {
		case TypeResource, TypeData, TypeModule:
		default:
			continue
		}
		countAttr, ok := block.Body.Attributes[ArgCount]
		if !ok {
			continue
		}
		if !createsMultipleInstances(countAttr.Expr) {
			continue
		}
		if err := runner.EmitIssue(
			r,
			"Use 'for_each' instead of 'count' to create multiple instances",
			countAttr.Expr.Range(),
		); err != nil {
			return err
		}
	}
	return nil
}

// createsMultipleInstances reports whether a count expression unambiguously
// produces more than one instance: a numeric literal >= 2, a length() call, or
// either of those reached through parentheses or a ternary result branch.
func createsMultipleInstances(expr hclsyntax.Expression) bool {
	switch e := expr.(type) {
	case *hclsyntax.ParenthesesExpr:
		return createsMultipleInstances(e.Expression)
	case *hclsyntax.ConditionalExpr:
		// Only the result branches affect the instance count; the condition
		// may freely call length() without implying multiple instances.
		return createsMultipleInstances(e.TrueResult) || createsMultipleInstances(e.FalseResult)
	case *hclsyntax.FunctionCallExpr:
		return e.Name == "length"
	case *hclsyntax.LiteralValueExpr:
		val := e.Val
		if val.Type() != cty.Number || val.IsNull() {
			return false
		}
		// Compare with arbitrary precision; cty numbers are big.Float-backed,
		// so converting to float64 could misclassify values near the threshold.
		return val.AsBigFloat().Cmp(big.NewFloat(2)) >= 0
	default:
		return false
	}
}
