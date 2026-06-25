package rules

import (
	"fmt"
	"sort"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// TerraformSingleTernaryPerLineRule checks that a single line contains at most
// one ternary (conditional) operation. Chained or nested ternaries on one line
// hurt readability and should be broken into local values.
type TerraformSingleTernaryPerLineRule struct {
	tflint.DefaultRule
}

func NewTerraformSingleTernaryPerLineRule() *TerraformSingleTernaryPerLineRule {
	return &TerraformSingleTernaryPerLineRule{}
}

func (r *TerraformSingleTernaryPerLineRule) Name() string {
	return "terraform_single_ternary_per_line"
}

func (r *TerraformSingleTernaryPerLineRule) Enabled() bool {
	return true
}

func (r *TerraformSingleTernaryPerLineRule) Severity() tflint.Severity {
	return tflint.WARNING
}

func (r *TerraformSingleTernaryPerLineRule) Link() string {
	return GetRuleDocLink(r.Name())
}

// ternaryCollector gathers every conditional (ternary) expression encountered
// while walking an HCL syntax tree.
type ternaryCollector struct {
	conditionals []*hclsyntax.ConditionalExpr
}

func (c *ternaryCollector) Enter(node hclsyntax.Node) hcl.Diagnostics {
	if cond, ok := node.(*hclsyntax.ConditionalExpr); ok {
		c.conditionals = append(c.conditionals, cond)
	}
	return nil
}

func (c *ternaryCollector) Exit(hclsyntax.Node) hcl.Diagnostics {
	return nil
}

func (r *TerraformSingleTernaryPerLineRule) Check(runner tflint.Runner) error {
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
		body, ok := syntaxFile.Body.(*hclsyntax.Body)
		if !ok {
			continue
		}
		if err := r.checkBody(body, runner); err != nil {
			return err
		}
	}
	return nil
}

func (r *TerraformSingleTernaryPerLineRule) checkBody(body *hclsyntax.Body, runner tflint.Runner) error {
	var collector ternaryCollector
	// Walk only returns diagnostics surfaced by the walker; ours never emits any.
	//nolint:errcheck // walker never produces diagnostics
	hclsyntax.Walk(body, &collector)

	// Group conditionals by the line on which they begin. A nested or chained
	// ternary appears as multiple ConditionalExpr nodes sharing a start line.
	byLine := make(map[int][]*hclsyntax.ConditionalExpr)
	for _, cond := range collector.conditionals {
		line := cond.Range().Start.Line
		byLine[line] = append(byLine[line], cond)
	}

	// Emit deterministically (sorted by line) so output is stable.
	lines := make([]int, 0, len(byLine))
	for line := range byLine {
		lines = append(lines, line)
	}
	sort.Ints(lines)

	for _, line := range lines {
		conds := byLine[line]
		if len(conds) < 2 {
			continue
		}
		// Anchor the issue at the first (outermost) ternary on the line.
		anchor := conds[0]
		for _, cond := range conds {
			if cond.Range().Start.Byte < anchor.Range().Start.Byte {
				anchor = cond
			}
		}
		if err := runner.EmitIssue(
			r,
			fmt.Sprintf(
				"Line contains %d ternary operations; use local values to keep at most one ternary per line",
				len(conds),
			),
			anchor.Range(),
		); err != nil {
			return err
		}
	}
	return nil
}
