package rules

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

// EvaluateBoolLiteral safely evaluates an expression as a boolean literal.
// Returns the boolean value, whether it's a literal (not a complex expression), and any error.
func EvaluateBoolLiteral(expr hcl.Expression) (value bool, isLiteral bool, err error) {
	if expr == nil {
		return false, false, fmt.Errorf("expression is nil")
	}

	// Check if it's a literal value expression
	if litExpr, ok := expr.(*hclsyntax.LiteralValueExpr); ok {
		if litExpr.Val.Type() == cty.Bool {
			return litExpr.Val.True(), true, nil
		}
		return false, false, fmt.Errorf("expression is not a boolean literal")
	}

	// Try to evaluate as a static expression
	val, diags := expr.Value(nil)
	if diags.HasErrors() {
		return false, false, fmt.Errorf("failed to evaluate expression: %s", diags.Error())
	}

	if val.Type() == cty.Bool {
		if val.IsKnown() && !val.IsNull() {
			return val.True(), false, nil
		}
		return false, false, fmt.Errorf("boolean value is unknown or null")
	}

	return false, false, fmt.Errorf("expression does not evaluate to a boolean")
}

// EvaluateStringLiteral safely evaluates an expression as a string literal.
// Returns the string value, whether it's a literal (not a complex expression), and any error.
func EvaluateStringLiteral(expr hcl.Expression) (value string, isLiteral bool, err error) {
	if expr == nil {
		return "", false, fmt.Errorf("expression is nil")
	}

	// Check if it's a literal value expression
	if litExpr, ok := expr.(*hclsyntax.LiteralValueExpr); ok {
		if litExpr.Val.Type() == cty.String {
			return litExpr.Val.AsString(), true, nil
		}
		return "", false, fmt.Errorf("expression is not a string literal")
	}

	// Check if it's a template expression with only literals
	if tmplExpr, ok := expr.(*hclsyntax.TemplateExpr); ok {
		if tmplExpr.IsStringLiteral() {
			val, diags := tmplExpr.Value(nil)
			if !diags.HasErrors() && val.Type() == cty.String {
				return val.AsString(), true, nil
			}
		}
	}

	// Try to evaluate as a static expression
	val, diags := expr.Value(nil)
	if diags.HasErrors() {
		return "", false, fmt.Errorf("failed to evaluate expression: %s", diags.Error())
	}

	if val.Type() == cty.String {
		if val.IsKnown() && !val.IsNull() {
			return val.AsString(), false, nil
		}
		return "", false, fmt.Errorf("string value is unknown or null")
	}

	return "", false, fmt.Errorf("expression does not evaluate to a string")
}

// EvaluateTypeExpr checks if an expression represents a Terraform type.
// Returns the type name, whether it's a valid type expression, and any error.
func EvaluateTypeExpr(expr hcl.Expression) (typeName string, isType bool, err error) {
	if expr == nil {
		return "", false, fmt.Errorf("expression is nil")
	}

	// Check if it's a scope traversal (e.g., bool, string, number)
	if scopeExpr, ok := expr.(*hclsyntax.ScopeTraversalExpr); ok {
		if len(scopeExpr.Traversal) == 1 {
			if root, ok := scopeExpr.Traversal[0].(hcl.TraverseRoot); ok {
				typeName := root.Name
				if isValidTerraformType(typeName) {
					return typeName, true, nil
				}
			}
		}
	}

	// Check if it's a function call (e.g., list(string), map(number))
	if funcExpr, ok := expr.(*hclsyntax.FunctionCallExpr); ok {
		funcName := funcExpr.Name
		if isValidTerraformTypeFunction(funcName) {
			// For complex types, return the raw representation
			return buildTypeString(funcExpr), true, nil
		}
	}

	// Try to get raw text representation as fallback
	if rangeExpr, ok := expr.(interface{ Range() hcl.Range }); ok {
		// This is a best-effort attempt to get type info from source
		_ = rangeExpr.Range()
		// We can't easily get the source text without file bytes here
		// This would need to be handled by the caller if needed
	}

	return "", false, fmt.Errorf("expression is not a valid type expression")
}

// IsNullLiteral checks if an expression is the literal 'null'.
func IsNullLiteral(expr hcl.Expression) (bool, error) {
	if expr == nil {
		return false, fmt.Errorf("expression is nil")
	}

	// Check if it's a literal value expression
	if litExpr, ok := expr.(*hclsyntax.LiteralValueExpr); ok {
		return litExpr.Val.IsNull(), nil
	}

	// Try to evaluate as a static expression
	val, diags := expr.Value(nil)
	if diags.HasErrors() {
		return false, fmt.Errorf("failed to evaluate expression: %s", diags.Error())
	}

	return val.IsNull(), nil
}

// EvaluateNumberLiteral safely evaluates an expression as a number literal.
// Returns the number value, whether it's a literal, and any error.
func EvaluateNumberLiteral(expr hcl.Expression) (value cty.Value, isLiteral bool, err error) {
	if expr == nil {
		return cty.NilVal, false, fmt.Errorf("expression is nil")
	}

	// Check if it's a literal value expression
	if litExpr, ok := expr.(*hclsyntax.LiteralValueExpr); ok {
		if litExpr.Val.Type() == cty.Number {
			return litExpr.Val, true, nil
		}
		return cty.NilVal, false, fmt.Errorf("expression is not a number literal")
	}

	// Check if it's a unary expression (e.g., negative numbers like -10)
	if unaryExpr, ok := expr.(*hclsyntax.UnaryOpExpr); ok {
		if unaryExpr.Op == hclsyntax.OpNegate {
			if litExpr, ok := unaryExpr.Val.(*hclsyntax.LiteralValueExpr); ok {
				if litExpr.Val.Type() == cty.Number {
					// This is a negative number literal
					val, diags := expr.Value(nil)
					if !diags.HasErrors() && val.Type() == cty.Number {
						return val, true, nil
					}
				}
			}
		}
	}

	// Try to evaluate as a static expression
	val, diags := expr.Value(nil)
	if diags.HasErrors() {
		return cty.NilVal, false, fmt.Errorf("failed to evaluate expression: %s", diags.Error())
	}

	if val.Type() == cty.Number {
		if val.IsKnown() && !val.IsNull() {
			return val, false, nil
		}
		return cty.NilVal, false, fmt.Errorf("number value is unknown or null")
	}

	return cty.NilVal, false, fmt.Errorf("expression does not evaluate to a number")
}

