package rules

import (
	"fmt"
	"sort"
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
	metaArgs := r.collectMetaArgumentsInLexOrder(block)
	if len(metaArgs) == 0 {
		return nil
	}
	return r.checkMetaArgSequence(metaArgs, desiredSequence, block, blockLabels, runner)
}

func (r *TerraformArgumentOrderRule) getDesiredSequence(blockType string) []string {
	switch blockType {
	case TypeResource:
		return []string{ArgProvider, ArgCount + "|" + ArgForEach, ArgLifecycle, ArgDependsOn}
	case TypeModule:
		return []string{ArgCount + "|" + ArgForEach, ArgDependsOn}
	default:
		return nil
	}
}

// collectMetaArgumentsInLexOrder collects meta-arguments (count, for_each, provider, depends_on, lifecycle)
// in the exact lexical order they appear in the block. This ensures we don’t incorrectly treat
// “depends_on” or others as out-of-order if they actually appear properly in the file.
func (r *TerraformArgumentOrderRule) collectMetaArgumentsInLexOrder(block *hclsyntax.Block) []string {
	type item struct {
		Name     string
		Type     string // "attr" or "block"
		StartIdx int    // position of the line/byte
	}

	var items []item

	for _, attr := range block.Body.Attributes {
		items = append(items, item{
			Name:     attr.Name,
			Type:     TypeAttr,
			StartIdx: attr.Range().Start.Byte,
		})
	}
	for _, childBlock := range block.Body.Blocks {
		items = append(items, item{
			Name:     childBlock.Type,
			Type:     TypeBlock,
			StartIdx: childBlock.DefRange().Start.Byte,
		})
	}

	// sort them by lexical StartIdx so we see them in the actual file order
	sort.Slice(items, func(i, j int) bool {
		return items[i].StartIdx < items[j].StartIdx
	})

	var metaArgs []string
	for _, i := range items {
		// only record recognized meta arguments
		if i.Type == TypeAttr {
			switch i.Name {
			case ArgCount, ArgForEach, ArgProvider, ArgDependsOn:
				metaArgs = append(metaArgs, i.Name)
			}
		} else if i.Type == TypeBlock {
			// for blocks, only "lifecycle" is a meta-argument
			if i.Name == ArgLifecycle {
				metaArgs = append(metaArgs, i.Name)
			}
		}
	}
	return metaArgs
}

// checkCountOrForEach tries to see whether "count|for_each" is out of order,
// comparing foundIdx to lastIndex. Returns updated lastIndex or an error if out-of-order.
func (r *TerraformArgumentOrderRule) checkCountOrForEach(
	foundIdx, forEachIdx, lastIndex int,
	block *hclsyntax.Block, blockLabels string,
	desiredSequence []string, runner tflint.Runner,
) (int, error) {
	if foundIdx < lastIndex {
		// out-of-order
		argName := ArgCount
		if foundIdx == forEachIdx {
			argName = ArgForEach
		}
		return lastIndex, runner.EmitIssue(r,
			fmt.Sprintf(
				"Out-of-order meta argument '%s' in %s '%s'. Expected sequence: %s",
				argName, block.Type, blockLabels,
				strings.Join(desiredSequence, " -> "),
			),
			metaArgRange(block, argName),
		)
	}
	return foundIdx, nil
}

// checkSingleArg ensures that idx >= 0 appears after lastIndex, or else it’s out-of-order.
func (r *TerraformArgumentOrderRule) checkSingleArg(
	want string, idx, lastIndex int,
	block *hclsyntax.Block, blockLabels string,
	desiredSequence []string, runner tflint.Runner,
) (int, error) {
	if idx < lastIndex {
		return lastIndex, runner.EmitIssue(r,
			fmt.Sprintf(
				"Out-of-order meta argument '%s' in %s '%s'. Expected sequence: %s",
				want, block.Type, blockLabels,
				strings.Join(desiredSequence, " -> "),
			),
			metaArgRange(block, want),
		)
	}
	return idx, nil
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
		if _, already := metaArgIndices[arg]; !already {
			metaArgIndices[arg] = i // store first occurrence only
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
			var foundIdx int
			switch {
			case countIdx >= 0 && forEachIdx >= 0:
				if countIdx < forEachIdx {
					foundIdx = countIdx
				} else {
					foundIdx = forEachIdx
				}
			case countIdx >= 0:
				foundIdx = countIdx
			default:
				foundIdx = forEachIdx
			}

			newIndex, err := r.checkCountOrForEach(foundIdx, forEachIdx, lastIndex, block, blockLabels, desiredSequence, runner)
			if err != nil {
				return err
			}
			lastIndex = newIndex
		} else {
			idx := getIndex(want)
			// not present -> skip
			if idx < 0 {
				continue
			}
			newIndex, err := r.checkSingleArg(want, idx, lastIndex, block, blockLabels, desiredSequence, runner)
			if err != nil {
				return err
			}
			lastIndex = newIndex
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
