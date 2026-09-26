package rules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func TestTerraformFilenameConvention(t *testing.T) {
	mismatchMessage := func(filename string) string {
		return "Terraform filename '" + filename + "' does not match the pattern " +
			"'<name>.tf' or '<name>.<area>.tf' (all snake_case alphanumerics)"
	}

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
					Message: mismatchMessage("Main.Example.tf"),
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
					Message: mismatchMessage("my_name.my_area.extra.tf"),
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
