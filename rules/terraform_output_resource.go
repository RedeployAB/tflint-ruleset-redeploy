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

	// We parse the expression and see if it’s a single traversal referencing a resource or data
	// e.g. "aws_instance.foo" or "data.aws_instance.foo" with no sub-attributes
	traversals := expr.Variables()
	if len(traversals) == 0 {
		return nil
	}

	// If *any* of the traversals is a "bare" reference to resource/data => report
	for _, trav := range traversals {
		// trav is a hcl.Traversal
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
// E.g. "aws_instance.foo" or "data.aws_instance.foo" with no sub-attributes after the name.
func (r *TerraformOutputResourceRule) isEntireResourceReference(trav hcl.Traversal) bool {
	// We only consider it a “bare” entire resource if:
	//   1) trav length == 2: e.g. [Root("aws_instance"), Attr("my_example")]
	//   2) trav length == 3 and trav[0] == "data": e.g. [Root("data"), Attr("aws_iam_user"), Attr("blah")]
	//   3) No indexing or extra sub-attributes
	//
	// If any step is TraverseIndex(...) or if we have more than these minimal steps, it's partial or an attribute => ignore

	switch len(trav) {
	case 2:
		// e.g. resource.resource_name
		// Ensure no indexing
		if hasIndex(trav) {
			return false
		}
		return true
	case 3:
		// e.g. data.resource_name.example
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
