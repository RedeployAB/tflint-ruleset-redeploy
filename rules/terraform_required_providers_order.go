package rules

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// TerraformRequiredProvidersOrderRule checks that providers in required_providers
// are listed in alphabetical order.
//
// Source: AVM TFNFR26 - https://azure.github.io/Azure-Verified-Modules/specs/tf/res/
type TerraformRequiredProvidersOrderRule struct {
	tflint.DefaultRule
}

func NewTerraformRequiredProvidersOrderRule() *TerraformRequiredProvidersOrderRule {
	return &TerraformRequiredProvidersOrderRule{}
}

func (r *TerraformRequiredProvidersOrderRule) Name() string {
	return "terraform_required_providers_order"
}

func (r *TerraformRequiredProvidersOrderRule) Enabled() bool {
	return true
}

func (r *TerraformRequiredProvidersOrderRule) Severity() tflint.Severity {
	return tflint.ERROR
}

func (r *TerraformRequiredProvidersOrderRule) Link() string {
	return GetRuleDocLink(r.Name())
}

func (r *TerraformRequiredProvidersOrderRule) Check(runner tflint.Runner) error {
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

func (r *TerraformRequiredProvidersOrderRule) processBody(
	body *hclsyntax.Body,
	runner tflint.Runner,
) error {
	for _, block := range body.Blocks {
		if block.Type != TypeTerraform {
			continue
		}

		for _, sub := range block.Body.Blocks {
			if sub.Type != TypeRequiredProviders {
				continue
			}
			if err := r.checkRequiredProvidersOrder(sub, runner); err != nil {
				return err
			}
		}
	}

	return nil
}

// providerItem represents a provider declaration in required_providers
type providerItem struct {
	Name  string
	Range hcl.Range
	Start int
}

func (r *TerraformRequiredProvidersOrderRule) checkRequiredProvidersOrder(
	block *hclsyntax.Block,
	runner tflint.Runner,
) error {
	providers := r.collectProviders(block)
	if len(providers) <= 1 {
		return nil
	}

	// Sort by position in file to get actual order
	sort.Slice(providers, func(i, j int) bool {
		return providers[i].Start < providers[j].Start
	})

	// Create expected order (alphabetical by name, case-insensitive)
	expectedOrder := r.buildExpectedOrder(providers)

	// Check if actual order matches expected order
	for i, provider := range providers {
		if provider.Name == expectedOrder[i] {
			continue
		}

		return runner.EmitIssueWithFix(
			r,
			fmt.Sprintf(
				"Provider '%s' is out of alphabetical order. Expected order: %s",
				provider.Name,
				strings.Join(expectedOrder, ", "),
			),
			provider.Range,
			func(f tflint.Fixer) error {
				return r.fixProviderOrder(f, block, providers, expectedOrder)
			},
		)
	}

	return nil
}

func (r *TerraformRequiredProvidersOrderRule) collectProviders(
	block *hclsyntax.Block,
) []providerItem {
	providers := make([]providerItem, 0, len(block.Body.Attributes))
	for name, attr := range block.Body.Attributes {
		providers = append(providers, providerItem{
			Name:  name,
			Range: attr.Range(),
			Start: attr.Range().Start.Byte,
		})
	}
	return providers
}

func (r *TerraformRequiredProvidersOrderRule) buildExpectedOrder(
	providers []providerItem,
) []string {
	expectedOrder := make([]string, len(providers))
	for i, p := range providers {
		expectedOrder[i] = p.Name
	}
	sort.Slice(expectedOrder, func(i, j int) bool {
		return strings.ToLower(expectedOrder[i]) < strings.ToLower(expectedOrder[j])
	})
	return expectedOrder
}

func (r *TerraformRequiredProvidersOrderRule) fixProviderOrder(
	f tflint.Fixer,
	block *hclsyntax.Block,
	providers []providerItem,
	expectedOrder []string,
) error {
	if strings.HasSuffix(block.DefRange().Filename, ".json") {
		return tflint.ErrFixNotSupported
	}

	providerTexts := r.extractProviderTexts(f, providers)
	spacingMap := r.buildSpacingMap(f, providers)

	var result strings.Builder
	for i, name := range expectedOrder {
		if i > 0 {
			spacing := r.findSpacing(spacingMap, expectedOrder[i-1], name)
			result.WriteString(spacing)
		}
		result.WriteString(providerTexts[name])
	}

	fullRange := hcl.Range{
		Filename: providers[0].Range.Filename,
		Start:    providers[0].Range.Start,
		End:      providers[len(providers)-1].Range.End,
	}

	return f.ReplaceText(fullRange, result.String())
}

func (r *TerraformRequiredProvidersOrderRule) extractProviderTexts(
	f tflint.Fixer,
	providers []providerItem,
) map[string]string {
	providerTexts := make(map[string]string, len(providers))
	for _, provider := range providers {
		text := f.TextAt(provider.Range)
		providerTexts[provider.Name] = string(text.Bytes)
	}
	return providerTexts
}

func (r *TerraformRequiredProvidersOrderRule) buildSpacingMap(
	f tflint.Fixer,
	providers []providerItem,
) map[string]string {
	spacingMap := make(map[string]string, len(providers)-1)
	for i := 1; i < len(providers); i++ {
		betweenRange := hcl.Range{
			Filename: providers[i-1].Range.Filename,
			Start:    providers[i-1].Range.End,
			End:      providers[i].Range.Start,
		}
		betweenText := f.TextAt(betweenRange)
		key := providers[i-1].Name + "|||" + providers[i].Name
		spacingMap[key] = string(betweenText.Bytes)
	}
	return spacingMap
}

func (r *TerraformRequiredProvidersOrderRule) findSpacing(
	spacingMap map[string]string,
	prevName, currName string,
) string {
	if s, ok := spacingMap[prevName+"|||"+currName]; ok {
		return s
	}
	if s, ok := spacingMap[currName+"|||"+prevName]; ok {
		return s
	}
	return "\n\n"
}
