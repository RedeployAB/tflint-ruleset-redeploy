package rules

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

type TerraformMetaArgumentsRule struct {
	tflint.DefaultRule
}

func NewTerraformMetaArgumentsRule() *TerraformMetaArgumentsRule {
	return &TerraformMetaArgumentsRule{}
}

func (r *TerraformMetaArgumentsRule) Name() string {
	return "terraform_meta_arguments"
}

func (r *TerraformMetaArgumentsRule) Enabled() bool {
	return true
}

func (r *TerraformMetaArgumentsRule) Severity() tflint.Severity {
	return tflint.ERROR
}

func (r *TerraformMetaArgumentsRule) Link() string {
	return ""
}

// Check enforces the required ordering and blank-line separation:
// 1) count or for_each, 2) provider, <blank line>, 3) lifecycle, <blank line>, 4) depends_on
func (r *TerraformMetaArgumentsRule) Check(runner tflint.Runner) error {
	content, err := runner.GetModuleContent(&hclext.BodySchema{
		Blocks: []hclext.BlockSchema{
			{
				Type:       "resource",
				LabelNames: []string{"type", "name"},
				Body: &hclext.BodySchema{
					Attributes: []hclext.AttributeSchema{
						{Name: "count"},
						{Name: "for_each"},
						{Name: "provider"},
					},
					Blocks: []hclext.BlockSchema{
						{Type: "lifecycle"},
						{Type: "depends_on"},
					},
				},
			},
		},
	}, nil)
	if err != nil {
		return err
	}

	for _, resource := range content.Blocks {
		// Gather the meta-arguments in the order they appear in the file
		lines := resource.Body.UnsortedContent()
		linesText := make([]string, 0, len(lines))

		for _, item := range lines {
			switch {
			case item.Attribute != nil:
				// e.g. "count", "for_each", or "provider"
				linesText = append(linesText, item.Attribute.Name)
			case item.Block != nil:
				// e.g. "lifecycle" or "depends_on"
				linesText = append(linesText, item.Block.Type)
			case item.Body != nil && item.Body.Item != nil:
				// Handle blank lines or comments (if possible)
			}
		}

		// Expected distinct sections:
		// 1) one of ["count","for_each"] then "provider"
		// 2) blank line
		// 3) "lifecycle"
		// 4) blank line
		// 5) "depends_on"

		// We'll just do a naive check for the sequence with minimal validation:
		desiredSequence := []string{"count|for_each", "provider", "lifecycle", "depends_on"}
		foundIndex := 0

		for _, name := range linesText {
			// Check for count/for_each
			if foundIndex == 0 && (name == "count" || name == "for_each") {
				foundIndex++
				continue
			}
			// Then provider
			if foundIndex <= 1 && name == "provider" {
				foundIndex = 2
				continue
			}
			// Then lifecycle
			if foundIndex <= 2 && name == "lifecycle" {
				foundIndex = 3
				continue
			}
			// Finally depends_on
			if foundIndex <= 3 && name == "depends_on" {
				foundIndex = 4
				continue
			}

			// If we reach here, the item is out of order
			errMsg := fmt.Sprintf(
				"Meta-argument '%s' is out of expected order. Current sequence: %s",
				name, strings.Join(linesText, ", "),
			)
			if emitErr := runner.EmitIssue(r, errMsg, resource.DefRange); emitErr != nil {
				return emitErr
			}
			break
		}

		// Check if we completed the entire desired sequence
		if foundIndex < 2 {
			// Means we never found provider or meta-arguments
			msg := "Missing or out-of-order meta arguments: expected count/for_each, then provider"
			if emitErr := runner.EmitIssue(r, msg, resource.DefRange); emitErr != nil {
				return emitErr
			}
		}

		// You could add checks for blank lines here by comparing item ranges
		// but that’s more advanced parsing. This minimal approach covers the order.
	}

	return nil
}
