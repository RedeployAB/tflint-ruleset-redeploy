package rules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func TestTerraformModuleDependsOnRule(t *testing.T) {
	tests := []struct {
		Name   string
		File   string
		Issues helper.Issues
	}{
		{
			Name:   "OK - module without depends_on",
			File:   "module_depends_on_ok.tf",
			Issues: helper.Issues{},
		},
		{
			Name: "NOT OK - module with depends_on",
			File: "module_depends_on_not_ok.tf",
			Issues: helper.Issues{
				{
					Rule:    NewTerraformModuleDependsOnRule(),
					Message: "'depends_on' should not be used for modules",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 3, Column: 3},
						End:      hcl.Pos{Line: 3, Column: 30},
					},
				},
			},
		},
	}

	rule := NewTerraformModuleDependsOnRule()

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			content := readFixture(t, tc.File)
			runner := helper.TestRunner(t, map[string]string{
				"resource.tf": content,
			})

			if err := rule.Check(runner); err != nil {
				t.Fatalf("Unexpected error: %s", err)
			}

			helper.AssertIssues(t, tc.Issues, runner.Issues)
		})
	}
}
