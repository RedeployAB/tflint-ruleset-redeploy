package rules

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// TerraformOutputOrderRule checks that outputs are alphabetically ordered by name
type TerraformOutputOrderRule struct {
	tflint.DefaultRule
}

// NewTerraformOutputOrderRule creates a new rule instance
func NewTerraformOutputOrderRule() *TerraformOutputOrderRule {
	return &TerraformOutputOrderRule{}
}

// Name returns the rule name
func (r *TerraformOutputOrderRule) Name() string {
	return "terraform_output_order"
}

// Enabled returns whether the rule is enabled by default
func (r *TerraformOutputOrderRule) Enabled() bool {
	return true
}

// Severity returns the severity of the rule
func (r *TerraformOutputOrderRule) Severity() tflint.Severity {
	return tflint.ERROR
}

// Link returns the rule's reference link
func (r *TerraformOutputOrderRule) Link() string {
	return GetRuleDocLink(r.Name())
}

// Check checks that output blocks are ordered alphabetically by name
func (r *TerraformOutputOrderRule) Check(runner tflint.Runner) error {
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
			// skip parse errors
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

func (r *TerraformOutputOrderRule) processBody(body *hclsyntax.Body, runner tflint.Runner) error {
	// Collect all output blocks in lexical order
	var outputs []struct {
		Name  string
		Range hcl.Range
		Start int
	}

	for _, blk := range body.Blocks {
		if strings.EqualFold(blk.Type, TypeOutput) && len(blk.Labels) > 0 {
			outName := blk.Labels[0]
			outputs = append(outputs, struct {
				Name  string
				Range hcl.Range
				Start int
			}{
				Name:  outName,
				Range: blk.DefRange(),
				Start: blk.DefRange().Start.Byte,
			})
		}
		// Recurse into nested blocks
		if err := r.processBody(blk.Body, runner); err != nil {
			return err
		}
	}

	// If no outputs found here, nothing to check
	if len(outputs) == 0 {
		return nil
	}

	// Ensure outputs are sorted by their position in the file
	for i := 0; i < len(outputs)-1; i++ {
		for j := i + 1; j < len(outputs); j++ {
			if outputs[j].Start < outputs[i].Start {
				outputs[i], outputs[j] = outputs[j], outputs[i]
			}
		}
	}

	// Check alphabetical order
	lastName := ""
	for _, o := range outputs {
		if lastName != "" && o.Name < lastName {
			if err := r.emitIssue(runner, o.Range, o.Name); err != nil {
				return err
			}
		}
		lastName = o.Name
	}

	return nil
}

func (r *TerraformOutputOrderRule) emitIssue(
	runner tflint.Runner,
	rng hcl.Range,
	outName string,
) error {
	msg := fmt.Sprintf(
		`Out-of-order output %q. Output blocks must be alphabetically ordered by name.`,
		outName,
	)
	return runner.EmitIssue(r, msg, rng)
}
