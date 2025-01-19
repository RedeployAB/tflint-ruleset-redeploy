package rules

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

const (
	argDependsOn = "depends_on"
	argLifecycle = "lifecycle"
	argCount     = "count"
	argForEach   = "for_each"
	argProvider  = "provider"
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
	type contentItem struct {
		Name     string
		Type     string
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
			SrcRange: attr.Range(),
		})
	}

	for _, block := range body.Blocks {
		contentItems = append(contentItems, contentItem{
			Name:     block.Type,
			Type:     "block",
			Block:    block,
			SrcRange: block.DefRange(),
		})
	}

	sort.Slice(contentItems, func(i, j int) bool {
		return contentItems[i].SrcRange.Start.Byte < contentItems[j].SrcRange.Start.Byte
	})

	for _, item := range contentItems {
		if item.Type == "block" {
			if item.Block.Type == "resource" || item.Block.Type == "module" {
				if err := r.checkBlock(item.Block, runner); err != nil {
					return err
				}
			} else {
				if err := r.processBody(item.Block.Body, filename, runner); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (r *TerraformArgumentOrderRule) checkBlock(block *hclsyntax.Block, runner tflint.Runner) error {
	var desiredSequence []string
	if block.Type == "resource" {
		desiredSequence = []string{argCount + "|" + argForEach, argProvider, argLifecycle, argDependsOn}
	} else if block.Type == "module" {
		desiredSequence = []string{argCount + "|" + argForEach, argDependsOn}
	} else {
		return nil
	}

	blockLabels := strings.Join(block.Labels, " ")

	type contentItem struct {
		Name     string
		Type     string
		SrcRange hcl.Range
	}

	var contentItems []contentItem

	for _, attr := range block.Body.Attributes {
		contentItems = append(contentItems, contentItem{
			Name:     attr.Name,
			Type:     "attribute",
			SrcRange: attr.Range(),
		})
	}

	for _, childBlock := range block.Body.Blocks {
		contentItems = append(contentItems, contentItem{
			Name:     childBlock.Type,
			Type:     "block",
			SrcRange: childBlock.DefRange(),
		})
	}

	sort.Slice(contentItems, func(i, j int) bool {
		return contentItems[i].SrcRange.Start.Byte < contentItems[j].SrcRange.Start.Byte
	})

	metaArgs := []string{}
	for _, item := range contentItems {
		if item.Type == "attribute" {
			if item.Name == argCount || item.Name == argForEach || item.Name == argProvider || item.Name == argDependsOn {
				metaArgs = append(metaArgs, item.Name)
			}
		} else if item.Type == "block" {
			if item.Name == argLifecycle {
				metaArgs = append(metaArgs, item.Name)
			}
		}
	}

	if len(metaArgs) == 0 {
		return nil
	}

	expectedIndex := 0
	actualIndex := 0

	for actualIndex < len(metaArgs) {
		if expectedIndex >= len(desiredSequence) {
			return runner.EmitIssue(
				r,
				fmt.Sprintf("Out-of-order meta argument '%s' in %s '%s'. Expected sequence: %s", metaArgs[actualIndex], block.Type, blockLabels, strings.Join(desiredSequence, " -> ")),
				metaArgRange(block, metaArgs[actualIndex]),
			)
		}

		expected := desiredSequence[expectedIndex]
		actual := metaArgs[actualIndex]

		if (expected == argCount+"|"+argForEach && (actual == argCount || actual == argForEach)) || (expected == actual) {
			expectedIndex++
			actualIndex++
			continue
		}

		foundMatch := false
		for expectedIndex < len(desiredSequence) {
			expected := desiredSequence[expectedIndex]
			if (expected == argCount+"|"+argForEach && (actual == argCount || actual == argForEach)) || (expected == actual) {
				foundMatch = true
				break
			}
			expectedIndex++
		}

		if foundMatch {
			expectedIndex++
			actualIndex++
			continue
		}

		return runner.EmitIssue(
			r,
			fmt.Sprintf("Out-of-order meta argument '%s' in %s '%s'. Expected sequence: %s", actual, block.Type, blockLabels, strings.Join(desiredSequence, " -> ")),
			metaArgRange(block, actual),
		)
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
