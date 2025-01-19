package rules

import (
  "testing"

  hcl "github.com/hashicorp/hcl/v2"
  "github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func TestTerraformLocalsFileRule(t *testing.T) {
  tests := []struct {
    Name     string
    FileMap  map[string]string
    Expected helper.Issues
  }{
    {
      Name: "Valid - locals in locals.tf",
      FileMap: map[string]string{
        "locals.tf": readFixture(t, "locals_file_valid_main.tf"),
      },
      Expected: helper.Issues{},
    },
    {
      Name: "Valid - locals in locals.area.tf",
      FileMap: map[string]string{
        "locals.config.tf": readFixture(t, "locals_file_valid_area.tf"),
      },
      Expected: helper.Issues{},
    },
    {
      Name: "Invalid - locals in other file",
      FileMap: map[string]string{
        "main.tf": readFixture(t, "locals_file_invalid.tf"),
      },
      Expected: helper.Issues{
        {
          Rule:    NewTerraformLocalsFileRule(),
          Message: `"locals" block must be placed in "locals.tf" or "locals.<area>.tf", not "main.tf"`,
          Range: hcl.Range{
            Filename: "main.tf",
            Start:    hcl.Pos{Line: 1, Column: 1},
            End:      hcl.Pos{Line: 1, Column: 1},
          },
        },
      },
    },
  }

  rule := NewTerraformLocalsFileRule()
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
