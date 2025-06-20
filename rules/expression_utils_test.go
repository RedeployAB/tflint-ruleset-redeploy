package rules

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
	"github.com/zclconf/go-cty/cty"
)

func TestEvaluateBoolLiteral(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		attrName    string
		expected    bool
		expectLit   bool
		expectError bool
	}{
		{
			name:        "true literal",
			source:      `test = true`,
			attrName:    "test",
			expected:    true,
			expectLit:   true,
			expectError: false,
		},
		{
			name:        "false literal",
			source:      `test = false`,
			attrName:    "test",
			expected:    false,
			expectLit:   true,
			expectError: false,
		},
		{
			name:        "string literal should error",
			source:      `test = "true"`,
			attrName:    "test",
			expected:    false,
			expectLit:   false,
			expectError: true,
		},
		{
			name:        "number literal should error",
			source:      `test = 42`,
			attrName:    "test",
			expected:    false,
			expectLit:   false,
			expectError: true,
		},
		{
			name:        "null literal should error",
			source:      `test = null`,
			attrName:    "test",
			expected:    false,
			expectLit:   false,
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runner := helper.TestRunner(t, map[string]string{
				"test.tf": tc.source,
			})

			attr := getAttributeFromRunner(t, runner, tc.attrName)
			value, isLiteral, err := EvaluateBoolLiteral(attr.Expr)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error, but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if value != tc.expected {
				t.Errorf("Expected value %v, got %v", tc.expected, value)
			}

			if isLiteral != tc.expectLit {
				t.Errorf("Expected isLiteral %v, got %v", tc.expectLit, isLiteral)
			}
		})
	}
}

func TestEvaluateStringLiteral(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		attrName    string
		expected    string
		expectLit   bool
		expectError bool
	}{
		{
			name:        "simple string literal",
			source:      `test = "hello"`,
			attrName:    "test",
			expected:    "hello",
			expectLit:   true,
			expectError: false,
		},
		{
			name:        "empty string literal",
			source:      `test = ""`,
			attrName:    "test",
			expected:    "",
			expectLit:   true,
			expectError: false,
		},
		{
			name:        "string with escapes",
			source:      `test = "hello\nworld"`,
			attrName:    "test",
			expected:    "hello\nworld",
			expectLit:   true,
			expectError: false,
		},
		{
			name:        "boolean literal should error",
			source:      `test = true`,
			attrName:    "test",
			expected:    "",
			expectLit:   false,
			expectError: true,
		},
		{
			name:        "number literal should error",
			source:      `test = 42`,
			attrName:    "test",
			expected:    "",
			expectLit:   false,
			expectError: true,
		},
		{
			name:        "null literal should error",
			source:      `test = null`,
			attrName:    "test",
			expected:    "",
			expectLit:   false,
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runner := helper.TestRunner(t, map[string]string{
				"test.tf": tc.source,
			})

			attr := getAttributeFromRunner(t, runner, tc.attrName)
			value, isLiteral, err := EvaluateStringLiteral(attr.Expr)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error, but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if value != tc.expected {
				t.Errorf("Expected value %q, got %q", tc.expected, value)
			}

			if isLiteral != tc.expectLit {
				t.Errorf("Expected isLiteral %v, got %v", tc.expectLit, isLiteral)
			}
		})
	}
}

func TestEvaluateTypeExpr(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		attrName    string
		expected    string
		expectType  bool
		expectError bool
	}{
		{
			name:        "bool type",
			source:      `test = bool`,
			attrName:    "test",
			expected:    "bool",
			expectType:  true,
			expectError: false,
		},
		{
			name:        "string type",
			source:      `test = string`,
			attrName:    "test",
			expected:    "string",
			expectType:  true,
			expectError: false,
		},
		{
			name:        "number type",
			source:      `test = number`,
			attrName:    "test",
			expected:    "number",
			expectType:  true,
			expectError: false,
		},
		{
			name:        "any type",
			source:      `test = any`,
			attrName:    "test",
			expected:    "any",
			expectType:  true,
			expectError: false,
		},
		{
			name:        "list type function",
			source:      `test = list(string)`,
			attrName:    "test",
			expected:    "list(string)",
			expectType:  true,
			expectError: false,
		},
		{
			name:        "map type function",
			source:      `test = map(number)`,
			attrName:    "test",
			expected:    "map(number)",
			expectType:  true,
			expectError: false,
		},
		{
			name:        "set type function",
			source:      `test = set(string)`,
			attrName:    "test",
			expected:    "set(string)",
			expectType:  true,
			expectError: false,
		},
		{
			name:        "string literal should error",
			source:      `test = "bool"`,
			attrName:    "test",
			expected:    "",
			expectType:  false,
			expectError: true,
		},
		{
			name:        "boolean literal should error",
			source:      `test = true`,
			attrName:    "test",
			expected:    "",
			expectType:  false,
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runner := helper.TestRunner(t, map[string]string{
				"test.tf": tc.source,
			})

			attr := getAttributeFromRunner(t, runner, tc.attrName)
			typeName, isType, err := EvaluateTypeExpr(attr.Expr)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error, but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if typeName != tc.expected {
				t.Errorf("Expected type name %q, got %q", tc.expected, typeName)
			}

			if isType != tc.expectType {
				t.Errorf("Expected isType %v, got %v", tc.expectType, isType)
			}
		})
	}
}

