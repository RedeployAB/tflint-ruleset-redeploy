package rules

import (
	"fmt"
	"strings"

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

// parseAttributeText is a helper that calls GetAttributeRawText and handles errors.
// If skipOnError is true, it returns an empty string and nil error when GetAttributeRawText fails.
// If skipOnError is false, it propagates the error from GetAttributeRawText.
// When successful, it returns the text trimmed and converted to lowercase.
func parseAttributeText(attr *hclsyntax.Attribute, fileBytes []byte, skipOnError bool) (string, error) {
	src, err := GetAttributeRawText(attr, fileBytes)
	if err != nil {
		if skipOnError {
			return "", nil
		}
		return "", err
	}
	return strings.ToLower(strings.TrimSpace(src)), nil
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
