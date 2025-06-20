package rules

import (
	"fmt"

	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// GetAttributeRawText slices the file bytes from the attribute's expression range.
// That way we can parse "bool", "true", "null", "false", etc. as plain text.
// Returns an error if the attribute is nil, the range is invalid, or bounds checking fails.
func GetAttributeRawText(attr *hclsyntax.Attribute, fileBytes []byte) (string, error) {
	if attr == nil {
		return "", fmt.Errorf("attribute is nil")
	}
	if attr.Expr == nil {
		return "", fmt.Errorf("attribute expression is nil")
	}
	if fileBytes == nil {
		return "", fmt.Errorf("fileBytes is nil")
	}

	rng := attr.Expr.Range()
	if rng.End.Byte > len(fileBytes) {
		return "", fmt.Errorf("expression end byte (%d) exceeds file length (%d)", rng.End.Byte, len(fileBytes))
	}
	if rng.Start.Byte >= rng.End.Byte {
		return "", fmt.Errorf("invalid range: start byte (%d) >= end byte (%d)", rng.Start.Byte, rng.End.Byte)
	}
	if rng.Start.Byte < 0 {
		return "", fmt.Errorf("expression start byte (%d) is negative", rng.Start.Byte)
	}

	return string(fileBytes[rng.Start.Byte:rng.End.Byte]), nil
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
