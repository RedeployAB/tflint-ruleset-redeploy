package rules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func TestTerraformFilenameConvention(t *testing.T) {
	tests := []struct {
		Name     string
		Filename string
		Expected helper.Issues
	}{
		{
			Name:     "valid lowercase filename",
			Filename: "main.example.tf",
			Expected: helper.Issues{},
		},
		{
			Name:     "invalid filename with uppercase",
			Filename: "Main.Example.tf",
			Expected: helper.Issues{
				{
					Rule:    NewTerraformFilenameConventionRule(),
					Message: "Terraform filename 'Main.Example.tf' does not match the pattern '<name>.<area>.tf' with snake_case alphanumerics only",
					Range: hcl.Range{
						Filename: "Main.Example.tf",
						Start:    hcl.Pos{Line: 0, Column: 0},
						End:      hcl.Pos{Line: 0, Column: 0},
					},
				},
			},
		},
		{
			Name:     "valid single name",
			Filename: "main.tf",
			Expected: helper.Issues{},
		},
		{
			Name:     "valid single name with underscore",
			Filename: "my_name.tf",
			Expected: helper.Issues{},
		},
		{
			Name:     "valid name with area containing underscore",
			Filename: "my_name.my_area.tf",
			Expected: helper.Issues{},
		},
		{
			Name:     "invalid multiple periods",
			Filename: "my_name.my_area.extra.tf",
			Expected: helper.Issues{
				{
					Rule:    NewTerraformFilenameConventionRule(),
					Message: "Terraform filename 'my_name.my_area.extra.tf' does not match the pattern '<name>.tf' or '<name>.<area>.tf' (all snake_case alphanumerics)",
					Range: hcl.Range{
						Filename: "my_name.my_area.extra.tf",
						Start:    hcl.Pos{Line: 0, Column: 0},
						End:      hcl.Pos{Line: 0, Column: 0},
					},
				},
			},
		},
	}

	rule := NewTerraformFilenameConventionRule()

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			// Provide some dummy HCL content just so TFLint can parse the file
			runner := helper.TestRunner(t, map[string]string{
				test.Filename: `# dummy Terraform content`,
			})

			if err := rule.Check(runner); err != nil {
				t.Fatalf("Unexpected error occurred: %s", err)
			}

			helper.AssertIssues(t, test.Expected, runner.Issues)
		})
	}
}
