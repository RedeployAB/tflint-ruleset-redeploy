package rules

import (
	"fmt"
	"strings"

	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

type TerraformResourceNameRule struct {
	tflint.DefaultRule
}

func NewTerraformResourceNameRule() *TerraformResourceNameRule {
	return &TerraformResourceNameRule{}
}

func (r *TerraformResourceNameRule) Name() string {
	return "terraform_resource_name"
}

func (r *TerraformResourceNameRule) Enabled() bool {
	return true
}

func (r *TerraformResourceNameRule) Severity() tflint.Severity {
	return tflint.ERROR
}

func (r *TerraformResourceNameRule) Link() string {
	return ""
}

func (r *TerraformResourceNameRule) Check(runner tflint.Runner) error {
	// We retrieve *all* resources (no specific type)
	resources, err := runner.GetResourceContent("", &hclext.BodySchema{}, nil)
	if err != nil {
		return err
	}

	for _, resource := range resources.Blocks {
		// Resource blocks typically have 2 labels: [type, name]
		if len(resource.Labels) < 2 {
			continue
		}

		resourceType := resource.Labels[0]
		resourceName := resource.Labels[1]

		// Remove the provider prefix if the resource type is like "azurerm_..."
		splitted := strings.SplitN(resourceType, "_", 2)
		if len(splitted) < 2 {
			continue
		}

		typeWithoutProvider := splitted[1]

		// If the resource name includes that substring, emit an issue
		if strings.Contains(resourceName, typeWithoutProvider) {
			if err := runner.EmitIssue(
				r,
				fmt.Sprintf("Resource name repeats resource type '%s'", typeWithoutProvider),
				resource.DefRange,
			); err != nil {
				return err
			}
		}
	}

	return nil
}
