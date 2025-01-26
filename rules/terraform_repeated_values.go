package rules

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// TerraformRepeatedValuesRule checks if certain literal values
// are repeated above a threshold, recommending use of locals instead.
// The threshold can be configured through a rule setting, default=3.
type TerraformRepeatedValuesRule struct {
	tflint.DefaultRule

	// Threshold for repetition (default: 3)
	Threshold int
}

// For runner.DecodeRuleConfig parsing:
type terraformRepeatedValuesRuleConfig struct {
	Threshold int `hclext:"threshold,optional"`
}

func NewTerraformRepeatedValuesRule() *TerraformRepeatedValuesRule {
	return &TerraformRepeatedValuesRule{
		Threshold: 3, // default threshold
	}
}

func (r *TerraformRepeatedValuesRule) Name() string {
	return "terraform_repeated_values"
}

func (r *TerraformRepeatedValuesRule) Enabled() bool {
	return true
}

// We'll use a WARNING severity to encourage but not strictly enforce
func (r *TerraformRepeatedValuesRule) Severity() tflint.Severity {
	return tflint.WARNING
}

func (r *TerraformRepeatedValuesRule) Link() string {
	return ""
}

func (r *TerraformRepeatedValuesRule) Check(runner tflint.Runner) error {
	var cfg terraformRepeatedValuesRuleConfig
	// Attempt to decode user settings from .tflint.hcl
	if err := runner.DecodeRuleConfig(r.Name(), &cfg); err != nil {
		return err
	}
	// If user-specified threshold is > 0, override default
	if cfg.Threshold > 0 {
		r.Threshold = cfg.Threshold
	}

	files, err := runner.GetFiles()
	if err != nil {
		return err
	}

	// Collect literal occurrences across all resource blocks
	literalOccurrences := make(map[string][]literalOccurrence)

	for filename, hclFile := range files {
		if hclFile == nil || hclFile.Bytes == nil {
			continue
		}
		syntaxFile, diags := hclsyntax.ParseConfig(hclFile.Bytes, filename, hcl.InitialPos)
		if diags.HasErrors() {
			continue // skip parse errors
		}
		if body, ok := syntaxFile.Body.(*hclsyntax.Body); ok {
			if err := r.collectResourceLiterals(body, filename, hclFile.Bytes, literalOccurrences); err != nil {
				return err
			}
		}
	}

	// Check for literals repeated above the threshold
	for literal, occurrences := range literalOccurrences {
		if len(occurrences) >= r.Threshold {
			// Emit an issue for each occurrence
			msg := fmt.Sprintf("Value %q repeated %d times. Consider a local variable.", literal, len(occurrences))
			for _, occ := range occurrences {
				if err := runner.EmitIssue(r, msg, occ.Range); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// literalOccurrence helps us track where a literal is found
type literalOccurrence struct {
	Filename string
	Range    hcl.Range
}

// collectResourceLiterals recursively looks for "resource" blocks and
// extracts literal values from resource attributes (excluding booleans).
func (r *TerraformRepeatedValuesRule) collectResourceLiterals(
	body *hclsyntax.Body,
	filename string,
	fileBytes []byte,
	out map[string][]literalOccurrence,
) error {
	for _, block := range body.Blocks {
		blkType := strings.ToLower(block.Type)
		if blkType == TypeResource {
			// Gather from this resource
			for _, attr := range block.Body.Attributes {
				if err := r.collectLiteral(attr, filename, fileBytes, out); err != nil {
					return err
				}
			}
			// Recursively gather from child blocks
			for _, child := range block.Body.Blocks {
				// Gather from child block's attributes
				for _, attr := range child.Body.Attributes {
					if err := r.collectLiteral(attr, filename, fileBytes, out); err != nil {
						return err
					}
				}
				// Deeper recursion if needed
				if err := r.collectResourceLiterals(child.Body, filename, fileBytes, out); err != nil {
					return err
				}
			}
		} else {
			// Go deeper looking for resources
			if err := r.collectResourceLiterals(block.Body, filename, fileBytes, out); err != nil {
				return err
			}
		}
	}
	return nil
}

// collectLiteral checks if an attribute is a literal string
func (r *TerraformRepeatedValuesRule) collectLiteral(
	attr *hclsyntax.Attribute,
	filename string,
	fileBytes []byte,
	out map[string][]literalOccurrence,
) error {
	switch attr.Expr.(type) {
	case *hclsyntax.LiteralValueExpr:
		// Parse raw text. If it's "true"/"false", skip; else treat as a literal.
		raw := GetAttributeRawText(attr, fileBytes)
		raw = strings.TrimSpace(raw)
		lowered := strings.ToLower(raw)
		if lowered == "true" || lowered == "false" {
			// skip booleans
			return nil
		}
		// If it's not empty, consider it repeated
		if raw != "" {
			out[raw] = append(out[raw], literalOccurrence{
				Filename: filename,
				Range:    attr.Range(),
			})
		}
	case *hclsyntax.TemplateExpr:
		// Handle simple templates without interpolation
		valText := GetAttributeRawText(attr, fileBytes)
		if !strings.Contains(valText, "${") {
			val := strings.Trim(valText, `"`)
			if len(val) > 0 {
				out[val] = append(out[val], literalOccurrence{
					Filename: filename,
					Range:    attr.Range(),
				})
			}
		}
	}
	return nil
}
