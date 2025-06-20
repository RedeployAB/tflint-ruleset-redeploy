package rules

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

// Used for reading test fixtures
func readFixture(t *testing.T, filename string) string {
	path := filepath.Join("testdata", filename)
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed reading %s: %v", path, err)
	}
	return string(content)
}

func TestGetAttributeRawText(t *testing.T) {
	// We'll inline a small .tf snippet, parse it, and then verify
	// that GetAttributeRawText() retrieves each attribute's raw text correctly.
	source := `
variable "test" {
	description = "Just a test"
	type        = bool
	default     = null
	sensitive   = false
}
`

	// Use the TFLint helper's TestRunner to parse this snippet as "test.tf"
	runner := helper.TestRunner(t, map[string]string{
		"test.tf": source,
	})

	files, err := runner.GetFiles()
	if err != nil {
		t.Fatalf("Unexpected error from runner.GetFiles(): %v", err)
	}

	f, ok := files["test.tf"]
	if !ok || f.Bytes == nil {
		t.Fatal("Failed to retrieve file contents for test.tf")
	}

	// Parse the file as HCL syntax
	syntaxFile, diags := hclsyntax.ParseConfig(f.Bytes, "test.tf", hcl.InitialPos)
	if diags.HasErrors() {
		t.Fatalf("Parse error: %v", diags)
	}

	body, ok := syntaxFile.Body.(*hclsyntax.Body)
	if !ok {
		t.Fatal("Parsed file body is not an *hclsyntax.Body")
	}

	// Expect exactly one block: variable "test"
	if len(body.Blocks) != 1 {
		t.Fatalf("Expected 1 block, got %d", len(body.Blocks))
	}

	variableBlock := body.Blocks[0]

	// Fetch each attribute from the variable block
	descriptionAttr := variableBlock.Body.Attributes["description"]
	typeAttr := variableBlock.Body.Attributes["type"]
	defaultAttr := variableBlock.Body.Attributes["default"]
	sensitiveAttr := variableBlock.Body.Attributes["sensitive"]

	// Check each attribute's raw text
	descText, err := GetAttributeRawText(descriptionAttr, f.Bytes)
	if err != nil {
		t.Errorf("Unexpected error from GetAttributeRawText for description: %v", err)
	}
	if descText != `"Just a test"` {
		t.Errorf("Expected description = \"Just a test\", got %q", descText)
	}

	typeText, err := GetAttributeRawText(typeAttr, f.Bytes)
	if err != nil {
		t.Errorf("Unexpected error from GetAttributeRawText for type: %v", err)
	}
	if typeText != "bool" {
		t.Errorf("Expected type = bool, got %q", typeText)
	}

	defaultText, err := GetAttributeRawText(defaultAttr, f.Bytes)
	if err != nil {
		t.Errorf("Unexpected error from GetAttributeRawText for default: %v", err)
	}
	if defaultText != "null" {
		t.Errorf("Expected default = null, got %q", defaultText)
	}

	sensitiveText, err := GetAttributeRawText(sensitiveAttr, f.Bytes)
	if err != nil {
		t.Errorf("Unexpected error from GetAttributeRawText for sensitive: %v", err)
	}
	if sensitiveText != "false" {
		t.Errorf("Expected sensitive = false, got %q", sensitiveText)
	}

	// Test error conditions
	t.Run("nil attribute", func(t *testing.T) {
		_, err := GetAttributeRawText(nil, f.Bytes)
		if err == nil {
			t.Error("Expected error for nil attribute, got none")
		}
	})

	t.Run("nil fileBytes", func(t *testing.T) {
		_, err := GetAttributeRawText(descriptionAttr, nil)
		if err == nil {
			t.Error("Expected error for nil fileBytes, got none")
		}
	})
}

func TestMax(t *testing.T) {
	tests := []struct {
		a, b     int
		expected int
	}{
		{1, 2, 2},
		{2, 1, 2},
		{-1, 1, 1},
		{0, 0, 0},
	}

	for _, tc := range tests {
		got := Max(tc.a, tc.b)
		if got != tc.expected {
			t.Errorf("Max(%d, %d) = %d; want %d", tc.a, tc.b, got, tc.expected)
		}
	}
}
