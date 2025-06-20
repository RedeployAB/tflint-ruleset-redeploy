package rules

import (
	"fmt"
)

// GetRuleDocLink returns the URL to the documentation of a rule based on its name.
func GetRuleDocLink(ruleName string) string {
	return fmt.Sprintf("https://github.com/RedeployAB/tflint-ruleset-redeploy/blob/main/docs/rules/%s.md", ruleName)
}

// Max returns the maximum of two integers
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
