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
			blocks := r.collectOutputBlocks(body)
			if err := r.checkAllOutputBlocks(blocks, runner); err != nil {
				return err
			}
		}
	}
	return nil
}

// collectOutputBlocks gathers all "output" blocks from the HCL body recursively.
func (r *TerraformOutputResourceRule) collectOutputBlocks(body *hclsyntax.Body) []*hclsyntax.Block {
	var outputs []*hclsyntax.Block

	for _, blk := range body.Blocks {
		if strings.EqualFold(blk.Type, TypeOutput) {
			outputs = append(outputs, blk)
		}
		nested := r.collectOutputBlocks(blk.Body)
		outputs = append(outputs, nested...)
	}
	return outputs
}

// checkAllOutputBlocks checks each output block for entire resource references.
func (r *TerraformOutputResourceRule) checkAllOutputBlocks(
	blocks []*hclsyntax.Block,
	runner tflint.Runner,
) error {
	for _, blk := range blocks {
		if err := r.checkOutputBlock(blk, runner); err != nil {
			return err
		}
	}
	return nil
}

// gatherTraversals canonicalizes and filters out prefix traversals
func (r *TerraformOutputResourceRule) gatherTraversals(expr hcl.Expression) []hcl.Traversal {
	var collected []hcl.Traversal

	var walk func(hcl.Expression)
	walk = func(e hcl.Expression) {
		switch typed := e.(type) {
		case *hclsyntax.ScopeTraversalExpr:
			// direct reference, e.g. "aws_instance.example.id"
			collected = append(collected, typed.Traversal)

		case *hclsyntax.SplatExpr:
			// e.g. "aws_instance.multiple[*].id"
			// If the base expression is a ScopeTraversalExpr, we can combine its traversal with the splat operator and any trailing item.
			if base, ok := typed.Each.(*hclsyntax.ScopeTraversalExpr); ok {
				trav := append([]hcl.Traverser{}, base.Traversal...)
				// Append an explicit splat operator.
				trav = append(trav, hcl.TraverseSplat{})
				if typed.Item != nil {
					var itemExpr hcl.Expression = typed.Item
					// If Item is a ScopeTraversalExpr, append its traversal steps.
					if itemScope, ok := itemExpr.(*hclsyntax.ScopeTraversalExpr); ok {
						trav = append(trav, itemScope.Traversal...)
					} else {
						// Otherwise, try gathering traversals from Item and merge the first result.
						sub := r.gatherTraversals(itemExpr)
						if len(sub) > 0 {
							for _, t := range sub[0] {
								trav = append(trav, t)
							}
						}
					}
				}
				collected = append(collected, trav)
			} else {
				// Fallback if the Each part is not a ScopeTraversalExpr.
				walk(typed.Each)
				if typed.Item != nil {
					walk(typed.Item)
				}
			}

		case *hclsyntax.ConditionalExpr:
			// e.g. "condition ? trueVal : falseVal"
			walk(typed.Condition)
			walk(typed.TrueResult)
			walk(typed.FalseResult)

		case *hclsyntax.BinaryOpExpr:
			// e.g. "lhs == rhs"
			walk(typed.LHS)
			walk(typed.RHS)

		case *hclsyntax.UnaryOpExpr:
			// e.g. "-value"
			walk(typed.Val)

		case *hclsyntax.TemplateExpr:
			// e.g. "some string ${expression}"
			for _, part := range typed.Parts {
				walk(part)
			}

		case *hclsyntax.TupleConsExpr:
			// e.g. "[ expr1, expr2, ... ]"
			for _, elem := range typed.Exprs {
				walk(elem)
			}

		case *hclsyntax.ObjectConsExpr:
			// e.g. "{ key = expr }"
			for _, item := range typed.Items {
				walk(item.KeyExpr)
				walk(item.ValueExpr)
			}
		}
	}
	walk(expr)

	if len(collected) == 0 {
		return nil
	}

	var canonical []hcl.Traversal
	for _, trav := range collected {
		canonical = append(canonical, canonicalizeTraversal(trav))
	}
	return filterPrefixTraversals(canonical)
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

	// Use new helper
	traversals := r.gatherTraversals(expr)
	if len(traversals) == 0 {
		return nil
	}

	// If any of the traversals is a "bare" resource reference => report
	for _, trav := range traversals {
		// Only check if the root is "data" or a resource
		if !isResourceRootTraversal(trav) {
			continue
		}
		if r.isFullResourceReference(trav) {
			return runner.EmitIssue(
				r,
				"Output is referencing the entire resource or data, rather than a specific attribute. This can cause schema issues.",
				valAttr.Range(),
			)
		}
	}

	return nil
}

