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

// isDataOrResourceRef returns true if the traversal root is "data" or neither var/local/module nor empty (thus presumably a resource).
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
	return true // includes "data", "aws_*", "azurerm_*", etc.
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
		// Only check if the root is "data" or a resource
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
	// If an attribute step ends with "[*]" and there's still another step after it,
	// that implies something like "aws_instance.multiple[*].id" => partial => skip
	for i := 0; i < len(trav)-1; i++ {
		if attr, ok := trav[i].(hcl.TraverseAttr); ok {
			if strings.HasSuffix(attr.Name, "[*]") {
				// There's another step after "multiple[*]" => partial reference
				return false
			}
		}
	}

	switch len(trav) {
	case 2:
		// e.g., resource.resource_name
		// But if the attribute includes a dot (e.g. "example.id"),
		// it actually references a sub-attribute in one parse step, so it's partial.
		if attr, ok := trav[1].(hcl.TraverseAttr); ok {
			// If the attribute name includes "[*].", e.g. "multiple[*].id",
                	// this indicates a splat usage plus a final attribute => partial
                	if strings.Contains(attr.Name, "[*].") {
                		return false
                	}
			if strings.Contains(attr.Name, ".") {
				return false // referencing a sub-attribute
			}
		}
		return true
	case 3:
		// If the first step is var/local/module => not a resource => skip
		if root, ok := trav[0].(hcl.TraverseRoot); ok {
			switch root.Name {
			case "var", "local", "module":
				return false
			case "data":
				// "data.<type>.<name>" => entire
				return true
			}
		}

		// If the middle step is an index or splat for "*",
		// but there's a final attribute step, then it's partial.
		// Example: aws_instance.multiple[*].id => partial, not entire.
		switch mid := trav[1].(type) {
		case hcl.TraverseIndex:
			if mid.Key.Type() == cty.String && mid.Key.AsString() == "*" {
				if _, ok := trav[2].(hcl.TraverseAttr); ok {
					return false // partial reference
				}
			}
		case hcl.TraverseSplat:
			if _, ok := trav[2].(hcl.TraverseAttr); ok {
				return false // partial reference
			}
		}

		// If the last step is an attribute, it means we're referencing a sub-attribute (partial).
		// If it's an index or a splat, it's entire.
		switch trav[2].(type) {
		case hcl.TraverseAttr:
			return false // partial reference
		default:
			return true // entire reference
		}

	default:
		// If there's a 4th step (like .id), then it's partial => skip
		return false
	}
}

// filterPrefixTraversals removes any traversal that is a strict prefix of another longer traversal.
func filterPrefixTraversals(all []hcl.Traversal) []hcl.Traversal {
	var result []hcl.Traversal

outer:
	for i, t1 := range all {
		for j, t2 := range all {
			if i == j {
				continue
			}
			if isPrefix(t1, t2) {
				// Skip t1 if it's a prefix of t2
				continue outer
			}
		}
		// If t1 is not a prefix of any longer traversal
		result = append(result, t1)
	}
	return result
}

// isPrefix returns true if t1 is strictly a prefix (same steps in order) of t2.
func isPrefix(t1, t2 hcl.Traversal) bool {
	if len(t1) >= len(t2) {
		return false
	}
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
		// If a is TraverseAttr and b is TraverseIndex with the same string key, treat them as equal
		if bIndex, ok := b.(hcl.TraverseIndex); ok {
			if bIndex.Key.Type() == cty.String {
				if bIndex.Key.AsString() == aTyped.Name {
					return true
				}
			}
		}
	case hcl.TraverseIndex:
		// If the key is "*", we consider it equivalent to a splat
		if aTyped.Key.Type() == cty.String && aTyped.Key.AsString() == "*" {
			if _, isSplat := b.(hcl.TraverseSplat); isSplat {
				return true
			}
		}
		if bIndex, ok := b.(hcl.TraverseIndex); ok {
			if aTyped.Key.RawEquals(bIndex.Key) {
				return true
			}
		}
		// If a is TraverseIndex with a string key, and b is TraverseAttr with the same string name
		if bAttr, ok := b.(hcl.TraverseAttr); ok {
			if aTyped.Key.Type() == cty.String {
				if aTyped.Key.AsString() == bAttr.Name {
					return true
				}
			}
		}
	case hcl.TraverseSplat:
		// If we have a splat, and the other side is an index with key "*",
 		// treat them as equivalent. We want them recognized as the same step
 		// for prefix filtering.
		if bIndex, ok := b.(hcl.TraverseIndex); ok {
				if bIndex.Key.Type() == cty.String && bIndex.Key.AsString() == "*" {
					return true
				}
			}
		if _, ok := b.(hcl.TraverseSplat); ok {
			return true
		}
	}
	return false
}
