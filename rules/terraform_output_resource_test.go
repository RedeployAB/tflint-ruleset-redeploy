package rules

import (
	"testing"
	"reflect"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func TestTerraformOutputResourceRule(t *testing.T) {
	tests := []struct {
		Name    string
		Content string
		Issues  helper.Issues
	}{
		{
			Name: "OK - references a single attribute",
			Content: `
resource "aws_instance" "example" {}

output "out_ok" {
	value = aws_instance.example.id
}
`,
			Issues: helper.Issues{},
		},
		{
			Name: "NOT OK - references entire resource",
			Content: `
resource "aws_instance" "example" {}

output "out_bad" {
	value = aws_instance.example
}
`,
			Issues: helper.Issues{
				{
					Rule:    NewTerraformOutputResourceRule(),
					Message: "Output is referencing the entire resource or data, rather than a specific attribute. This can cause schema issues.",
					Range: hcl.Range{
						Filename: "main.tf",
						Start:    hcl.Pos{Line: 5, Column: 2},
						End:      hcl.Pos{Line: 5, Column: 30},
					},
				},
			},
		},
		{
			Name: "NOT OK - references entire data resource",
			Content: `
data "aws_caller_identity" "current" {}

output "caller" {
	value = data.aws_caller_identity.current
}
`,
			Issues: helper.Issues{
				{
					Rule:    NewTerraformOutputResourceRule(),
					Message: "Output is referencing the entire resource or data, rather than a specific attribute. This can cause schema issues.",
					Range: hcl.Range{
						Filename: "main.tf",
						Start:    hcl.Pos{Line: 5, Column: 2},
						End:      hcl.Pos{Line: 5, Column: 42},
					},
				},
			},
		},
		{
			Name: "OK - ternary with variable check referencing resource attribute",
			Content: `
variable "aks_identity_type" {}

resource "azurerm_user_assigned_identity" "aks" {
	name = "dummy"
}

output "some_output" {
	value = var.aks_identity_type == "UserAssigned" ? azurerm_user_assigned_identity.aks[0].client_id : null
}
`,
			Issues: helper.Issues{},
		},
		{
			Name: "OK - references ephemeral resource attribute",
			Content: `
resource "ephemeral" "test" {}

output "ok_ephemeral" {
	value = ephemeral.test.id
}
`,
			Issues: helper.Issues{},
		},
		{
			Name: "NOT OK - references entire ephemeral resource",
			Content: `
resource "ephemeral" "test" {}

output "bad_ephemeral" {
	value = ephemeral.test
}
`,
			Issues: helper.Issues{
				{
					Rule:    NewTerraformOutputResourceRule(),
					Message: "Output is referencing the entire resource or data, rather than a specific attribute. This can cause schema issues.",
					Range: hcl.Range{
						Filename: "main.tf",
						Start:    hcl.Pos{Line: 5, Column: 2},
						End:      hcl.Pos{Line: 5, Column: 24},
					},
				},
			},
		},
		{
			Name: "OK - references multiple instances with explicit index",
			Content: `
resource "aws_instance" "multiple" {
	count = 2
}

output "indexed_output" {
	value = aws_instance.multiple[1].id
}
`,
			Issues: helper.Issues{},
		},
		{
			Name: "OK - references multiple instances with splat",
			Content: `
resource "aws_instance" "multiple" {
	count = 2
}

output "splat_output" {
	value = aws_instance.multiple[*].id
}
`,
			Issues: helper.Issues{},
		},
		{
			Name: "NOT OK - references entire resource with index",
			Content: `
resource "aws_instance" "indexed" {
  count = 2
}

output "bad_index" {
  value = aws_instance.indexed[0]
}
`,
			Issues: helper.Issues{
				{
					Rule:    NewTerraformOutputResourceRule(),
					Message: "Output is referencing the entire resource or data, rather than a specific attribute. This can cause schema issues.",
					Range: hcl.Range{
						Filename: "main.tf",
						Start:    hcl.Pos{Line: 7, Column: 3},
						End:      hcl.Pos{Line: 7, Column: 34},
					},
				},
			},
		},
		{
			Name: "NOT OK - references entire resource with splat",
			Content: `
resource "aws_instance" "splat" {
  count = 2
}

output "bad_splat" {
  value = aws_instance.splat[*]
}
`,
			Issues: helper.Issues{
				{
					Rule:    NewTerraformOutputResourceRule(),
					Message: "Output is referencing the entire resource or data, rather than a specific attribute. This can cause schema issues.",
					Range: hcl.Range{
						Filename: "main.tf",
						Start:    hcl.Pos{Line: 7, Column: 3},
						End:      hcl.Pos{Line: 7, Column: 32},
					},
				},
			},
		},
	}

	rule := NewTerraformOutputResourceRule()
	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			runner := helper.TestRunner(t, map[string]string{
				"main.tf": tc.Content,
			})
			if err := rule.Check(runner); err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			helper.AssertIssues(t, tc.Issues, runner.Issues)
		})
	}
}

func TestGatherTraversals(t *testing.T) {
	rule := NewTerraformOutputResourceRule()
	cases := []struct {
		name     string
		exprStr  string
		expected []hcl.Traversal
	}{
		{
			name:    "simple attribute access",
			exprStr: "aws_instance.example.id",
			expected: []hcl.Traversal{
				{
					hcl.TraverseRoot{Name: "aws_instance"},
					hcl.TraverseAttr{Name: "example"},
					hcl.TraverseAttr{Name: "id"},
				},
			},
		},
		{
			name:    "resource reference without attribute",
			exprStr: "aws_instance.example",
			expected: []hcl.Traversal{
				{
					hcl.TraverseRoot{Name: "aws_instance"},
					hcl.TraverseAttr{Name: "example"},
				},
			},
		},
		{
			name:    "splat without item",
			exprStr: "aws_instance.splat[*]",
			expected: []hcl.Traversal{
				{
					hcl.TraverseRoot{Name: "aws_instance"},
					hcl.TraverseAttr{Name: "splat"},
					hcl.TraverseSplat{},
				},
			},
		},
		{
			name:    "conditional expression with prefix",
			exprStr: "true ? aws_instance.example.id : aws_instance.example",
			expected: []hcl.Traversal{
				{
					hcl.TraverseRoot{Name: "aws_instance"},
					hcl.TraverseAttr{Name: "example"},
					hcl.TraverseAttr{Name: "id"},
				},
			},
		},
		{
			name:    "tuple expression with prefix",
			exprStr: "[aws_instance.example.id, aws_instance.example]",
			expected: []hcl.Traversal{
				{
					hcl.TraverseRoot{Name: "aws_instance"},
					hcl.TraverseAttr{Name: "example"},
					hcl.TraverseAttr{Name: "id"},
				},
			},
		},
		{
			name:    "object expression with mixed references",
			exprStr: `{"a": aws_instance.example, "b": aws_instance.example.id}`,
			expected: []hcl.Traversal{
				{
					hcl.TraverseRoot{Name: "aws_instance"},
					hcl.TraverseAttr{Name: "example"},
					hcl.TraverseAttr{Name: "id"},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			expr, diags := hclsyntax.ParseExpression([]byte(tc.exprStr), "test.hcl", hcl.InitialPos)
			if diags.HasErrors() {
				t.Fatalf("Failed to parse expression %q: %s", tc.exprStr, diags.Error())
			}
			got := rule.gatherTraversals(expr)
			if !reflect.DeepEqual(got, tc.expected) {
				t.Errorf("For expression %q, expected traversals:\n%#v\ngot:\n%#v", tc.exprStr, tc.expected, got)
			}
		})
	}
}
