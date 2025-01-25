package rules

import (
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
