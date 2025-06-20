package rules

import (
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
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
	return GetRuleDocLink(r.Name())
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
		if strings.ToLower(block.Type) == TypeVariable {
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

func (r *TerraformVariableNullableRule) checkBoolDefaultNotNull(
	typeVal, defaultVal *hclsyntax.Attribute,
	runner tflint.Runner,
) error {
	if typeVal != nil && defaultVal != nil {
		// Use the new expression utility for type checking
		typeName, isType, err := EvaluateTypeExpr(typeVal.Expr)
		if err != nil {
			return err
		}
		if isType && typeName == TypeBool {
			// Use the new expression utility for null checking
			isDefaultNull, err := IsNullLiteral(defaultVal.Expr)
			if err != nil {
				return err
			}
			if isDefaultNull {
				return runner.EmitIssue(
					r,
					"boolean variables cannot have default = null",
					defaultVal.Range(),
				)
			}
		}
	}
	return nil
}

func (r *TerraformVariableNullableRule) checkNullDefaultAndNullableNotDeclared(
	defaultVal, nullableVal *hclsyntax.Attribute,
	runner tflint.Runner,
) error {
	// Use the new expression utility for null checking
	isDefaultNull, err := IsNullLiteral(defaultVal.Expr)
	if err != nil {
		return err
	}
	if isDefaultNull && nullableVal != nil {
		return runner.EmitIssue(
			r,
			"nullable must not be declared if default = null",
			nullableVal.Range(),
		)
	}
	return nil
}

func (r *TerraformVariableNullableRule) checkNullableFalseIfDeclared(
	nullableVal *hclsyntax.Attribute,
	runner tflint.Runner,
) error {
	// Use the new expression utility for boolean evaluation
	value, isLiteral, err := EvaluateBoolLiteral(nullableVal.Expr)
	if err != nil {
		return err
	}
	if isLiteral && value {
		return runner.EmitIssue(
			r,
			"nullable should not be set to true (the default is already true)",
			nullableVal.Range(),
		)
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
	if err := r.checkBoolDefaultNotNull(typeVal, defaultVal, runner); err != nil {
		return err
	}

	// 2) If default = null => must NOT define nullable
	if defaultVal != nil {
		if err := r.checkNullDefaultAndNullableNotDeclared(defaultVal, nullableVal, runner); err != nil {
			return err
		}
	}

	// 3) If nullable is declared => must be false
	if nullableVal != nil {
		if err := r.checkNullableFalseIfDeclared(nullableVal, runner); err != nil {
			return err
		}
	}

	return nil
}

// Note: The old string-based helper functions (isTypeBool, isAttrNull, isAttrTrue)
// have been replaced with the new expression utilities for more robust evaluation.
