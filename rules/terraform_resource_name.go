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
	return GetRuleDocLink(r.Name())
}

func (r *TerraformResourceNameRule) Check(runner tflint.Runner) error {
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
		resourceType := block.Labels[0]
		resourceName := block.Labels[1]

		parts := strings.SplitN(resourceType, "_", 2)
		if len(parts) < 2 {
			continue
		}

		typeWithoutProvider := parts[1]
		if strings.Contains(resourceName, typeWithoutProvider) {
			if err := runner.EmitIssue(
				r,
				fmt.Sprintf("Resource name repeats resource type '%s'", typeWithoutProvider),
				block.DefRange,
			); err != nil {
				return err
			}
		}
	}

	return nil
}
