package rules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func TestTerraformProviderMinimumMajorVersionRule(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Expected helper.Issues
	}{
		{
			Name: "skip approximate ~> version",
			Content: `
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.0"
    }
  }
}`,
			Expected: helper.Issues{},
		},
		{
			Name: "skip exact = version",
			Content: `
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "= 4.0"
    }
  }
}`,
			Expected: helper.Issues{},
		},
		{
			Name: "invalid only min version (>=4.0)",
			Content: `
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 4.0"
    }
  }
}`,
			Expected: helper.Issues{
				{
					Rule:    NewTerraformProviderMinimumMajorVersionRule(),
					Message: "Provider 'aws' has a minimum version constraint but no maximum (version=\">= 4.0\")",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 6, Column: 17},
						End:      hcl.Pos{Line: 6, Column: 25},
					},
				},
			},
		},
		{
			Name: "valid min+max (>=4.0, <5.0)",
			Content: `
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">=4.0, < 5.0.0"
    }
  }
}`,
			Expected: helper.Issues{},
		},
		{
			Name: "valid min+max (>4.0, <5.0)",
			Content: `
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "> 4.0, < 5.0.0"
    }
  }
}`,
			Expected: helper.Issues{},
		},
		{
			Name: "invalid only max (<4.0)",
			Content: `
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "< 4.0"
    }
  }
}`,
			Expected: helper.Issues{
				{
					Rule:    NewTerraformProviderMinimumMajorVersionRule(),
					Message: "Provider 'aws' has only a maximum version constraint; a minimum version is required (version=\"< 4.0\")",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 6, Column: 17},
						End:      hcl.Pos{Line: 6, Column: 24},
					},
				},
			},
		},
	}

	rule := NewTerraformProviderMinimumMajorVersionRule()
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			runner := helper.TestRunner(t, map[string]string{
				"resource.tf": tc.Content,
			})
			if err := rule.Check(runner); err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			helper.AssertIssues(t, tc.Expected, runner.Issues)
		})
	}
}
