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
	// 1) Build a dictionary mapping metaArg -> index in metaArgs
	metaArgIndices := make(map[string]int)
	for i, arg := range metaArgs {
		// If there's a collision (like "count" and "for_each" both found),
		// keep the earlier index if we want whichever is first
		if _, exists := metaArgIndices[arg]; !exists {
			metaArgIndices[arg] = i
		}
	}

	// 2) We'll keep track of the highest index found so far
	lastIndex := -1

	// Helper to find the actual index of a meta-argument if it exists
	// returns -1 if not found
	getIndex := func(arg string) int {
		if i, ok := metaArgIndices[arg]; ok {
			return i
		}
		return -1
	}

	for i := 0; i < len(desiredSequence); i++ {
		want := desiredSequence[i]

		// If we are dealing with "count|for_each"
		if want == ArgCount+"|"+ArgForEach {
			countIdx := getIndex(ArgCount)
			forEachIdx := getIndex(ArgForEach)

			// If both are absent, skip
			if countIdx < 0 && forEachIdx < 0 {
				continue
			}
			// If both are present, pick whichever is earlier
			foundIdx := -1
			if countIdx >= 0 && forEachIdx >= 0 {
				if countIdx < forEachIdx {
					foundIdx = countIdx
				} else {
					foundIdx = forEachIdx
				}
			} else if countIdx >= 0 {
				foundIdx = countIdx
			} else {
				foundIdx = forEachIdx
			}

			if foundIdx < lastIndex {
				// out-of-order
				argName := ArgCount
				if foundIdx == forEachIdx {
					argName = ArgForEach
				}

				return runner.EmitIssue(r,
					fmt.Sprintf(
						"Out-of-order meta argument '%s' in %s '%s'. Expected sequence: %s",
						argName, block.Type, blockLabels,
						strings.Join(desiredSequence, " -> "),
					),
					metaArgRange(block, argName),
				)
			}

			lastIndex = foundIdx
		} else {
			// Normal argument
			idx := getIndex(want)
			if idx < 0 {
				// not present -> skip
				continue
			}
			// If present, must appear after lastIndex
			if idx < lastIndex {
				// out-of-order
				return runner.EmitIssue(r,
					fmt.Sprintf(
						"Out-of-order meta argument '%s' in %s '%s'. Expected sequence: %s",
						want, block.Type, blockLabels,
						strings.Join(desiredSequence, " -> "),
					),
					metaArgRange(block, want),
				)
			}
			lastIndex = idx
		}
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
