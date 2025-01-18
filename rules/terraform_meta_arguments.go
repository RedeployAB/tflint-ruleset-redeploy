package rules

import (
	"fmt"
	"strings"

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
// For resources:
//   1) count or for_each
//   2) provider
//   <blank line>
//   3) lifecycle
//   <blank line>
//   4) depends_on
//
// For modules:
//   1) count or for_each
//   <blank line>
//   2) depends_on
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
			{
				Type:       "module",
				LabelNames: []string{"name"},
				Body: &hclext.BodySchema{
					Attributes: []hclext.AttributeSchema{
						{Name: "count"},
						{Name: "for_each"},
					},
					Blocks: []hclext.BlockSchema{
						{Type: "depends_on"},
					},
				},
			},
		},
	}, nil)
	if err != nil {
		return err
	}

	for _, block := range content.Blocks {
		// Gather the meta-arguments; order is not preserved
		attributes, err := block.Body.Attributes()
		if err != nil {
			return err
		}
		blocks, err := block.Body.Blocks()
		if err != nil {
			return err
		}

		// Populate linesText from the attribute and block names
		linesText := make([]string, 0, len(attributes)+len(blocks))

		for _, attr := range attributes {
			linesText = append(linesText, attr.Name)
		}
		for _, b := range blocks {
			linesText = append(linesText, b.Type)
		}

		// Define the desired sequences based on block type
		var desiredSequence []string
		if block.Type == "resource" {
			desiredSequence = []string{"count|for_each", "provider", "lifecycle", "depends_on"}
		} else if block.Type == "module" {
			desiredSequence = []string{"count|for_each", "depends_on"}
		} else {
			continue // Skip other block types
		}

		foundIndex := 0
		for _, name := range linesText {
			// Check for count/for_each
			if foundIndex == 0 && (name == "count" || name == "for_each") {
				foundIndex++
				continue
			}
			// For resources: then provider
			if block.Type == "resource" && foundIndex == 1 && name == "provider" {
				foundIndex++
				continue
			}
			// For resources: then lifecycle
			if block.Type == "resource" && foundIndex == 2 && name == "lifecycle" {
				foundIndex++
				continue
			}
			// For modules and resources: depends_on
			if (block.Type == "module" && foundIndex >= 1 && name == "depends_on") ||
				(block.Type == "resource" && foundIndex >= 1 && name == "depends_on") {
				foundIndex = len(desiredSequence)
				continue
			}

			// If we reach here, the item is out of order
			errMsg := fmt.Sprintf(
				"Meta-argument '%s' is out of expected order in %s '%s'. Current sequence: %s",
				name, block.Type, strings.Join(block.Labels, " "), strings.Join(linesText, ", "),
			)
			if emitErr := runner.EmitIssue(r, errMsg, block.DefRange); emitErr != nil {
				return emitErr
			}
			break
		}

		// Check if we completed the entire desired sequence
		if foundIndex < len(desiredSequence) {
			msg := fmt.Sprintf(
				"Missing or out-of-order meta arguments in %s '%s'. Expected sequence: %s",
				block.Type, strings.Join(block.Labels, " "), strings.Join(desiredSequence, " -> "),
			)
			if emitErr := runner.EmitIssue(r, msg, block.DefRange); emitErr != nil {
				return emitErr
			}
		}
	}

	return nil
}
