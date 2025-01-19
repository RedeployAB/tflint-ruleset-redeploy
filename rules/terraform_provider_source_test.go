package rules

import (
  "testing"

  hcl "github.com/hashicorp/hcl/v2"
  "github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func TestTerraformProviderSourceOrderRule(t *testing.T) {
  tests := []struct {
    Name     string
    Content  string
    Expected helper.Issues
  }{
    {
      Name:    "Valid source/version order",
      Content: readFixture(t, "provider_source_order_valid.tf"),
      Expected: helper.Issues{},
    },
    {
      Name:    "Invalid order (version before source)",
      Content: readFixture(t, "provider_source_order_invalid.tf"),
      Expected: helper.Issues{
        {
          Rule:    NewTerraformProviderSourceOrderRule(),
          Message: "Provider 'aws': 'version' must appear after 'source'",
          Range: hcl.Range{
            Filename: "resource.tf",
            Start:    hcl.Pos{Line: 4, Column: 5},
            End:      hcl.Pos{Line: 4, Column: 12},
          },
        },
      },
    },
  }

  rule := NewTerraformProviderSourceOrderRule()

  for _, tc := range tests {
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