// isFullResourceReference checks if the traversal looks like a bare resource reference
// e.g., "aws_instance.foo", "aws_instance.foo[0]", "aws_instance.foo[*]",
// or "data.aws_instance.foo" with no sub-attributes after the name.
func (r *TerraformOutputResourceRule) isFullResourceReference(trav hcl.Traversal) bool {
	length := len(trav)
	if length < 2 {
		return false
	}
	if !isResourceRootTraversal(trav) {
		return false
	}

	// Special handling: if the root is "data", then a 3-step reference (data + 2 attributes) is "entire data resource".
	// e.g. data.aws_caller_identity.current => length=3 => entire
	if trav[0].(hcl.TraverseRoot).Name == "data" && length == 3 {
		return true
	}

	// If we only have two steps (e.g., "aws_instance.example"),
	// that is the entire resource (no sub-attributes).
	if length == 2 {
		return true
	}

	// If it ends with an attribute, then we assume it's referencing
	// some sub-attribute => partial => no error.
	if endsWithAttribute(trav) {
		return false
	}

	// Otherwise (e.g., "aws_instance.example[0]", "aws_instance.example[*]", etc.), it's entire.
	return true
}

// isResourceRootTraversal returns true if the traversal root is "data" or neither var/local/module nor empty (thus presumably a resource).
func isResourceRootTraversal(trav hcl.Traversal) bool {
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

// endsWithAttribute returns true if the final step is a TraverseAttr.
func endsWithAttribute(trav hcl.Traversal) bool {
	if len(trav) == 0 {
		return false
	}
	_, ok := trav[len(trav)-1].(hcl.TraverseAttr)
	return ok
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
			// If they're exactly the same, return true
			if aTyped.Name == bTyped.Name {
				return true
			}

			// If one side is "multiple" and the other is "multiple[*]", treat them as prefix
			if aTyped.Name+"[*]" == bTyped.Name || bTyped.Name+"[*]" == aTyped.Name {
				return true
			}

			// Handle "example" and "example.id" prefix
			if strings.HasPrefix(bTyped.Name, aTyped.Name+".") {
				return true
			}
			if strings.HasPrefix(aTyped.Name, bTyped.Name+".") {
				return true
			}
		}
	case hcl.TraverseIndex:
		if bTyped, ok := b.(hcl.TraverseIndex); ok {
			return aTyped.Key.RawEquals(bTyped.Key)
		}
	case hcl.TraverseSplat:
		_, ok := b.(hcl.TraverseSplat)
		return ok
	}
	return false
}

// canonicalizeTraversal breaks up any hcl.TraverseAttr that includes [*] or [index] or . into multiple steps.
func canonicalizeTraversal(trav hcl.Traversal) hcl.Traversal {
	var result hcl.Traversal

	for _, step := range trav {
		switch s := step.(type) {
		case hcl.TraverseRoot:
			result = append(result, hcl.TraverseRoot{Name: s.Name})
		case hcl.TraverseAttr:
			// If s.Name includes brackets or dots, split it
			subSteps := splitAttrName(s.Name)
			// subSteps might be ["multiple", "[*]", "id"]
			// We convert each piece into the right kind of traverser
			for _, sub := range subSteps {
				if sub == "[*]" {
					result = append(result, hcl.TraverseSplat{})
				} else if strings.HasPrefix(sub, "[") && strings.HasSuffix(sub, "]") {
					// e.g., "[0]" => TraverseIndex with key=0
					indexKey := strings.Trim(sub, "[]")
					result = append(result, makeIndexStep(indexKey))
				} else {
					// Plain attribute
					result = append(result, hcl.TraverseAttr{Name: sub})
				}
			}
		case hcl.TraverseIndex:
			result = append(result, hcl.TraverseIndex{Key: s.Key})
		case hcl.TraverseSplat:
			result = append(result, hcl.TraverseSplat{})
		default:
			result = append(result, step)
		}
	}

	return result
}

// splitAttrName splits an attribute like "multiple[*].id" into ["multiple", "[*]", "id"].
func splitAttrName(name string) []string {
	// Walk through name and cut on "." or bracket groups.
	// We want to preserve bracket groups as separate tokens: "[*]", "[0]", etc.

	var parts []string
	var buf strings.Builder

	for i := 0; i < len(name); i++ {
		c := name[i]

		switch c {
		case '.':
			// dot means we finished one part
			if buf.Len() > 0 {
				parts = append(parts, buf.String())
				buf.Reset()
			}
		case '[':
			// flush anything we had before '['
			if buf.Len() > 0 {
				parts = append(parts, buf.String())
				buf.Reset()
			}
			// now read until closing ']'
			j := i + 1
			for j < len(name) && name[j] != ']' {
				j++
			}
			if j < len(name) && name[j] == ']' {
				// j is at the ']'
				parts = append(parts, name[i:j+1]) // e.g., "[*]" or "[0]"
				i = j                              // skip ahead
			} else {
				// If no closing ']', treat as normal character
				buf.WriteByte(c)
			}
		default:
			buf.WriteByte(c)
		}
	}

	if buf.Len() > 0 {
		parts = append(parts, buf.String())
	}
	return parts
}

// makeIndexStep tries to parse "[0]" into a numeric index step, or "*" into a splat, etc.
func makeIndexStep(keyStr string) hcl.Traverser {
	if keyStr == "*" {
		return hcl.TraverseSplat{}
	}
	// try to parse int
	if idxVal, err := cty.ParseNumberVal(keyStr); err == nil {
		return hcl.TraverseIndex{Key: idxVal}
	}
	// fallback to string key
	return hcl.TraverseIndex{Key: cty.StringVal(keyStr)}
}
