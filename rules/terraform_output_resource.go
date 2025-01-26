package rules

import (
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"

	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// TerraformOutputResourceRule checks if a resource (including data) is output directly,
// rather than referencing a specific attribute. This can cause schema issues or breakage.
type TerraformOutputResourceRule struct {
	tflint.DefaultRule
}

// NewTerraformOutputResourceRule creates a new rule instance.
func NewTerraformOutputResourceRule() *TerraformOutputResourceRule {
	return &TerraformOutputResourceRule{}
}

// Name returns the rule name.
func (r *TerraformOutputResourceRule) Name() string {
	return "terraform_output_resource"
}

// Enabled returns whether the rule is enabled by default.
func (r *TerraformOutputResourceRule) Enabled() bool {
	return true
}

// Severity returns the severity of the rule.
func (r *TerraformOutputResourceRule) Severity() tflint.Severity {
	return tflint.ERROR
}

// Link returns the rule's reference link.
func (r *TerraformOutputResourceRule) Link() string {
	return ""
}

// Check checks for outputs that reference entire resources.
func (r *TerraformOutputResourceRule) Check(runner tflint.Runner) error {
	files, err := runner.GetFiles()
	if err != nil {
		return err
	}

	for filename, hclFile := range files {
		if hclFile == nil || hclFile.Bytes == nil {
			continue
		}
		// Use hclsyntax.ParseConfig instead of hcl.ParseConfig
		syntaxFile, diags := hclsyntax.ParseConfig(hclFile.Bytes, filename, hcl.InitialPos)
		if diags.HasErrors() {
			continue
		}

		// Cast syntaxFile.Body to *hclsyntax.Body
		if body, ok := syntaxFile.Body.(*hclsyntax.Body); ok {
			if err := r.processBody(body, runner); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *TerraformOutputResourceRule) processBody(body *hclsyntax.Body, runner tflint.Runner) error {
	for _, blk := range body.Blocks {
		if strings.EqualFold(blk.Type, TypeOutput) {
			if err := r.checkOutputBlock(blk, runner); err != nil {
				return err
			}
		}
		// Recurse into nested blocks
		if err := r.processBody(blk.Body, runner); err != nil {
			return err
		}
	}
	return nil
}

// isDataOrResourceRef returns true if the traversal root is "data" or
// neither var/local/module nor empty (thus presumably a resource).
// Also treat "ephemeral" as a resource root.
func isDataOrResourceRef(trav hcl.Traversal) bool {
	if len(trav) == 0 {
		return false
	}
	root, ok := trav[0].(hcl.TraverseRoot)
	if !ok {
		return false
	}
	switch root.Name {
	case "var", "local", "module":
		return false
	}
	return true // includes "data", "aws_*", "azurerm_*", "ephemeral", etc.
}

// checkOutputBlock inspects if the "value" attribute references an entire resource/data.
func (r *TerraformOutputResourceRule) checkOutputBlock(
	block *hclsyntax.Block,
	runner tflint.Runner,
) error {
	valAttr, ok := block.Body.Attributes["value"]
	if !ok {
		return nil
	}
	expr := valAttr.Expr

	// We parse the expression and see if it's a traversal referencing a resource or data.
	traversals := expr.Variables()
	if len(traversals) == 0 {
		return nil
	}

	// To avoid catching partial references, we'll skip any traversal that is a prefix of a longer one.
	filtered := filterPrefixTraversals(traversals)

	// If any of the filtered traversals is a "bare" reference => report
	for _, trav := range filtered {
		// Only check if the root is "data" or a resource (including ephemeral).
		if !isDataOrResourceRef(trav) {
			continue
		}
		if r.isEntireResourceReference(trav) {
			return runner.EmitIssue(
				r,
				"Output is referencing the entire resource or data, rather than a specific attribute. This can cause schema issues.",
				valAttr.Range(),
			)
		}
	}

	return nil
}

// isEntireResourceReference checks if the traversal looks like a bare resource reference
// e.g., "aws_instance.foo", "aws_instance.foo[0]", "aws_instance.foo[*]",
// or "data.aws_instance.foo" with no sub-attributes after the name.
func (r *TerraformOutputResourceRule) isEntireResourceReference(trav hcl.Traversal) bool {
	switch len(trav) {
	case 2:
		// e.g., resource.resource_name