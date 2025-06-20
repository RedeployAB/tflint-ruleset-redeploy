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
		isBool, err := isTypeBool(typeVal, runner)
		if err != nil {
			return err
		}
		if isBool {
			isDefaultNull, err := isAttrNull(defaultVal)
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
	isDefaultNull, err := isAttrNull(defaultVal)
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
	isNullableTrue, err := isAttrTrue(nullableVal, runner)
	if err != nil {
		return err
	}
	if isNullableTrue {
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

// We'll do proper HCL expression evaluation.
func isTypeBool(attr *hclsyntax.Attribute, runner tflint.Runner) (bool, error) {
	// Type expressions like "bool" are not evaluated as strings, they are identifiers
	// We need to check the expression type directly
	if _, ok := attr.Expr.(*hclsyntax.LiteralValueExpr); ok {
		// If it's a literal value, check if it's the string "bool"
		var val string
		if err := runner.EvaluateExpr(attr.Expr, &val, nil); err == nil {
			return strings.ToLower(strings.TrimSpace(val)) == "bool", nil
		}
	} else if scopeExpr, ok := attr.Expr.(*hclsyntax.ScopeTraversalExpr); ok {
		// For type expressions like `bool`, check the traversal
		if len(scopeExpr.Traversal) == 1 {
			if root, ok := scopeExpr.Traversal[0].(hcl.TraverseRoot); ok {
				return strings.ToLower(root.Name) == "bool", nil
			}
		}
	}
	return false, nil
}

func isAttrNull(attr *hclsyntax.Attribute) (bool, error) {
	// Check if the expression is a literal null
	if expr, ok := attr.Expr.(*hclsyntax.LiteralValueExpr); ok {
		// The Val field is a cty.Value - check if it's null
		return expr.Val.IsNull(), nil
	}
	return false, nil
}

func isAttrTrue(attr *hclsyntax.Attribute, runner tflint.Runner) (bool, error) {
	var value bool
	if err := runner.EvaluateExpr(attr.Expr, &value, nil); err != nil {
		// For non-boolean expressions, return false
		return false, nil
	}
	return value, nil
}
