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

// Helper function to parse test source and extract variable block
func parseTestVariable(t *testing.T, source string) (*hclsyntax.Block, []byte) {
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

	syntaxFile, diags := hclsyntax.ParseConfig(f.Bytes, "test.tf", hcl.InitialPos)
	if diags.HasErrors() {
		t.Fatalf("Parse error: %v", diags)
	}

	body, ok := syntaxFile.Body.(*hclsyntax.Body)
	if !ok {
		t.Fatal("Parsed file body is not an *hclsyntax.Body")
	}

	if len(body.Blocks) != 1 {
		t.Fatalf("Expected 1 block, got %d", len(body.Blocks))
	}

	return body.Blocks[0], f.Bytes
}

func TestGetAttributeRawText(t *testing.T) {
	source := `
variable "test" {
	description = "Just a test"
	type        = bool
	default     = null
	sensitive   = false
}
`

	variableBlock, fileBytes := parseTestVariable(t, source)

	// Table-driven tests for attributes
	tests := []struct {
		name     string
		attrName string
		expected string
	}{
		{"description", "description", `"Just a test"`},
		{"type", "type", "bool"},
		{"default", "default", "null"},
		{"sensitive", "sensitive", "false"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			attr := variableBlock.Body.Attributes[tc.attrName]
			text, err := GetAttributeRawText(attr, fileBytes)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if text != tc.expected {
				t.Errorf("Expected %s = %s, got %q", tc.attrName, tc.expected, text)
			}
		})
	}

	// Test error conditions
	t.Run("nil attribute", func(t *testing.T) {
		_, err := GetAttributeRawText(nil, fileBytes)
		if err == nil {
			t.Error("Expected error for nil attribute, got none")
		}
	})

	t.Run("nil fileBytes", func(t *testing.T) {
		attr := variableBlock.Body.Attributes["description"]
		_, err := GetAttributeRawText(attr, nil)
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
