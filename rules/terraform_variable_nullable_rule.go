package rules

import (
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// TerraformVariableNullableRule checks for proper usage of 'nullable' in variable blocks.
type TerraformVariableNullableRule struct {
	tflint.DefaultRule
}

func NewTerraformVariableNullableRule() *TerraformVariableNullableRule {
	return &TerraformVariableNullableRule{}
}

func (r *TerraformVariableNullableRule) Name() string {
	return "terraform_variable_nullable"
}

func (r *TerraformVariableNullableRule) Enabled() bool {
	return true
}

func (r *TerraformVariableNullableRule) Severity() tflint.Severity {
	return tflint.ERROR
}

func (r *TerraformVariableNullableRule) Link() string {
	return ""
}

func (r *TerraformVariableNullableRule) Check(runner tflint.Runner) error {
	files, err := runner.GetFiles()
	if err != nil {
		return err
	}
	for filename, hclFile := range files {
		if hclFile == nil || hclFile.Bytes == nil {
			continue
		}
		syntaxFile, diags := hclsyntax.ParseConfig(hclFile.Bytes, filename, hcl.InitialPos)
		if diags.HasErrors() {
			// Skip if parse error
			continue
		}
		if body, ok := syntaxFile.Body.(*hclsyntax.Body); ok {
			if err := r.processBody(body, filename, runner); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *TerraformVariableNullableRule) processBody(
	body *hclsyntax.Body,
	filename string,
	runner tflint.Runner,
) error {
	for _, block := range body.Blocks {
		// Only examine blocks of type "variable"
		if strings.ToLower(block.Type) == "variable" {
			if err := r.checkVariableBlock(block, runner); err != nil {
				return err
			}
		}
		// Recurse into nested blocks
		if err := r.processBody(block.Body, filename, runner); err != nil {
			return err
		}
	}
	return nil
}

func (r *TerraformVariableNullableRule) checkVariableBlock(
	block *hclsyntax.Block,
	runner tflint.Runner,
) error {
	var defaultVal *hclsyntax.Attribute
	var nullableVal *hclsyntax.Attribute
	var typeVal *hclsyntax.Attribute

	// We'll need the raw file bytes for slicing
	files, err := runner.GetFiles()
	if err != nil {
		return err
	}
	fileBytes := files[block.DefRange().Filename].Bytes

	// Gather relevant attributes
	for _, attr := range block.Body.Attributes {
		switch strings.ToLower(attr.Name) {
		case "default":
			defaultVal = attr
		case "nullable":
			nullableVal = attr
		case "type":
			typeVal = attr
		}
	}

	// 1) If type = bool => default must not be null
	if typeVal != nil && defaultVal != nil {
		if isBool, _ := isTypeBool(typeVal, fileBytes); isBool {
			if isDefaultNull, _ := isAttrNull(defaultVal, fileBytes); isDefaultNull {
				return runner.EmitIssue(
					r,
					"boolean variables cannot have default = null",
					defaultVal.Range(),
				)
			}
		}
	}

	// 2) If default = null => must NOT define nullable
	if defaultVal != nil {
		if isDefaultNull, _ := isAttrNull(defaultVal, fileBytes); isDefaultNull {
			if nullableVal != nil {
				return runner.EmitIssue(
					r,
					"nullable must not be declared if default = null",
					nullableVal.Range(),
				)
			}
		}
	}

	// 3) If nullable is declared => must be false
	if nullableVal != nil {
		if isNullableTrue, _ := isAttrTrue(nullableVal, fileBytes); isNullableTrue {
			return runner.EmitIssue(
				r,
				"nullable should not be set to true (the default is already true)",
				nullableVal.Range(),
			)
		}
	}
	return nil
}

// We'll do simple textual checks by slicing the file bytes from the attribute’s expression Range.
func isTypeBool(attr *hclsyntax.Attribute, fileBytes []byte) (bool, error) {
	src := getAttributeRawText(attr, fileBytes)
	src = strings.ToLower(strings.TrimSpace(src))
	return (src == "bool"), nil
}

func isAttrNull(attr *hclsyntax.Attribute, fileBytes []byte) (bool, error) {
	src := getAttributeRawText(attr, fileBytes)
	src = strings.ToLower(strings.TrimSpace(src))
	return (src == "null"), nil
}

func isAttrTrue(attr *hclsyntax.Attribute, fileBytes []byte) (bool, error) {
	src := getAttributeRawText(attr, fileBytes)
	src = strings.ToLower(strings.TrimSpace(src))
	return (src == "true"), nil
}

// getAttributeRawText slices the original file bytes from the attribute’s expression range.
// Then we can match "bool", "true", "null", etc. as plain text.
func getAttributeRawText(attr *hclsyntax.Attribute, fileBytes []byte) string {
	rng := attr.Expr.Range()

	// Validate we don't go out-of-bounds
	if rng.End.Byte > len(fileBytes) {
		return ""
	}
	if rng.Start.Byte >= rng.End.Byte {
		return ""
	}

	return string(fileBytes[rng.Start.Byte:rng.End.Byte])
}
