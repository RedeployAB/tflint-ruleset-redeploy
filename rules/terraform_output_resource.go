package rules

import (
	"sort"
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
	return GetRuleDocLink(r.Name())
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
		if strings.EqualFold(blk.Type, "output") {
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

// checkOutputBlock inspects if the "value" attribute references an entire resource/data.
func (r *TerraformOutputResourceRule) checkOutputBlock(
	block *hclsyntax.Block,
	runner tflint.Runner,
) error {
	valAttr, ok := block.Body.Attributes["value"]
	if !ok {
		return nil
	}

	// Special handling for for expressions
	if forExpr, ok := valAttr.Expr.(*hclsyntax.ForExpr); ok {
		return r.checkForExpression(forExpr, runner, valAttr.Range())
	}

	// Gather traversals from the expression
	traversals := r.gatherTraversals(valAttr.Expr)
	if len(traversals) == 0 {
		return nil
	}

	// Check if any traversal is a bare resource reference
	for _, trav := range traversals {
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

// checkForExpression checks if a for expression outputs entire resources
func (r *TerraformOutputResourceRule) checkForExpression(
	forExpr *hclsyntax.ForExpr,
	runner tflint.Runner,
	exprRange hcl.Range,
) error {
	// Check if the value expression is just the loop variable
	// For example: [for x in resource : x] is bad
	// But: [for x in resource : x.attr] is OK
	if scopeTrav, ok := forExpr.ValExpr.(*hclsyntax.ScopeTraversalExpr); ok {
		// If it's just a single variable reference (the loop variable),
		// then we're outputting the entire resource
		if len(scopeTrav.Traversal) == 1 {
			if root, ok := scopeTrav.Traversal[0].(hcl.TraverseRoot); ok {
				// Check if this is the loop variable
				if forExpr.ValVar == root.Name ||
					(forExpr.KeyVar != "" && forExpr.KeyVar == root.Name) {
					// The value expression is just the loop variable
					// Check if the collection is a resource
					collTraversals := r.gatherTraversals(forExpr.CollExpr)
					for _, trav := range collTraversals {
						if isResourceRootTraversal(trav) && r.isFullResourceReference(trav) {
							return runner.EmitIssue(
								r,
								"Output is referencing the entire resource or data, rather than a specific attribute. This can cause schema issues.",
								exprRange,
							)
						}
					}
				}
			}
		}
	}

	// For all other cases (accessing attributes on loop variable, complex expressions, etc.), it's OK
	return nil
}

// gatherTraversals extracts and normalizes traversals from an expression
func (r *TerraformOutputResourceRule) gatherTraversals(expr hcl.Expression) []hcl.Traversal {
	// Check if expression requires manual walking
	switch expr.(type) {
	case *hclsyntax.SplatExpr, *hclsyntax.FunctionCallExpr, *hclsyntax.ForExpr:
		return r.collectAndCanonicalizeTraversals(expr)
	}

	// For other expressions, try HCL's built-in variable extraction first
	vars := expr.Variables()
	if len(vars) > 0 {
		if r.needsCanonicalization(vars) {
			return r.canonicalizeTraversals(vars)
		}
		return filterPrefixTraversals(vars)
	}

	// Fall back to manual walking for complex cases
	return r.collectAndCanonicalizeTraversals(expr)
}

// collectAndCanonicalizeTraversals walks expression and canonicalizes traversals
func (r *TerraformOutputResourceRule) collectAndCanonicalizeTraversals(expr hcl.Expression) []hcl.Traversal {
	var collected []hcl.Traversal
	r.walkExpression(expr, &collected)
	return r.canonicalizeTraversals(collected)
}

// needsCanonicalization checks if any traversal needs canonicalization
func (r *TerraformOutputResourceRule) needsCanonicalization(traversals []hcl.Traversal) bool {
	for _, trav := range traversals {
		for _, step := range trav {
			if attr, ok := step.(hcl.TraverseAttr); ok {
				if strings.Contains(attr.Name, "[") || strings.Contains(attr.Name, ".") {
					return true
				}
			}
		}
	}
	return false
}

// canonicalizeTraversals canonicalizes a slice of traversals
func (r *TerraformOutputResourceRule) canonicalizeTraversals(traversals []hcl.Traversal) []hcl.Traversal {
	var canonical []hcl.Traversal
	for _, trav := range traversals {
		canonical = append(canonical, canonicalizeTraversal(trav))
	}
	return filterPrefixTraversals(canonical)
}

// walkExpression recursively walks the expression to collect traversals
func (r *TerraformOutputResourceRule) walkExpression(e hcl.Expression, collected *[]hcl.Traversal) {
	switch typed := e.(type) {
	case *hclsyntax.ScopeTraversalExpr:
		*collected = append(*collected, typed.Traversal)

	case *hclsyntax.SplatExpr:
		r.walkSplatExpr(typed, collected)

	case *hclsyntax.ConditionalExpr:
		r.walkConditionalExpr(typed, collected)

	case *hclsyntax.TemplateExpr:
		r.walkTemplateExpr(typed, collected)

	case *hclsyntax.TupleConsExpr:
		r.walkTupleExpr(typed, collected)

	case *hclsyntax.ObjectConsExpr:
		r.walkObjectExpr(typed, collected)

	case *hclsyntax.BinaryOpExpr:
		r.walkBinaryOpExpr(typed, collected)

	case *hclsyntax.UnaryOpExpr:
		r.walkExpression(typed.Val, collected)

	case *hclsyntax.FunctionCallExpr:
		r.walkFunctionCallExpr(typed, collected)

	case *hclsyntax.ForExpr:
		r.walkForExpr(typed, collected)

	default:
		// Try to get variables from the expression
		vars := e.Variables()
		*collected = append(*collected, vars...)
	}
}

// walkConditionalExpr walks a conditional expression
func (r *TerraformOutputResourceRule) walkConditionalExpr(e *hclsyntax.ConditionalExpr, collected *[]hcl.Traversal) {
	r.walkExpression(e.Condition, collected)
	r.walkExpression(e.TrueResult, collected)
	r.walkExpression(e.FalseResult, collected)
}

// walkTemplateExpr walks a template expression
func (r *TerraformOutputResourceRule) walkTemplateExpr(e *hclsyntax.TemplateExpr, collected *[]hcl.Traversal) {
	for _, part := range e.Parts {
		r.walkExpression(part, collected)
	}
}

// walkTupleExpr walks a tuple expression
func (r *TerraformOutputResourceRule) walkTupleExpr(e *hclsyntax.TupleConsExpr, collected *[]hcl.Traversal) {
	for _, elem := range e.Exprs {
		r.walkExpression(elem, collected)
	}
}

// walkObjectExpr walks an object expression
func (r *TerraformOutputResourceRule) walkObjectExpr(e *hclsyntax.ObjectConsExpr, collected *[]hcl.Traversal) {
	for _, item := range e.Items {
		r.walkExpression(item.ValueExpr, collected)
	}
}

// walkBinaryOpExpr walks a binary operation expression
func (r *TerraformOutputResourceRule) walkBinaryOpExpr(e *hclsyntax.BinaryOpExpr, collected *[]hcl.Traversal) {
	r.walkExpression(e.LHS, collected)
	r.walkExpression(e.RHS, collected)
}

// walkFunctionCallExpr walks a function call expression
func (r *TerraformOutputResourceRule) walkFunctionCallExpr(e *hclsyntax.FunctionCallExpr, collected *[]hcl.Traversal) {
	for _, arg := range e.Args {
		r.walkExpression(arg, collected)
	}
}

// walkForExpr walks a for expression
func (r *TerraformOutputResourceRule) walkForExpr(e *hclsyntax.ForExpr, collected *[]hcl.Traversal) {
	// For expressions like: [for item in collection : item.attr]
	// We only want to walk the collection expression, not the value expression
	// because the value expression references the loop variable, not the resource
	r.walkExpression(e.CollExpr, collected)
	// Don't walk ValExpr or KeyExpr as they reference the loop variable

	// Walk the condition if present
	if e.CondExpr != nil {
		r.walkExpression(e.CondExpr, collected)
	}
}

// walkSplatExpr handles splat expressions specially
func (r *TerraformOutputResourceRule) walkSplatExpr(e *hclsyntax.SplatExpr, collected *[]hcl.Traversal) {
	// Get the base traversal from the source
	sourceVars := e.Source.Variables()

	// Check if there's an Each expression (e.g., the .id in resource[*].id)
	var eachSteps []hcl.Traverser
	if e.Each != nil {
		// Try to extract traversal steps from the Each expression
		if scopeTrav, ok := e.Each.(*hclsyntax.ScopeTraversalExpr); ok {
			// Skip the first step if it's a root (usually it's a relative traversal)
			for i, step := range scopeTrav.Traversal {
				if i == 0 {
					if _, isRoot := step.(hcl.TraverseRoot); isRoot {
						continue
					}
				}
				eachSteps = append(eachSteps, step)
			}
		} else if relTrav, ok := e.Each.(*hclsyntax.RelativeTraversalExpr); ok {
			eachSteps = relTrav.Traversal
		}
	}

	for _, trav := range sourceVars {
		// Build the complete traversal: source + splat + each
		fullTrav := make(hcl.Traversal, 0, len(trav)+1+len(eachSteps))
		fullTrav = append(fullTrav, trav...)
		fullTrav = append(fullTrav, hcl.TraverseSplat{})
		fullTrav = append(fullTrav, eachSteps...)
		*collected = append(*collected, fullTrav)
	}
}

// canonicalizeTraversal normalizes traversals with complex attribute names
func canonicalizeTraversal(trav hcl.Traversal) hcl.Traversal {
	var result hcl.Traversal

	for _, step := range trav {
		switch s := step.(type) {
		case hcl.TraverseRoot:
			result = append(result, s)
		case hcl.TraverseAttr:
			// Check if the attribute name contains brackets or dots
			if strings.Contains(s.Name, "[") || strings.Contains(s.Name, ".") {
				// Split the attribute name into parts
				parts := splitAttrName(s.Name)
				for _, part := range parts {
					result = append(result, parseTraversalPart(part)...)
				}
			} else {
				result = append(result, s)
			}
		default:
			result = append(result, step)
		}
	}

	return result
}

// splitAttrName splits an attribute name that may contain dots and brackets
func splitAttrName(name string) []string {
	var parts []string
	var current strings.Builder

	for i := 0; i < len(name); i++ {
		switch name[i] {
		case '.':
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
		case '[':
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
			// Find the closing bracket
			j := i + 1
			for j < len(name) && name[j] != ']' {
				j++
			}
			if j < len(name) {
				parts = append(parts, name[i:j+1])
				i = j
			} else {
				current.WriteByte(name[i])
			}
		default:
			current.WriteByte(name[i])
		}
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}
	return parts
}

// parseTraversalPart converts a string part into traversal steps
func parseTraversalPart(part string) []hcl.Traverser {
	if part == "[*]" {
		return []hcl.Traverser{hcl.TraverseSplat{}}
	}
	if strings.HasPrefix(part, "[") && strings.HasSuffix(part, "]") {
		key := strings.Trim(part, "[]\"'")
		if idx, err := cty.ParseNumberVal(key); err == nil {
			return []hcl.Traverser{hcl.TraverseIndex{Key: idx}}
		}
		return []hcl.Traverser{hcl.TraverseIndex{Key: cty.StringVal(key)}}
	}
	return []hcl.Traverser{hcl.TraverseAttr{Name: part}}
}

// filterPrefixTraversals removes traversals that are prefixes of other traversals.
// This implementation uses O(n log n) sorting instead of O(n²) comparison.
func filterPrefixTraversals(traversals []hcl.Traversal) []hcl.Traversal {
	if len(traversals) <= 1 {
		return traversals
	}

	// Sort traversals by their string representation for efficient prefix checking
	type traversalWithKey struct {
		trav hcl.Traversal
		key  string
	}

	sorted := make([]traversalWithKey, len(traversals))
	for i, t := range traversals {
		sorted[i] = traversalWithKey{trav: t, key: traversalKey(t)}
	}

	// Sort by key - this groups potential prefixes together
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].key < sorted[j].key
	})

	var result []hcl.Traversal
	for i := 0; i < len(sorted); i++ {
		// Check if the next traversal has this one as a prefix
		isAPrefix := false
		if i+1 < len(sorted) && strings.HasPrefix(sorted[i+1].key, sorted[i].key+".") {
			// Verify it's actually a prefix (not just string prefix)
			if isPrefix(sorted[i].trav, sorted[i+1].trav) {
				isAPrefix = true
			}
		}
		if !isAPrefix {
			result = append(result, sorted[i].trav)
		}
	}
	return result
}

// traversalKey creates a string key for a traversal for sorting purposes.
// Uses strings.Builder for better performance (2x faster, fewer allocations).
func traversalKey(trav hcl.Traversal) string {
	var sb strings.Builder
	for i, step := range trav {
		if i > 0 {
			sb.WriteByte('.')
		}
		switch s := step.(type) {
		case hcl.TraverseRoot:
			sb.WriteString(s.Name)
		case hcl.TraverseAttr:
			sb.WriteString(s.Name)
		case hcl.TraverseIndex:
			// Use a placeholder for index to maintain grouping
			sb.WriteString("[idx]")
		case hcl.TraverseSplat:
			sb.WriteString("[*]")
		}
	}
	return sb.String()
}

// isPrefix checks if t1 is a prefix of t2
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

// stepEqual compares two traversal steps
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
		if bTyped, ok := b.(hcl.TraverseIndex); ok {
			return aTyped.Key.RawEquals(bTyped.Key)
		}
	case hcl.TraverseSplat:
		_, ok := b.(hcl.TraverseSplat)
		return ok
	}
	return false
}

