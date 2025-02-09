package rules

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// variableBlock represents a variable block with its ordering metadata.
type variableBlock struct {
	Name       string
	HasDefault bool
	Range      hcl.Range
	Start      int
}

// TerraformVariableOrderRule checks that variables are ordered as:
// 1) Required variables (no default) in alphabetical order
// 2) Optional variables (has default) in alphabetical order
type TerraformVariableOrderRule struct {
	tflint.DefaultRule
}

// NewTerraformVariableOrderRule creates a new rule instance.
func NewTerraformVariableOrderRule() *TerraformVariableOrderRule {
	return &TerraformVariableOrderRule{}
}

// Name returns the rule name.
func (r *TerraformVariableOrderRule) Name() string {
	return "terraform_variable_order"
}

// Enabled returns whether the rule is enabled by default.
func (r *TerraformVariableOrderRule) Enabled() bool {
	return true
}

// Severity returns the severity of the rule.
func (r *TerraformVariableOrderRule) Severity() tflint.Severity {
	return tflint.ERROR
}

// Link returns the rule's reference link.
func (r *TerraformVariableOrderRule) Link() string {
	return GetRuleDocLink(r.Name())
}

// Check checks the order of variable blocks.
func (r *TerraformVariableOrderRule) Check(runner tflint.Runner) error {
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
			// Skip parse errors
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

func (r *TerraformVariableOrderRule) processBody(body *hclsyntax.Body, runner tflint.Runner) error {
	// Collect all variable blocks in lexical order
	var varBlocks []variableBlock

	for _, blk := range body.Blocks {
		if strings.EqualFold(blk.Type, TypeVariable) && len(blk.Labels) > 0 {
			vName := blk.Labels[0]

			// Check whether "default" attribute is present
			_, hasDefault := blk.Body.Attributes["default"]

			varBlocks = append(varBlocks, variableBlock{
				Name:       vName,
				HasDefault: hasDefault,
				Range:      blk.DefRange(),
				Start:      blk.DefRange().Start.Byte,
			})
		}
		// Recurse into nested blocks
		if err := r.processBody(blk.Body, runner); err != nil {
			return err
		}
	}

	// If no variables found here, nothing to check
	if len(varBlocks) == 0 {
		return nil
	}

	// Sort varBlocks by their starting position
	sort.Slice(varBlocks, func(i, j int) bool {
		return varBlocks[i].Start < varBlocks[j].Start
	})

	if err := r.checkVarBlockOrder(varBlocks, runner); err != nil {
		return err
	}

	return nil
}

func (r *TerraformVariableOrderRule) checkVarBlockOrder(varBlocks []variableBlock, runner tflint.Runner) error {
	lastRequiredName := ""
	lastOptionalName := ""
	seenOptional := false

	for _, vb := range varBlocks {
		if !vb.HasDefault {
			// Required variable: if we've already seen an optional variable or out-of-order name, emit an issue.
			if seenOptional || (lastRequiredName != "" && vb.Name < lastRequiredName) {
				return r.emitIssue(runner, vb.Range, vb.Name)
			}
			lastRequiredName = vb.Name
		} else {
			// Optional variable: check alphabetical order.
			if lastOptionalName != "" && vb.Name < lastOptionalName {
				return r.emitIssue(runner, vb.Range, vb.Name)
			}
			lastOptionalName = vb.Name
			seenOptional = true
		}
	}
	return nil
}

func (r *TerraformVariableOrderRule) emitIssue(
	runner tflint.Runner,
	rng hcl.Range,
	varName string,
) error {
	msg := fmt.Sprintf(
		`Out-of-order variable %q. Required variables must come first in alphabetical order, followed by optional variables in alphabetical order.`,
		varName,
	)
	return runner.EmitIssue(r, msg, rng)
}
