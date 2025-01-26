package rules

import (
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
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
		syntaxFile, diags := hclsyntax.ParseConfig(hclFile.Bytes, filename, hcl.InitialPos)
		if diags.HasErrors() {
			continue
		}

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
		// Recurse
		if err := r.processBody(blk.Body, runner); err != nil {
			return err
		}
	}
	return nil
}

// isDataOrResourceRef returns true if the traversal root is "data" or
// neither var/local/module nor empty (thus presumably a resource).
// Also treat "ephemeral" as a resource root (per user request).
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

	// We parse the expression and see if it’s a traversal referencing a resource or data.
	traversals := expr.Variables()
	if len(traversals) == 0 {
		return nil
	}

	// To avoid catching partial references (like "aws_instance.foo" as part of
	// "aws_instance.foo.id"), we'll skip any traversal that is a prefix of a longer one.
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
// e.g., "aws_instance.foo" or "data.aws_instance.foo" with no sub-attributes after the name.
func (r *TerraformOutputResourceRule) isEntireResourceReference(trav hcl.Traversal) bool {
	// We only consider it a “bare” entire resource if:
	//   1) trav length == 2: e.g., [Root("aws_instance"), Attr("my_example")]
	//   2) trav length == 3 and trav[0] == "data": e.g., [Root("data"), Attr("aws_iam_user"), Attr("blah")]
	//   3) No indexing or extra sub-attributes
	//
	// If any step is TraverseIndex(...) or if we have more than these minimal steps,
	// it's partial or an attribute => ignore

	switch len(trav) {
	case 2:
		// e.g., resource.resource_name
		// Ensure no indexing
		if hasIndex(trav) {
			return false
		}
		return true
	case 3:
		// e.g., data.resource_type.resource_name
		root, okRoot := trav[0].(hcl.TraverseRoot)
		if !okRoot || root.Name != "data" {
			return false
		}
		// Ensure no indexing
		if hasIndex(trav) {
			return false
		}
		return true
	default:
		return false
	}
}

// hasIndex returns true if the traversal has any bracket index steps.
func hasIndex(trav hcl.Traversal) bool {
	for _, step := range trav {
		if _, ok := step.(hcl.TraverseIndex); ok {
			return true
		}
	}
	return false
}

// filterPrefixTraversals removes any traversal that is a strict prefix of another
// longer traversal. This happens, e.g., when Terraform's parser enumerates both
// "aws_instance.foo" and "aws_instance.foo.id".
func filterPrefixTraversals(all []hcl.Traversal) []hcl.Traversal {
	var result []hcl.Traversal

outer:
	for i, t1 := range all {
		for j, t2 := range all {
			if i == j {
				continue
			}
			if isPrefix(t1, t2) {
				// skip t1 if it's a prefix of t2
				continue outer
			}
		}
		// If we reach here, t1 is not a prefix of any longer traversal
		result = append(result, t1)
	}
	return result
}

// isPrefix returns true if t1 is strictly a prefix (same steps in order) of t2,
// and t2 has more steps. For example:
//   t1 = [Root("aws_instance"), Attr("foo")]
//   t2 = [Root("aws_instance"), Attr("foo"), Attr("id")]
// => isPrefix(t1, t2) == true
func isPrefix(t1, t2 hcl.Traversal) bool {
	if len(t1) >= len(t2) {
		return false
	}
	// Compare each step
	for i := range t1 {
		if !stepEqual(t1[i], t2[i]) {
			return false
		}
	}
	return true
}

// stepEqual does a basic comparison of hcl.Traverser steps
func stepEqual(a, b hcl.Traverser) bool {
	switch aTyped := a.(type) {
	case hcl.TraverseRoot:
		if bTyped, ok := b.(hcl.TraverseRoot); ok {
			return aTyped.Name == bTyped.Name
		}
	case hcl.TraverseAttr:
		if bTyped, ok := b.(hcl.TraverseAttr); ok {
			return aTyped.Name == bTyped.Name
		}
	case hcl.TraverseIndex:
		if _, ok := b.(hcl.TraverseIndex); ok {
			// We can't trivially check the Key equality, but to confirm prefix
			// we just confirm they're both index steps.
			return true
		}
	}
	return false
}