// isFullResourceReference checks if the traversal is a complete resource reference
func (r *TerraformOutputResourceRule) isFullResourceReference(trav hcl.Traversal) bool {
	length := len(trav)
	if length < 2 {
		return false
	}

	root, ok := trav[0].(hcl.TraverseRoot)
	if !ok {
		return false
	}

	// Skip variables, locals, and modules
	switch root.Name {
	case TypeVar, TypeLocal, TypeModule:
		return false
	}

	// For data sources: data.TYPE.NAME is a full reference (3 parts)
	if root.Name == TypeData && length == 3 {
		return true
	}

	// For resources: TYPE.NAME is a full reference (2 parts)
	if root.Name != TypeData && length == 2 {
		return true
	}

	// Check if the traversal ends with an attribute access
	// If it does, it's accessing a specific attribute, not the entire resource
	if _, ok := trav[length-1].(hcl.TraverseAttr); ok && length > minResourceLength(root.Name) {
		return false
	}

	// Otherwise, it's a full resource reference (including indexed/splat access)
	return true
}

// isResourceRootTraversal checks if the traversal starts with a resource or data reference
func isResourceRootTraversal(trav hcl.Traversal) bool {
	if len(trav) == 0 {
		return false
	}
	root, ok := trav[0].(hcl.TraverseRoot)
	if !ok {
		return false
	}
	switch root.Name {
	case TypeVar, TypeLocal, TypeModule:
		return false
	}
	return true
}

// minResourceLength returns the minimum length for a complete resource reference
func minResourceLength(rootName string) int {
	if rootName == TypeData {
		return 3 // data.TYPE.NAME
	}
	return 2 // TYPE.NAME
}
