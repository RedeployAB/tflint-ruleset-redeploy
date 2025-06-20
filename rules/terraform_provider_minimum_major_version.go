package rules

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// TerraformProviderMinimumMajorVersionRule enforces that, if a provider's
// "version" constraint has a minimum (">=" or ">"), then it must also have
// a maximum ("<" or "<=") in the same string. It also flags purely
// maximum-only constraints with no minimum as invalid.
//
// Skips approximate versions (~> ...), exact (= ...) constraints, constraints containing '!=', or no version at all.
type TerraformProviderMinimumMajorVersionRule struct {
	tflint.DefaultRule
}

const argVersion = "version"

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
	return GetRuleDocLink(r.Name())
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

	for _, kv := range obj.Items {
		keyName := strings.TrimSpace(hcl.ExprAsKeyword(kv.KeyExpr))
		if keyName == "" {
			// Skip if key extraction fails
			continue
		}
		if keyName != argVersion {
			continue
		}

		var tmp string
		if err := runner.EvaluateExpr(kv.ValueExpr, &tmp, nil); err != nil {
			// If for some reason it isn't a string, skip this provider
			return nil
		}
		// EvaluateExpr succeeded => store it
		versionString = tmp
		versionRange = kv.ValueExpr.Range()
		break
	}

	// If no version => skip
	if versionString == "" {
		return nil
	}
	trimmed := strings.TrimSpace(versionString)

	// We only skip if the constraint is approximate (contains "~>"),
	// or if the constraint includes "!=" (exclusion).
	if strings.Contains(trimmed, "!=") {
		return nil
	}
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
		// Invalid: has min but no max
		return runner.EmitIssue(
			r,
			fmt.Sprintf("Provider '%s' has a minimum version constraint but no maximum (version=%q)", providerName, versionString),
			versionRange,
		)
	case !hasMin && hasMax:
		// Invalid: has max but no min
		return runner.EmitIssue(
			r,
			fmt.Sprintf("Provider '%s' has only a maximum version constraint; a minimum version is required (version=%q)", providerName, versionString),
			versionRange,
		)
	default:
		// Neither min nor max constraints; possibly exact version?
		// Skip
		return nil
	}
}
