package rules

import (
	"reflect"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

func TestSplitAttrName(t *testing.T) {
	cases := []struct {
		Input  string
		Expect []string
	}{
		{
			Input:  "multiple[*].id",
			Expect: []string{"multiple", "[*]", "id"},
		},
		{
			Input:  "some.resource.name",
			Expect: []string{"some", "resource", "name"},
		},
		{
			Input:  "name_with[\"brackets\"]",
			Expect: []string{"name_with", "[\"brackets\"]"},
		},
		{
			Input:  "name.with.dots[and][brackets]",
			Expect: []string{"name", "with", "dots", "[and]", "[brackets]"},
		},
		{
			Input:  "data.resource[0].name",
			Expect: []string{"data", "resource", "[0]", "name"},
		},
	}

	for _, c := range cases {
		got := splitAttrName(c.Input)
		if !reflect.DeepEqual(got, c.Expect) {
			t.Errorf("splitAttrName(%q) = %v; want %v", c.Input, got, c.Expect)
		}
	}
}

func TestCanonicalizeTraversal(t *testing.T) {
	cases := []struct {
		name  string
		input hcl.Traversal
		want  hcl.Traversal
	}{
		{
			name: "Splat with attribute",
			input: hcl.Traversal{
				hcl.TraverseRoot{Name: "aws_instance"},
				hcl.TraverseAttr{Name: "multiple[*].id"},
			},
			want: hcl.Traversal{
				hcl.TraverseRoot{Name: "aws_instance"},
				hcl.TraverseAttr{Name: "multiple"},
				hcl.TraverseSplat{},
				hcl.TraverseAttr{Name: "id"},
			},
		},
		{
			name: "Splat only",
			input: hcl.Traversal{
				hcl.TraverseRoot{Name: "aws_instance"},
				hcl.TraverseAttr{Name: "multiple[*]"},
			},
			want: hcl.Traversal{
				hcl.TraverseRoot{Name: "aws_instance"},
				hcl.TraverseAttr{Name: "multiple"},
				hcl.TraverseSplat{},
			},
		},
		{
			name: "Splat separated",
			input: hcl.Traversal{
				hcl.TraverseRoot{Name: "aws_instance"},
				hcl.TraverseAttr{Name: "multiple"},
				hcl.TraverseSplat{},
			},
			want: hcl.Traversal{
				hcl.TraverseRoot{Name: "aws_instance"},
				hcl.TraverseAttr{Name: "multiple"},
				hcl.TraverseSplat{},
			},
		},
		{
			name: "Attributes only",
			input: hcl.Traversal{
				hcl.TraverseRoot{Name: "aws_instance"},
				hcl.TraverseAttr{Name: "multiple"},
				hcl.TraverseAttr{Name: "id"},
			},
			want: hcl.Traversal{
				hcl.TraverseRoot{Name: "aws_instance"},
				hcl.TraverseAttr{Name: "multiple"},
				hcl.TraverseAttr{Name: "id"},
			},
		},