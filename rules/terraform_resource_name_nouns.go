package rules

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

type TerraformResourceNameNounsRule struct {
	tflint.DefaultRule
}

func NewTerraformResourceNameNounsRule() *TerraformResourceNameNounsRule {
	return &TerraformResourceNameNounsRule{}
}

func (r *TerraformResourceNameNounsRule) Name() string {
	return "terraform_resource_name_nouns"
}

func (r *TerraformResourceNameNounsRule) Enabled() bool {
	return true
}

func (r *TerraformResourceNameNounsRule) Severity() tflint.Severity {
	return tflint.ERROR
}

func (r *TerraformResourceNameNounsRule) Link() string {
	return ""
}

func (r *TerraformResourceNameNounsRule) Check(runner tflint.Runner) error {
	// Retrieve all "resource" blocks
	content, err := runner.GetModuleContent(&hclext.BodySchema{
		Blocks: []hclext.BlockSchema{
			{
				Type:       "resource",
				LabelNames: []string{"type", "name"},
			},
		},
	}, nil)
	if err != nil {
		return err
	}

	for _, block := range content.Blocks {
		resourceName := block.Labels[1] // e.g., "switch" or "load_balancers"

		// Skip if the resource name is literally "this"
		if resourceName == "this" {
			continue
		}

		// Check if the name ends with 's' (non-singular)
		if strings.HasSuffix(resourceName, "s") {
			if err := runner.EmitIssue(
				r,
				fmt.Sprintf("Resource name '%s' must be singular", resourceName),
				block.DefRange,
			); err != nil {
				return err
			}
		}

		// Check if the name is at least 3 characters long (descriptive)
		if len(resourceName) < 3 {
			if err := runner.EmitIssue(
				r,
				fmt.Sprintf("Resource name '%s' is too short to be descriptive", resourceName),
				block.DefRange,
			); err != nil {
				return err
			}
		}
	}

	return nil
}