func TestIsNullLiteral(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		attrName    string
		expected    bool
		expectError bool
	}{
		{
			name:        "null literal",
			source:      `test = null`,
			attrName:    "test",
			expected:    true,
			expectError: false,
		},
		{
			name:        "boolean true literal",
			source:      `test = true`,
			attrName:    "test",
			expected:    false,
			expectError: false,
		},
		{
			name:        "boolean false literal",
			source:      `test = false`,
			attrName:    "test",
			expected:    false,
			expectError: false,
		},
		{
			name:        "string literal",
			source:      `test = "null"`,
			attrName:    "test",
			expected:    false,
			expectError: false,
		},
		{
			name:        "number literal",
			source:      `test = 0`,
			attrName:    "test",
			expected:    false,
			expectError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runner := helper.TestRunner(t, map[string]string{
				"test.tf": tc.source,
			})

			attr := getAttributeFromRunner(t, runner, tc.attrName)
			isNull, err := IsNullLiteral(attr.Expr)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error, but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if isNull != tc.expected {
				t.Errorf("Expected isNull %v, got %v", tc.expected, isNull)
			}
		})
	}
}

func TestEvaluateNumberLiteral(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		attrName    string
		expected    int64
		expectLit   bool
		expectError bool
	}{
		{
			name:        "positive integer",
			source:      `test = 42`,
			attrName:    "test",
			expected:    42,
			expectLit:   true,
			expectError: false,
		},
		{
			name:        "negative integer",
			source:      `test = -10`,
			attrName:    "test",
			expected:    -10,
			expectLit:   true,
			expectError: false,
		},
		{
			name:        "zero",
			source:      `test = 0`,
			attrName:    "test",
			expected:    0,
			expectLit:   true,
			expectError: false,
		},
		{
			name:        "string literal should error",
			source:      `test = "42"`,
			attrName:    "test",
			expected:    0,
			expectLit:   false,
			expectError: true,
		},
		{
			name:        "boolean literal should error",
			source:      `test = true`,
			attrName:    "test",
			expected:    0,
			expectLit:   false,
			expectError: true,
		},
		{
			name:        "null literal should error",
			source:      `test = null`,
			attrName:    "test",
			expected:    0,
			expectLit:   false,
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runner := helper.TestRunner(t, map[string]string{
				"test.tf": tc.source,
			})

			attr := getAttributeFromRunner(t, runner, tc.attrName)
			value, isLiteral, err := EvaluateNumberLiteral(attr.Expr)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error, but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if value.Type() != cty.Number {
				t.Errorf("Expected number type, got %s", value.Type().FriendlyName())
				return
			}

			intValue, _ := value.AsBigFloat().Int64()
			if intValue != tc.expected {
				t.Errorf("Expected value %d, got %d", tc.expected, intValue)
			}

			if isLiteral != tc.expectLit {
				t.Errorf("Expected isLiteral %v, got %v", tc.expectLit, isLiteral)
			}
		})
	}
}

func TestIsLiteralExpression(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		attrName string
		expected bool
	}{
		{
			name:     "boolean literal",
			source:   `test = true`,
			attrName: "test",
			expected: true,
		},
		{
			name:     "string literal",
			source:   `test = "hello"`,
			attrName: "test",
			expected: true,
		},
		{
			name:     "number literal",
			source:   `test = 42`,
			attrName: "test",
			expected: true,
		},
		{
			name:     "null literal",
			source:   `test = null`,
			attrName: "test",
			expected: true,
		},
		// Note: Function calls and variable references would not be literals
		// but they're harder to test without more complex setup
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runner := helper.TestRunner(t, map[string]string{
				"test.tf": tc.source,
			})

			attr := getAttributeFromRunner(t, runner, tc.attrName)
			isLiteral := IsLiteralExpression(attr.Expr)

			if isLiteral != tc.expected {
				t.Errorf("Expected isLiteral %v, got %v", tc.expected, isLiteral)
			}
		})
	}
}

