package rules

import (
	"fmt"
	"path/filepath"
	"regexp"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// TerraformLocalsFileRule enforces that any "locals" block appears only in files
// named "locals.tf" or "locals.*.tf".
type TerraformLocalsFileRule struct {
	tflint.DefaultRule
}

// NewTerraformLocalsFileRule creates a new rule instance
func NewTerraformLocalsFileRule() *TerraformLocalsFileRule {
	return &TerraformLocalsFileRule{}
}

func (r *TerraformLocalsFileRule) Name() string {
	return "terraform_locals_file"
}

func (r *TerraformLocalsFileRule) Enabled() bool {
	return true
}

func (r *TerraformLocalsFileRule) Severity() tflint.Severity {
	return tflint.ERROR
}

func (r *TerraformLocalsFileRule) Link() string {
	return GetRuleDocLink(r.Name())
}

var validLocalsPattern = regexp.MustCompile(`^locals(\.[^.]+)?\.tf$`)

func (r *TerraformLocalsFileRule) Check(runner tflint.Runner) error {
	files, err := runner.GetFiles()
	if err != nil {
		return err
	}

	for filename, hclFile := range files {
		if hclFile == nil || hclFile.Bytes == nil {
			continue
		}
		// If the file name *is* a valid locals file, no issues regardless of blocks.
		// Otherwise, parse and see if it has a "locals" block.
		if validLocalsPattern.MatchString(filepath.Base(filename)) {
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

func (r *TerraformLocalsFileRule) processBody(
	body *hclsyntax.Body,
	filename string,
	runner tflint.Runner,
) error {
	for _, blk := range body.Blocks {
		// Block types are always lowercase in Terraform
		if blk.Type == TypeLocals {
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

func (r *TerraformLocalsFileRule) emitIssue(
	runner tflint.Runner,
	rng hcl.Range,
	filename string,
) error {
	msg := fmt.Sprintf(`"locals" block must be placed in "locals.tf" or "locals.<area>.tf", not %q`, filepath.Base(filename))
	return runner.EmitIssue(r, msg, rng)
}
