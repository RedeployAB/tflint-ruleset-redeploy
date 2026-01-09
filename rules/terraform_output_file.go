package rules

import (
	"fmt"
	"path/filepath"
	"regexp"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// TerraformOutputFileRule enforces that any "output" block appears only in files
// named "outputs.tf" or "outputs.*.tf".
type TerraformOutputFileRule struct {
	tflint.DefaultRule
}

// NewTerraformOutputFileRule creates a new rule instance.
func NewTerraformOutputFileRule() *TerraformOutputFileRule {
	return &TerraformOutputFileRule{}
}

// Name returns the rule name.
func (r *TerraformOutputFileRule) Name() string {
	return "terraform_output_file"
}

// Enabled returns whether the rule is enabled by default.
func (r *TerraformOutputFileRule) Enabled() bool {
	return true
}

// Severity returns the severity of the rule.
func (r *TerraformOutputFileRule) Severity() tflint.Severity {
	return tflint.ERROR
}

// Link returns the rule's reference link.
func (r *TerraformOutputFileRule) Link() string {
	return GetRuleDocLink(r.Name())
}

var validOutputsPattern = regexp.MustCompile(`^outputs(\.[^.]+)?\.tf$`)

// Check checks whether "output" blocks are in the correct files.
func (r *TerraformOutputFileRule) Check(runner tflint.Runner) error {
	files, err := runner.GetFiles()
	if err != nil {
		return err
	}

	for filename, hclFile := range files {
		if hclFile == nil || hclFile.Bytes == nil {
			continue
		}
		// If the file name is a valid outputs file, skip it
		if validOutputsPattern.MatchString(filepath.Base(filename)) {
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

func (r *TerraformOutputFileRule) processBody(
	body *hclsyntax.Body,
	filename string,
	runner tflint.Runner,
) error {
	for _, blk := range body.Blocks {
		// Block types are always lowercase in Terraform
		if blk.Type == TypeOutput {
			if err := r.emitIssue(runner, blk.DefRange(), filename); err != nil {
				return err
			}
		}
		// Recurse into nested blocks
		if err := r.processBody(blk.Body, filename, runner); err != nil {
			return err
		}
	}
	return nil
}

func (r *TerraformOutputFileRule) emitIssue(
	runner tflint.Runner,
	rng hcl.Range,
	filename string,
) error {
	msg := fmt.Sprintf(`"output" block must be placed in "outputs.tf" or "outputs.<area>.tf", not %q`, filepath.Base(filename))
	return runner.EmitIssue(r, msg, rng)
}
