package rules

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// GetRuleDocLink returns the URL to the documentation of a rule based on its name.
func GetRuleDocLink(ruleName string) string {
	return fmt.Sprintf("https://github.com/RedeployAB/tflint-ruleset-redeploy/blob/main/docs/rules/%s.md", ruleName)
}

// evaluateBooleanAttribute evaluates an HCL expression as a boolean value.
// Returns (value, true) if the expression successfully evaluates to a boolean.
// Returns (false, false) if the expression cannot be evaluated as a boolean.
func evaluateBooleanAttribute(runner tflint.Runner, expr hcl.Expression) (bool, bool) {
	var value bool
	if err := runner.EvaluateExpr(expr, &value, nil); err != nil {
		return false, false
	}
	return value, true
}
