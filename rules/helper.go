package rules

import (
	"fmt"

	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// GetAttributeRawText slices the file bytes from the attribute's expression range.
// That way we can parse "bool", "true", "null", "false", etc. as plain text.
func GetAttributeRawText(attr *hclsyntax.Attribute, fileBytes []byte) string {
	rng := attr.Expr.Range()
	if rng.End.Byte > len(fileBytes) {
		return ""
	}
	if rng.Start.Byte >= rng.End.Byte {
		return ""
	}
	return string(fileBytes[rng.Start.Byte:rng.End.Byte])
}

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
