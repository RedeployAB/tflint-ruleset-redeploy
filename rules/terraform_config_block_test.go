package rules

import (
  "testing"

  hcl "github.com/hashicorp/hcl/v2"
  "github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func TestTerraformConfigBlockFileRule(t *testing.T) {
  tests := []struct {
    Name     string
    FileMap  map[string]string
    Expected helper.Issues
  }{
    {
      Name: "Valid - terraform block in terraform.tf",
      FileMap: map[string]string{
        "terraform.tf": `
terraform {
  required_version = ">= 1.0"
}
`,
      },
      Expected: helper.Issues{},
    },
    {
      Name: "Invalid - terraform block in main.tf",
      FileMap: map[string]string{
        "main.tf": `
terraform {
  backend "s3" {}
}
`,
      },
      Expected: helper.Issues{
        {
          Rule:    NewTerraformConfigBlockFileRule(),
          Message: `"terraform" config block must appear in "terraform.tf", not "main.tf"`,
          Range: hcl.Range{
            Filename: "main.tf",
            Start:    hcl.Pos{Line: 2, Column: 1},
            End:      hcl.Pos{Line: 2, Column: 10},
          },
        },
      },
    },
  }

  rule := NewTerraformConfigBlockFileRule()
  for _, tc := range tests {
    t.Run(tc.Name, func(t *testing.T) {
      runner := helper.TestRunner(t, tc.FileMap)
      if err := rule.Check(runner); err != nil {
        t.Fatalf("Unexpected error: %v", err)
      }
      helper.AssertIssues(t, tc.Expected, runner.Issues)
    })
  }
}
