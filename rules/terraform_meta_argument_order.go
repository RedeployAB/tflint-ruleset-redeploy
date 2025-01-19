package rules

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

type TerraformArgumentOrderRule struct {
	tflint.DefaultRule
}

func NewTerraformMetaArgumentOrderRule() *TerraformArgumentOrderRule {
	return &TerraformArgumentOrderRule{}
}

func (r *TerraformArgumentOrderRule) Name() string {
	return "terraform_meta_argument_order"
}

func (r *TerraformArgumentOrderRule) Enabled() bool {
	return true
}

func (r *TerraformArgumentOrderRule) Severity() tflint.Severity {
	return tflint.ERROR
}

func (r *TerraformArgumentOrderRule) Link() string {
	return ""
}

func (r *TerraformArgumentOrderRule) Check(runner tflint.Runner) error {
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
			if err := r.processBody(body, filename, runner); err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *TerraformArgumentOrderRule) processBody(body *hclsyntax.Body, filename string, runner tflint.Runner) error {
	for _, block := range body.Blocks {
		if block.Type == TypeResource || block.Type == TypeModule {
			if err := r.checkBlock(block, runner); err != nil {
				return err
			}
		} else {
			if err := r.processBody(block.Body, filename, runner); err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *TerraformArgumentOrderRule) checkBlock(block *hclsyntax.Block, runner tflint.Runner) error {
	desiredSequence := r.getDesiredSequence(block.Type)
	if desiredSequence == nil {
		return nil
	}
	blockLabels := strings.Join(block.Labels, " ")
	metaArgs := r.collectMetaArguments(block)
	if len(metaArgs) == 0 {
		return nil
	}
	return r.checkMetaArgSequence(metaArgs, desiredSequence, block, blockLabels, runner)
}

func (r *TerraformArgumentOrderRule) getDesiredSequence(blockType string) []string {
	switch blockType {
	case TypeResource:
		return []string{ArgCount + "|" + ArgForEach, ArgProvider, ArgLifecycle, ArgDependsOn}
	case TypeModule:
		return []string{ArgCount + "|" + ArgForEach, ArgDependsOn}
	default:
		return nil
	}
}

func (r *TerraformArgumentOrderRule) collectMetaArguments(block *hclsyntax.Block) []string {
	type contentItem struct {
		Name string
		Type string
	}
	var items []contentItem
	for _, attr := range block.Body.Attributes {
		items = append(items, contentItem{Name: attr.Name, Type: TypeAttr})
	}
	for _, childBlock := range block.Body.Blocks {
		items = append(items, contentItem{Name: childBlock.Type, Type: TypeBlock})
	}
	var metaArgs []string
	for _, it := range items {
		if (it.Type == TypeAttr && (it.Name == ArgCount || it.Name == ArgForEach || it.Name == ArgProvider || it.Name == ArgDependsOn)) ||
			(it.Type == TypeBlock && it.Name == ArgLifecycle) {
			metaArgs = append(metaArgs, it.Name)
		}
	}
	return metaArgs
}

func (r *TerraformArgumentOrderRule) checkMetaArgSequence(
	metaArgs, desiredSequence []string,
	block *hclsyntax.Block,
	blockLabels string,
	runner tflint.Runner,
) error {
	expectedIndex := 0
	actualIndex := 0

	// Helper to see if metaArgs includes an item
	metaArgIn := func(arg string) bool {
		for _, m := range metaArgs {
			if m == arg {
				return true
			}
		}
		return false
	}

	// Simple helper for matching "count|for_each" or exact match
	matchArg := func(actual, expected string) bool {
		if expected == ArgCount+"|"+ArgForEach {
			return (actual == ArgCount || actual == ArgForEach)
		}
		return (actual == expected)
	}

	for actualIndex < len(metaArgs) {
		// Skip any unneeded items in desiredSequence that do not appear in metaArgs at all
		for expectedIndex < len(desiredSequence) && !metaArgIn(desiredSequence[expectedIndex]) {
			expectedIndex++
		}

		// If we've run out of expected slots, it's out-of-order
		if expectedIndex >= len(desiredSequence) {
			return runner.EmitIssue(
				r,
				fmt.Sprintf(
					"Out-of-order meta argument '%s' in %s '%s'. Expected sequence: %s",
					metaArgs[actualIndex], block.Type, blockLabels, strings.Join(desiredSequence, " -> "),
				),
				metaArgRange(block, metaArgs[actualIndex]),
			)
		}

		actual := metaArgs[actualIndex]
		expected := desiredSequence[expectedIndex]

		// If this actual doesn't match the next expected exactly, it's out-of-order
		if !matchArg(actual, expected) {
			return runner.EmitIssue(
				r,
				fmt.Sprintf(
					"Out-of-order meta argument '%s' in %s '%s'. Expected sequence: %s",
					actual, block.Type, blockLabels, strings.Join(desiredSequence, " -> "),
				),
				metaArgRange(block, actual),
			)
		}

		// Move to next expected + actual
		expectedIndex++
		actualIndex++
	}
	return nil
}

func metaArgRange(block *hclsyntax.Block, argName string) hcl.Range {
	for _, attr := range block.Body.Attributes {
		if attr.Name == argName {
			return attr.Range()
		}
	}
	for _, childBlock := range block.Body.Blocks {
		if childBlock.Type == argName {
			return childBlock.DefRange()
		}
	}
	return block.DefRange()
}