func TestGetExpressionType(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		attrName    string
		expected    cty.Type
		expectError bool
	}{
		{
			name:        "boolean expression",
			source:      `test = true`,
			attrName:    "test",
			expected:    cty.Bool,
			expectError: false,
		},
		{
			name:        "string expression",
			source:      `test = "hello"`,
			attrName:    "test",
			expected:    cty.String,
			expectError: false,
		},
		{
			name:        "number expression",
			source:      `test = 42`,
			attrName:    "test",
			expected:    cty.Number,
			expectError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runner := helper.TestRunner(t, map[string]string{
				"test.tf": tc.source,
			})

			attr := getAttributeFromRunner(t, runner, tc.attrName)
			exprType, err := GetExpressionType(attr.Expr)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error, but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if !exprType.Equals(tc.expected) {
				t.Errorf("Expected type %s, got %s", tc.expected.FriendlyName(), exprType.FriendlyName())
			}
		})
	}
}

func TestEvaluateAttributeValue(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		attrName    string
		expectedVal interface{}
		expectLit   bool
		expectError bool
	}{
		{
			name:        "boolean attribute",
			source:      `test = true`,
			attrName:    "test",
			expectedVal: true,
			expectLit:   true,
			expectError: false,
		},
		{
			name:        "string attribute",
			source:      `test = "hello"`,
			attrName:    "test",
			expectedVal: "hello",
			expectLit:   true,
			expectError: false,
		},
		{
			name:        "number attribute",
			source:      `test = 42`,
			attrName:    "test",
			expectedVal: 42,
			expectLit:   true,
			expectError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runner := helper.TestRunner(t, map[string]string{
				"test.tf": tc.source,
			})

			attr := getAttributeFromRunner(t, runner, tc.attrName)
			value, isLiteral, err := EvaluateAttributeValue(attr)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error, but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Check the value based on its type
			switch tc.expectedVal.(type) {
			case bool:
				if value.Type() != cty.Bool || value.True() != tc.expectedVal.(bool) {
					t.Errorf("Expected boolean value %v, got %v", tc.expectedVal, value)
				}
			case string:
				if value.Type() != cty.String || value.AsString() != tc.expectedVal.(string) {
					t.Errorf("Expected string value %q, got %v", tc.expectedVal, value)
				}
			case int:
				if value.Type() != cty.Number {
					t.Errorf("Expected number type, got %s", value.Type().FriendlyName())
				} else {
					intVal, _ := value.AsBigFloat().Int64()
					if intVal != int64(tc.expectedVal.(int)) {
						t.Errorf("Expected number value %d, got %d", tc.expectedVal, intVal)
					}
				}
			}

			if isLiteral != tc.expectLit {
				t.Errorf("Expected isLiteral %v, got %v", tc.expectLit, isLiteral)
			}
		})
	}
}

// Test legacy compatibility functions

func TestEvaluateBoolLiteralFromRawText(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    bool
		expectLit   bool
		expectError bool
	}{
		{
			name:        "true",
			input:       "true",
			expected:    true,
			expectLit:   true,
			expectError: false,
		},
		{
			name:        "false",
			input:       "false",
			expected:    false,
			expectLit:   true,
			expectError: false,
		},
		{
			name:        "TRUE (case insensitive)",
			input:       "TRUE",
			expected:    true,
			expectLit:   true,
			expectError: false,
		},
		{
			name:        "False (case insensitive)",
			input:       "False",
			expected:    false,
			expectLit:   true,
			expectError: false,
		},
		{
			name:        "with whitespace",
			input:       "  true  ",
			expected:    true,
			expectLit:   true,
			expectError: false,
		},
		{
			name:        "invalid value",
			input:       "yes",
			expected:    false,
			expectLit:   false,
			expectError: true,
		},
		{
			name:        "numeric value",
			input:       "1",
			expected:    false,
			expectLit:   false,
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			value, isLiteral, err := EvaluateBoolLiteralFromRawText(tc.input)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error, but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if value != tc.expected {
				t.Errorf("Expected value %v, got %v", tc.expected, value)
			}

			if isLiteral != tc.expectLit {
				t.Errorf("Expected isLiteral %v, got %v", tc.expectLit, isLiteral)
			}
		})
	}
}

