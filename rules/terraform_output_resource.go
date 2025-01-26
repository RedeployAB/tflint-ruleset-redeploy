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
	// Usually:
	// resource:  [ TraverseRoot("aws_instance"), TraverseAttr("foo") ]
	// data:      [ TraverseRoot("data"), TraverseAttr("aws_iam_user"), TraverseAttr("blah") ]
	// We want to ensure there's *no* extra attribute after the resource name (like .id, .arn, etc.)

	if len(trav) < 2 {
		return false
	}

	// For a normal resource "aws_instance.foo", length == 2
	// For data "data.aws_iam_user.example", length == 3
	// We consider the last step in the traversal to see if that’s the final resource name,
	// with no further .something steps

	// Case 1: normal resource => exactly 2 segments => entire resource
	if len(trav) == 2 {
		return true
	}
	// Case 2: data resource => e.g. "data", "aws_iam_user", "example" => length == 3 => entire resource
	if len(trav) == 3 {
		first, okA := trav[0].(hcl.TraverseRoot)
		_, okB := trav[1].(hcl.TraverseAttr)
		if okA && okB && first.Name == "data" {
			return true
		}
	}

	return false
}
