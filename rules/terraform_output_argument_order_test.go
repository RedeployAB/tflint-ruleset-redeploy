package rules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func TestTerraformOutputArgumentOrderRule(t *testing.T) {
	tests := []struct {
		Name    string
		Content string
		Issues  helper.Issues
	}{
		{
			Name: "OK - minimal (only value)",
			Content: `
output "min_output" {
  value = "just a test"
}
`,
			Issues: helper.Issues{},
		},
		{
			Name: "OK - all attributes in correct order",
			Content: `
output "full_output" {
  description = "some desc"
  value       = "some val"
  ephemeral   = true
  sensitive   = true
  depends_on  = []
}
`,
			Issues: helper.Issues{},
		},
		{
			Name: "NOT OK - 'sensitive' comes before 'ephemeral'",
			Content: `
output "bad_order" {

  description = "some desc"

  value = "some val"

  sensitive = true
  ephemeral = true
}
`,
			Issues: helper.Issues{
				{
					Rule:    NewTerraformOutputArgumentOrderRule(),
					Message: "Out-of-order argument 'sensitive'. Expected sequence: description, value, ephemeral, sensitive, depends_on",
					Range: hcl.Range{
						Filename: "outputs.tf",
						Start:    hcl.Pos{Line: 7, Column: 3},
						End:      hcl.Pos{Line: 7, Column: 17},
					},
				},
			},
		},
	}

	rule := NewTerraformOutputArgumentOrderRule()

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runner := helper.TestRunner(t, map[string]string{
				"outputs.tf": tc.Content,
			})

			if err := rule.Check(runner); err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			helper.AssertIssues(t, tc.Issues, runner.Issues)
		})
	}
}
