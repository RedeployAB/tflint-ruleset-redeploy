package rules

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

type TerraformProviderSourceOrderRule struct {
	tflint.DefaultRule
}

const argSource = "source"

func NewTerraformProviderSourceOrderRule() *TerraformProviderSourceOrderRule {
	return &TerraformProviderSourceOrderRule{}
}

func (r *TerraformProviderSourceOrderRule) Name() string {
	return "terraform_provider_source_order"
}

func (r *TerraformProviderSourceOrderRule) Enabled() bool {
	return true
}

func (r *TerraformProviderSourceOrderRule) Severity() tflint.Severity {
	return tflint.ERROR
}

func (r *TerraformProviderSourceOrderRule) Link() string {
	return GetRuleDocLink(r.Name())
}

func (r *TerraformProviderSourceOrderRule) Check(runner tflint.Runner) error {
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
			if err := r.processBody(body, runner); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *TerraformProviderSourceOrderRule) processBody(
	body *hclsyntax.Body,
	runner tflint.Runner,
) error {
	for _, block := range body.Blocks {
		// Looking for a "terraform" block
		if block.Type == "terraform" {
			// Inside it, we want sub-block "required_providers"
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

func (r *TerraformProviderSourceOrderRule) checkRequiredProvidersBlock(
	block *hclsyntax.Block,
	runner tflint.Runner,
) error {
	// Each attribute name in required_providers is a provider name.
	// The value is an object with `source` and/or `version` attributes.
	// We want to ensure `source` appears before `version`.
	for name, attr := range block.Body.Attributes {
		// attr.Expr should be an object, so parse it as an HCL syntax object
		// or skip if it's not an object expression.
		if obj, ok := attr.Expr.(*hclsyntax.ObjectConsExpr); ok {
			if err := r.checkProviderObject(obj, name, runner); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *TerraformProviderSourceOrderRule) checkProviderObject(
	obj *hclsyntax.ObjectConsExpr,
	providerName string,
	runner tflint.Runner,
) error {
	// Collect all attributes in lexical order
	type item struct {
		Key      string
		Index    int
		HCLRange hcl.Range
	}
	var items []item
	for _, kv := range obj.Items {
		keyName := strings.TrimSpace(hcl.ExprAsKeyword(kv.KeyExpr))
		// If key extraction fails, skip
		if keyName == "" {
			continue
		}
		items = append(items, item{
			Key:      keyName,
			Index:    kv.KeyExpr.Range().Start.Byte,
			HCLRange: kv.KeyExpr.Range(),
		})
	}
	// Sort items by their Index (position in file)
	sort.Slice(items, func(i, j int) bool {
		return items[i].Index < items[j].Index
	})

	var sourcePos, versionPos *item
	for i := range items {
		if items[i].Key == argSource {
			sourcePos = &items[i]
		}
		if items[i].Key == "version" {
			versionPos = &items[i]
		}
	}
	// If both are present, ensure source comes before version
	if sourcePos != nil && versionPos != nil {
		if sourcePos.Index > versionPos.Index {
			// Report an issue
			return runner.EmitIssue(
				r,
				fmt.Sprintf("Provider '%s': 'version' must appear after 'source'", providerName),
				versionPos.HCLRange,
			)
		}
	}
	return nil
}