// IsLiteralExpression checks if an expression is a literal (not computed).
func IsLiteralExpression(expr hcl.Expression) bool {
	if expr == nil {
		return false
	}

	switch typed := expr.(type) {
	case *hclsyntax.LiteralValueExpr:
		return true
	case *hclsyntax.TemplateExpr:
		return typed.IsStringLiteral()
	case *hclsyntax.UnaryOpExpr:
		// Check if it's a negative number literal
		if typed.Op == hclsyntax.OpNegate {
			if litExpr, ok := typed.Val.(*hclsyntax.LiteralValueExpr); ok {
				return litExpr.Val.Type() == cty.Number
			}
		}
	}

	return false
}

// GetExpressionType returns the cty.Type of an expression if it can be determined statically.
func GetExpressionType(expr hcl.Expression) (cty.Type, error) {
	if expr == nil {
		return cty.NilType, fmt.Errorf("expression is nil")
	}

	// Try to evaluate the expression to get its type
	val, diags := expr.Value(nil)
	if diags.HasErrors() {
		return cty.NilType, fmt.Errorf("failed to evaluate expression: %s", diags.Error())
	}

	return val.Type(), nil
}

// EvaluateAttributeValue is a convenience function that evaluates an attribute's expression
// and returns the result with type information.
func EvaluateAttributeValue(attr *hclsyntax.Attribute) (cty.Value, bool, error) {
	if attr == nil {
		return cty.NilVal, false, fmt.Errorf("attribute is nil")
	}

	isLiteral := IsLiteralExpression(attr.Expr)
	val, diags := attr.Expr.Value(nil)
	if diags.HasErrors() {
		return cty.NilVal, isLiteral, fmt.Errorf("failed to evaluate attribute: %s", diags.Error())
	}

	return val, isLiteral, nil
}

// Helper functions for type validation

// isValidTerraformType checks if a string represents a valid Terraform primitive type.
// Type names from the HCL AST are always lowercase.
func isValidTerraformType(typeName string) bool {
	switch typeName {
	case "bool", "string", "number", "any":
		return true
	default:
		return false
	}
}

// isValidTerraformTypeFunction checks if a string represents a valid Terraform type function.
// Function names from the HCL AST are always lowercase.
func isValidTerraformTypeFunction(funcName string) bool {
	switch funcName {
	case "list", "set", "map", "object", "tuple":
		return true
	default:
		return false
	}
}

// buildTypeString constructs a type string from a function call expression.
func buildTypeString(funcExpr *hclsyntax.FunctionCallExpr) string {
	// This is a simplified version - in practice, this would need more sophisticated handling
	// of nested type expressions
	result := funcExpr.Name + "("

	for i, arg := range funcExpr.Args {
		if i > 0 {
			result += ", "
		}

		// Recursively handle nested type expressions
		if argFunc, ok := arg.(*hclsyntax.FunctionCallExpr); ok {
			result += buildTypeString(argFunc)
		} else if scopeExpr, ok := arg.(*hclsyntax.ScopeTraversalExpr); ok {
			if len(scopeExpr.Traversal) == 1 {
				if root, ok := scopeExpr.Traversal[0].(hcl.TraverseRoot); ok {
					result += root.Name
				}
			}
		} else {
			// Fallback for other expression types
			result += "unknown"
		}
	}

	result += ")"
	return result
}

// Legacy compatibility functions - these maintain backward compatibility
// with existing string-based evaluation patterns in the codebase

// EvaluateBoolLiteralFromRawText evaluates a boolean from raw text (legacy compatibility).
func EvaluateBoolLiteralFromRawText(rawText string) (bool, bool, error) {
	trimmed := strings.ToLower(strings.TrimSpace(rawText))
	switch trimmed {
	case StringTrue:
		return true, true, nil
	case StringFalse:
		return false, true, nil
	default:
		return false, false, fmt.Errorf("text %q is not a boolean literal", rawText)
	}
}

// EvaluateStringLiteralFromRawText evaluates a string from raw text (legacy compatibility).
func EvaluateStringLiteralFromRawText(rawText string) (string, bool, error) {
	trimmed := strings.TrimSpace(rawText)

	// Check if it's a quoted string
	if len(trimmed) >= 2 && trimmed[0] == '"' && trimmed[len(trimmed)-1] == '"' {
		// Remove quotes and unescape
		unquoted := trimmed[1 : len(trimmed)-1]
		// Basic unescaping - in practice, this would need more sophisticated handling
		unquoted = strings.ReplaceAll(unquoted, `\"`, `"`)
		unquoted = strings.ReplaceAll(unquoted, `\\`, `\`)
		return unquoted, true, nil
	}

	return "", false, fmt.Errorf("text %q is not a string literal", rawText)
}

// IsNullLiteralFromRawText checks if raw text represents null (legacy compatibility).
func IsNullLiteralFromRawText(rawText string) bool {
	return strings.ToLower(strings.TrimSpace(rawText)) == StringNull
}
