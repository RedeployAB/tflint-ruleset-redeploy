package rules

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
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

// Check enforces the required ordering of meta-arguments in resource and module blocks.
// For resources:
//   1) count or for_each
//   2) provider
//   3) lifecycle
//   4) depends_on
//
// For modules:
//   1) count or for_each
//   2) depends_on
func (r *TerraformMetaArgumentsRule) Check(runner tflint.Runner) error {
	files, err := runner.GetFiles()
	if err != nil {
		return err
	}

	for filename, hclFile := range files {
		// Skip if no content
		if hclFile == nil || hclFile.Bytes == nil {
			continue
		}

		// Use hclsyntax.ParseConfig() to parse the raw bytes in lexical order
		syntaxFile, diags := hclsyntax.ParseConfig(hclFile.Bytes, filename, hcl.InitialPos)
		if diags.HasErrors() {
			// Either skip or emit an issue if parse fails
			continue
		}

		// Now syntaxFile.Body is an *hclsyntax.Body that preserves ordering
		fileBody := syntaxFile.Body

		syntaxBody, ok := fileBody.(*hclsyntax.Body)
		if !ok {
			// Cannot parse body; skip this file
			continue
		}

		if err := r.processBody(syntaxBody, filename, runner); err != nil {
			return err
		}
	}

	// If everything succeeded
	return nil
}

func (r *TerraformMetaArgumentsRule) processBody(body *hclsyntax.Body, filename string, runner tflint.Runner) error {
	// Collect attributes and blocks with their positions
	type contentItem struct {
		Name     string
		Type     string // "attribute" or "block"
		Attr     *hclsyntax.Attribute
		Block    *hclsyntax.Block
		SrcRange hcl.Range
	}

	var contentItems []contentItem

	for _, attr := range body.Attributes {
		contentItems = append(contentItems, contentItem{
			Name:     attr.Name,
			Type:     "attribute",
			Attr:     attr,
			SrcRange: attr.Range(), // Use attr.Range() for the correct range
		})
	}

	for _, block := range body.Blocks {
		contentItems = append(contentItems, contentItem{
			Name:     block.Type,
			Type:     "block",
			Block:    block,
			SrcRange: block.DefRange(), // Use block.DefRange()
		})
	}

	// Sort contentItems by their position in the file to preserve ordering
	sort.Slice(contentItems, func(i, j int) bool {
		return contentItems[i].SrcRange.Start.Byte < contentItems[j].SrcRange.Start.Byte
	})

	// Iterate over contentItems in order
	for _, item := range contentItems {
		if item.Type == "block" {
			// Check resource and module blocks
			if item.Block.Type == "resource" || item.Block.Type == "module" {
				if err := r.checkBlock(item.Block, runner); err != nil {
					return err
				}
			} else {
				// Recursively process other blocks
				if err := r.processBody(item.Block.Body, filename, runner); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (r *TerraformMetaArgumentsRule) checkBlock(block *hclsyntax.Block, runner tflint.Runner) error {
	// Define the expected meta-argument sequence
	var desiredSequence []string
	if block.Type == "resource" {
		desiredSequence = []string{"count|for_each", "provider", "lifecycle", "depends_on"}
	} else if block.Type == "module" {
		desiredSequence = []string{"count|for_each", "depends_on"}
	} else {
		// Not a resource or module block
		return nil
	}

	// Get the block labels for reporting
	blockLabels := strings.Join(block.Labels, " ")

	// Collect block content items in order
	type contentItem struct {
		Name     string
		Type     string // "attribute" or "block"
		SrcRange hcl.Range
	}

	var contentItems []contentItem

	for _, attr := range block.Body.Attributes {
		contentItems = append(contentItems, contentItem{
			Name:     attr.Name,
			Type:     "attribute",
			SrcRange: attr.Range(), // Use attr.Range() for correct range
		})
	}

	for _, childBlock := range block.Body.Blocks {
		contentItems = append(contentItems, contentItem{
			Name:     childBlock.Type,
			Type:     "block",
			SrcRange: childBlock.DefRange(), // Use childBlock.DefRange()
		})
	}

	// Sort the content items to preserve ordering
	sort.Slice(contentItems, func(i, j int) bool {
		return contentItems[i].SrcRange.Start.Byte < contentItems[j].SrcRange.Start.Byte
	})

	// Collect meta-arguments in the order they appear
	metaArgs := []string{}
	for _, item := range contentItems {
		if item.Type == "attribute" {
			if item.Name == "count" || item.Name == "for_each" || item.Name == "provider" || item.Name == "depends_on" {
				metaArgs = append(metaArgs, item.Name)
			}
		} else if item.Type == "block" {
			if item.Name == "lifecycle" {
				metaArgs = append(metaArgs, item.Name)
			}
		}
	}

	// Allow blocks with zero meta-arguments
	if len(metaArgs) == 0 {
		// If there are no meta-arguments, allow it with no issues
		return nil
	}

	// Verify the order of meta-arguments
	expectedIndex := 0
	actualIndex := 0

	for actualIndex < len(metaArgs) && expectedIndex < len(desiredSequence) {
		expected := desiredSequence[expectedIndex]
		actual := metaArgs[actualIndex]

		if expected == "count|for_each" {
			if actual == "count" || actual == "for_each" {
				expectedIndex++
				actualIndex++
				continue
			}
		} else if expected == actual {
			expectedIndex++
			actualIndex++
			continue
		}

		// Out-of-order meta-argument found
		msg := fmt.Sprintf(
			"Missing or out-of-order meta arguments in %s '%s'. Expected sequence: %s",
			block.Type, blockLabels, strings.Join(desiredSequence, " -> "),
		)
		return runner.EmitIssue(r, msg, block.DefRange())
	}

	// Check for any remaining expected meta-arguments
	if expectedIndex < len(desiredSequence) {
		msg := fmt.Sprintf(
			"Missing meta arguments in %s '%s'. Expected sequence: %s",
			block.Type, blockLabels, strings.Join(desiredSequence, " -> "),
		)
		return runner.EmitIssue(r, msg, block.DefRange())
	}

	return nil
}
