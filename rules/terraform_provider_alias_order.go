package rules

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// TerraformProviderAliasOrderRule enforces the HashiCorp style guide rules for
// provider aliasing: the `alias` argument must be the first argument of an
// aliased provider block, and a default (un-aliased) provider block must be
// declared before any aliased instance of the same provider.
type TerraformProviderAliasOrderRule struct {
	tflint.DefaultRule
}

const argAlias = "alias"

func NewTerraformProviderAliasOrderRule() *TerraformProviderAliasOrderRule {
	return &TerraformProviderAliasOrderRule{}
}

func (r *TerraformProviderAliasOrderRule) Name() string {
	return "terraform_provider_alias_order"
}

func (r *TerraformProviderAliasOrderRule) Enabled() bool {
	return true
}

func (r *TerraformProviderAliasOrderRule) Severity() tflint.Severity {
	return tflint.ERROR
}

func (r *TerraformProviderAliasOrderRule) Link() string {
	return GetRuleDocLink(r.Name())
}

func (r *TerraformProviderAliasOrderRule) Check(runner tflint.Runner) error {
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
		if err := r.checkFile(body, runner); err != nil {
			return err
		}
	}
	return nil
}

func (r *TerraformProviderAliasOrderRule) checkFile(body *hclsyntax.Body, runner tflint.Runner) error {
	// defaultStart tracks, per provider name, the start byte of the default
	// (un-aliased) provider block within this file. Ordering is only defined
	// within a single file, so the check is scoped per file.
	defaultStart := make(map[string]int)
	for _, block := range body.Blocks {
		if block.Type != TypeProvider || len(block.Labels) == 0 {
			continue
		}
		if _, isAliased := block.Body.Attributes[argAlias]; !isAliased {
			name := block.Labels[0]
			start := block.DefRange().Start.Byte
			if cur, ok := defaultStart[name]; !ok || start < cur {
				defaultStart[name] = start
			}
		}
	}

	for _, block := range body.Blocks {
		if block.Type != TypeProvider || len(block.Labels) == 0 {
			continue
		}
		name := block.Labels[0]
		aliasAttr, isAliased := block.Body.Attributes[argAlias]
		if !isAliased {
			continue
		}
		if err := r.checkAliasFirst(block, aliasAttr, name, runner); err != nil {
			return err
		}
		// A default block exists later in the file than this aliased block.
		if start, ok := defaultStart[name]; ok && block.DefRange().Start.Byte < start {
			if err := runner.EmitIssue(
				r,
				fmt.Sprintf("Provider '%s': default (un-aliased) provider must be declared before aliased providers", name),
				block.DefRange(),
			); err != nil {
				return err
			}
		}
	}
	return nil
}

// checkAliasFirst reports an issue when any other argument or nested block
// begins before the `alias` argument within the provider block.
func (r *TerraformProviderAliasOrderRule) checkAliasFirst(
	block *hclsyntax.Block,
	aliasAttr *hclsyntax.Attribute,
	name string,
	runner tflint.Runner,
) error {
	aliasStart := aliasAttr.NameRange.Start.Byte

	aliasIsFirst := true
	for argName, attr := range block.Body.Attributes {
		if argName != argAlias && attr.NameRange.Start.Byte < aliasStart {
			aliasIsFirst = false
		}
	}
	for _, nested := range block.Body.Blocks {
		if nested.DefRange().Start.Byte < aliasStart {
			aliasIsFirst = false
		}
	}

	if !aliasIsFirst {
		return runner.EmitIssue(
			r,
			fmt.Sprintf("Provider '%s': 'alias' must be the first argument in the provider block", name),
			aliasAttr.NameRange,
		)
	}
	return nil
}
