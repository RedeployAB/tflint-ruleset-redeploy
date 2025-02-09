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
		{
			name: "Number index",
			input: hcl.Traversal{
				hcl.TraverseRoot{Name: "aws_instance"},
				hcl.TraverseAttr{Name: "multiple"},
				hcl.TraverseIndex{Key: cty.NumberIntVal(1)},
				hcl.TraverseAttr{Name: "id"},
			},
			want: hcl.Traversal{
				hcl.TraverseRoot{Name: "aws_instance"},
				hcl.TraverseAttr{Name: "multiple"},
				hcl.TraverseIndex{Key: cty.NumberIntVal(1)},
				hcl.TraverseAttr{Name: "id"},
			},
		},
		{
			name: "Key index",
			input: hcl.Traversal{
				hcl.TraverseRoot{Name: "aws_instance"},
				hcl.TraverseAttr{Name: "multiple"},
				hcl.TraverseIndex{Key: cty.StringVal("key")},
				hcl.TraverseAttr{Name: "id"},
			},
			want: hcl.Traversal{
				hcl.TraverseRoot{Name: "aws_instance"},
				hcl.TraverseAttr{Name: "multiple"},
				hcl.TraverseIndex{Key: cty.StringVal("key")},
				hcl.TraverseAttr{Name: "id"},
			},
		},
		{
			name: "Splat only with 'splat'",
			input: hcl.Traversal{
				hcl.TraverseRoot{Name: "aws_instance"},
				hcl.TraverseAttr{Name: "name[*]"},
			},
			want: hcl.Traversal{
				hcl.TraverseRoot{Name: "aws_instance"},
				hcl.TraverseAttr{Name: "name"},
				hcl.TraverseSplat{},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := canonicalizeTraversal(c.input)
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("canonicalizeTraversal(%v) => %v, want %v", c.input, got, c.want)
			}
		})
	}
}

func TestFilterPrefixTraversals(t *testing.T) {
	// Original test cases
	t1 := hcl.Traversal{
		hcl.TraverseRoot{Name: "aws_instance"},
		hcl.TraverseAttr{Name: "example"},
	}

	t2 := hcl.Traversal{
		hcl.TraverseRoot{Name: "aws_instance"},
		hcl.TraverseAttr{Name: "example"},
		hcl.TraverseAttr{Name: "id"},
	}

	list := []hcl.Traversal{t1, t2}
	got := filterPrefixTraversals(list)
	want := []hcl.Traversal{t2}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("filterPrefixTraversals(%v) => %v, want %v", list, got, want)
	}

	// Additional test for splat prefix
	t4 := hcl.Traversal{
		hcl.TraverseRoot{Name: "aws_instance"},
		hcl.TraverseAttr{Name: "multiple"},
		hcl.TraverseSplat{},
	}
	t5 := hcl.Traversal{
		hcl.TraverseRoot{Name: "aws_instance"},
		hcl.TraverseAttr{Name: "multiple"},
		hcl.TraverseSplat{},
		hcl.TraverseAttr{Name: "id"},
	}
	list2 := []hcl.Traversal{t4, t5}
	got2 := filterPrefixTraversals(list2)
	want2 := []hcl.Traversal{t5}
	if !reflect.DeepEqual(got2, want2) {
		t.Errorf("filterPrefixTraversals(%v) => %v, want %v", list2, got2, want2)
	}
}

func TestStepEqual(t *testing.T) {
	cases := []struct {
		Name   string
		StepA  hcl.Traverser
		StepB  hcl.Traverser
		Expect bool
	}{
		{
			Name:   "Root steps equal",
			StepA:  hcl.TraverseRoot{Name: "aws_instance"},
			StepB:  hcl.TraverseRoot{Name: "aws_instance"},
			Expect: true,
		},
		{
			Name:   "Root steps not equal",
			StepA:  hcl.TraverseRoot{Name: "aws_instance"},
			StepB:  hcl.TraverseRoot{Name: "aws_s3_bucket"},
			Expect: false,
		},
		{
			Name:   "Attribute steps equal",
			StepA:  hcl.TraverseAttr{Name: "id"},
			StepB:  hcl.TraverseAttr{Name: "id"},
			Expect: true,
		},
		{
			Name:   "Attribute steps not equal",
			StepA:  hcl.TraverseAttr{Name: "id"},
			StepB:  hcl.TraverseAttr{Name: "name"},
			Expect: false,
		},
		{
			Name:   "TraverseAttr with prefix",
			StepA:  hcl.TraverseAttr{Name: "example"},
			StepB:  hcl.TraverseAttr{Name: "example.id"},
			Expect: true,
		},
		{
			Name:   "TraverseAttr with splat",
			StepA:  hcl.TraverseAttr{Name: "multiple"},
			StepB:  hcl.TraverseAttr{Name: "multiple[*]"},
			Expect: true,
		},
		{
			Name:   "Splat steps equal",
			StepA:  hcl.TraverseSplat{},
			StepB:  hcl.TraverseSplat{},
			Expect: true,
		},
		{
			Name:   "Splat vs Attribute not equal",
			StepA:  hcl.TraverseSplat{},
			StepB:  hcl.TraverseAttr{Name: "id"},
			Expect: false,
		},
		{
			Name:   "Index steps with same key equal",
			StepA:  hcl.TraverseIndex{Key: cty.NumberIntVal(1)},
			StepB:  hcl.TraverseIndex{Key: cty.NumberIntVal(1)},
			Expect: true,
		},
		{
			Name:   "Index steps with different keys not equal",
			StepA:  hcl.TraverseIndex{Key: cty.NumberIntVal(1)},
			StepB:  hcl.TraverseIndex{Key: cty.NumberIntVal(2)},
			Expect: false,
		},
		{
			Name:   "Splat vs Splat with extra attribute => true",
			StepA:  hcl.TraverseSplat{},
			StepB:  hcl.TraverseSplat{},
			Expect: true,
		},
		{
			Name:   "Splat vs Attribute => false",
			StepA:  hcl.TraverseSplat{},
			StepB:  hcl.TraverseAttr{Name: "id"},
			Expect: false,
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			got := stepEqual(c.StepA, c.StepB)
			if got != c.Expect {
				t.Errorf("stepEqual(%v, %v) = %v; want %v", c.StepA, c.StepB, got, c.Expect)
			}
		})
	}
}

func TestIsFullResourceReference(t *testing.T) {
	rule := NewTerraformOutputResourceRule()

	cases := []struct {
		Name     string
		Trav     hcl.Traversal
		Expected bool
	}{
		{
			Name: "Single step (aws_instance) => false (incomplete)",
			Trav: hcl.Traversal{
				hcl.TraverseRoot{Name: "aws_instance"},
			},
			Expected: false,
		},
		{
			Name: "aws_instance.example => entire resource",
			Trav: hcl.Traversal{
				hcl.TraverseRoot{Name: "aws_instance"},
				hcl.TraverseAttr{Name: "example"},
			},
			Expected: true,
		},
		{
			Name: "aws_instance.example.id => partial attribute",
			Trav: hcl.Traversal{
				hcl.TraverseRoot{Name: "aws_instance"},
				hcl.TraverseAttr{Name: "example"},
				hcl.TraverseAttr{Name: "id"},
			},
			Expected: false,
		},
		{
			Name: "aws_instance.example[0] => entire resource (indexed)",
			Trav: hcl.Traversal{
				hcl.TraverseRoot{Name: "aws_instance"},
				hcl.TraverseAttr{Name: "example"},
				hcl.TraverseIndex{Key: cty.NumberIntVal(0)},
			},
			Expected: true,
		},