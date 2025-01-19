package rules

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/zclconf/go-cty/cty"
)

// TerraformProviderMinimumMajorVersionRule enforces that, if a provider's
// "version" constraint has a minimum (">=" or ">"), then it must also have
// a maximum ("<" or "<=") in the same string. It also flags purely
// maximum-only constraints with no minimum as invalid.
//
// Skips approximate versions (~> ...), exact (= ...) constraints, or no version at all.
type TerraformProviderMinimumMajorVersionRule struct {
	tflint.DefaultRule
}

func NewTerraformProviderMinimumMajorVersionRule() *TerraformProviderMinimumMajorVersionRule {
	return &TerraformProviderMinimumMajorVersionRule{}
}

func (r *TerraformProviderMinimumMajorVersionRule) Name() string {
	return "terraform_provider_minimum_major_version"
}

func (r *TerraformProviderMinimumMajorVersionRule) Enabled() bool {
	return true
}

func (r *TerraformProviderMinimumMajorVersionRule) Severity() tflint.Severity {
	return tflint.ERROR
}

func (r *TerraformProviderMinimumMajorVersionRule) Link() string {
	return ""
}

func (r *TerraformProviderMinimumMajorVersionRule) Check(runner tflint.Runner) error {
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
			// Skip parse-error files
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

func (r *TerraformProviderMinimumMajorVersionRule) processBody(
	body *hclsyntax.Body,
	runner tflint.Runner,
) error {
	for _, block := range body.Blocks {
		if block.Type == "terraform" {
			for _, sub := range block.Body.Blocks {
				if sub.Type == "required_providers" {
					if err := r.checkRequiredProvidersBlock(sub, runner); err != nil {
						return err
					}
				}
			}
		}
		// Recurse
		if err := r.processBody(block.Body, runner); err != nil {
			return err
		}
	}
	return nil
}

func (r *TerraformProviderMinimumMajorVersionRule) checkRequiredProvidersBlock(
	block *hclsyntax.Block,
	runner tflint.Runner,
) error {
	for providerName, attr := range block.Body.Attributes {
		if obj, ok := attr.Expr.(*hclsyntax.ObjectConsExpr); ok {
			if err := r.checkProviderObject(obj, providerName, runner); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *TerraformProviderMinimumMajorVersionRule) checkProviderObject(
	obj *hclsyntax.ObjectConsExpr,
	providerName string,
	runner tflint.Runner,
) error {
	var versionString string
	var versionRange hcl.Range

	// Gather attributes in lexical order
	type item struct {
		Key   string
		Value string
		Range hcl.Range
		Idx   int
	}
	var items []item
	for _, kv := range obj.Items {
		keyName := strings.TrimSpace(hcl.ExprAsKeyword(kv.KeyExpr))
		if keyName == "" {
			continue
		}
		rng := kv.KeyExpr.Range()
		valueSrc := strings.TrimSpace(hcl.ExprAsKeyword(kv.ValueExpr))
		items = append(items, item{Key: keyName, Value: valueSrc, Range: rng, Idx: rng.Start.Byte})
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].Idx < items[j].Idx
	})

	for _, it := range items {
		if it.Key == "version" {
			versionString = it.Value
			// If the expression is a string literal, we can decode it from kv.ValueExpr
			if lit, ok := kvByKey(obj.Items, "version"); ok {
				if strLit, ok := lit.ValueExpr.(*hclsyntax.LiteralValueExpr); ok {
					versionRange = strLit.Range() // Use the value's range instead of the key's
					v, diags := strLit.Value(nil)
					if !diags.HasErrors() && v.Type() == cty.String {
						versionString = v.AsString()
					}
				}
			} else {
				versionRange = it.Range
			}
			break
		}
	}

	// If no version => skip
	if versionString == "" {
		return nil
	}
	trimmed := strings.TrimSpace(versionString)

	// We only skip if the constraint is approximate (contains "~>"),
	// or if it starts with "=" (exact). But do NOT skip if it's ">=..." or "<=..."
	if strings.Contains(trimmed, "~>") {
		return nil
	}
	if strings.HasPrefix(trimmed, "=") {
		return nil
	}

	hasMin := strings.Contains(trimmed, ">") // catches > or >=
	hasMax := strings.Contains(trimmed, "<") // catches < or <=

	switch {
	case hasMin && hasMax:
		// Good: has both min and max
		return nil
	case hasMin && !hasMax:
		// Invalid: has a minimum but no max
		return runner.EmitIssue(
			r,
			fmt.Sprintf(
				"Provider '%s' has a minimum version constraint but no maximum (version=%q)",
				providerName, trimmed,
			),
			versionRange,
		)
	case !hasMin && hasMax:
		// Invalid: only a max constraint => no min
		return runner.EmitIssue(
			r,
			fmt.Sprintf(
				"Provider '%s' has only a maximum version constraint; a minimum version is required (version=%q)",
				providerName, trimmed,
			),
			versionRange,
		)
	default:
		// Neither min nor max => skip
		return nil
	}
}

// kvByKey is a helper that returns the first key-value item with the given name
func kvByKey(items []hclsyntax.ObjectConsItem, key string) (hclsyntax.ObjectConsItem, bool) {
	for _, it := range items {
		if strings.TrimSpace(hcl.ExprAsKeyword(it.KeyExpr)) == key {
			return it, true
		}
	}
	return hclsyntax.ObjectConsItem{}, false
}