func TestEvaluateStringLiteralFromRawText(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		expectLit   bool
		expectError bool
	}{
		{
			name:        "simple quoted string",
			input:       `"hello"`,
			expected:    "hello",
			expectLit:   true,
			expectError: false,
		},
		{
			name:        "empty quoted string",
			input:       `""`,
			expected:    "",
			expectLit:   true,
			expectError: false,
		},
		{
			name:        "quoted string with escapes",
			input:       `"hello\"world"`,
			expected:    `hello"world`,
			expectLit:   true,
			expectError: false,
		},
		{
			name:        "quoted string with backslashes",
			input:       `"path\\to\\file"`,
			expected:    `path\to\file`,
			expectLit:   true,
			expectError: false,
		},
		{
			name:        "with whitespace",
			input:       `  "hello"  `,
			expected:    "hello",
			expectLit:   true,
			expectError: false,
		},
		{
			name:        "unquoted string",
			input:       "hello",
			expected:    "",
			expectLit:   false,
			expectError: true,
		},
		{
			name:        "partially quoted",
			input:       `"hello`,
			expected:    "",
			expectLit:   false,
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			value, isLiteral, err := EvaluateStringLiteralFromRawText(tc.input)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error, but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if value != tc.expected {
				t.Errorf("Expected value %q, got %q", tc.expected, value)
			}

			if isLiteral != tc.expectLit {
				t.Errorf("Expected isLiteral %v, got %v", tc.expectLit, isLiteral)
			}
		})
	}
}

func TestIsNullLiteralFromRawText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "null",
			input:    "null",
			expected: true,
		},
		{
			name:     "NULL (case insensitive)",
			input:    "NULL",
			expected: true,
		},
		{
			name:     "Null (case insensitive)",
			input:    "Null",
			expected: true,
		},
		{
			name:     "with whitespace",
			input:    "  null  ",
			expected: true,
		},
		{
			name:     "quoted null",
			input:    `"null"`,
			expected: false,
		},
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "other value",
			input:    "false",
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := IsNullLiteralFromRawText(tc.input)

			if result != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

// Test error cases

func TestEvaluateBoolLiteral_NilExpression(t *testing.T) {
	_, _, err := EvaluateBoolLiteral(nil)
	if err == nil {
		t.Error("Expected error for nil expression, but got none")
	}
}

func TestEvaluateStringLiteral_NilExpression(t *testing.T) {
	_, _, err := EvaluateStringLiteral(nil)
	if err == nil {
		t.Error("Expected error for nil expression, but got none")
	}
}

func TestEvaluateTypeExpr_NilExpression(t *testing.T) {
	_, _, err := EvaluateTypeExpr(nil)
	if err == nil {
		t.Error("Expected error for nil expression, but got none")
	}
}

func TestIsNullLiteral_NilExpression(t *testing.T) {
	_, err := IsNullLiteral(nil)
	if err == nil {
		t.Error("Expected error for nil expression, but got none")
	}
}

func TestEvaluateNumberLiteral_NilExpression(t *testing.T) {
	_, _, err := EvaluateNumberLiteral(nil)
	if err == nil {
		t.Error("Expected error for nil expression, but got none")
	}
}

func TestGetExpressionType_NilExpression(t *testing.T) {
	_, err := GetExpressionType(nil)
	if err == nil {
		t.Error("Expected error for nil expression, but got none")
	}
}

func TestEvaluateAttributeValue_NilAttribute(t *testing.T) {
	_, _, err := EvaluateAttributeValue(nil)
	if err == nil {
		t.Error("Expected error for nil attribute, but got none")
	}
}

func TestIsLiteralExpression_NilExpression(t *testing.T) {
	result := IsLiteralExpression(nil)
	if result {
		t.Error("Expected false for nil expression, but got true")
	}
}

// Helper function to extract attribute from test runner
func getAttributeFromRunner(t *testing.T, runner *helper.Runner, attrName string) *hclsyntax.Attribute {
	files, err := runner.GetFiles()
	if err != nil {
		t.Fatalf("Failed to get files: %v", err)
	}

	file, ok := files["test.tf"]
	if !ok || file.Bytes == nil {
		t.Fatal("Failed to get test.tf file")
	}

	syntaxFile, diags := hclsyntax.ParseConfig(file.Bytes, "test.tf", hcl.InitialPos)
	if diags.HasErrors() {
		t.Fatalf("Parse error: %v", diags)
	}

	body, ok := syntaxFile.Body.(*hclsyntax.Body)
	if !ok {
		t.Fatal("Body is not *hclsyntax.Body")
	}

	attr, exists := body.Attributes[attrName]
	if !exists {
		t.Fatalf("Attribute %q not found", attrName)
	}

	return attr
}