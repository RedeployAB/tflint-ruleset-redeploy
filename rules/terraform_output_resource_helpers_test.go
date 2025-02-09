package rules

import (
	"reflect"
	"testing"

	"github.com/hashicorp/hcl/v2"
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
			name: "Attributes only",
			input: hcl.Traversal{
				hcl.TraverseRoot{Name: "aws_instance"},
				hcl.TraverseAttr{Name: "name"},
			},
			want: hcl.Traversal{
				hcl.TraverseRoot{Name: "aws_instance"},
				hcl.TraverseAttr{Name: "name"},
			},
		},
		{
			name: "Index with attribute",
			input: hcl.Traversal{
				hcl.TraverseRoot{Name: "aws_instance"},
				hcl.TraverseAttr{Name: "instances[0].id"},
			},
			want: hcl.Traversal{
				hcl.TraverseRoot{Name: "aws_instance"},
				hcl.TraverseAttr{Name: "instances"},
				hcl.TraverseIndex{Key: cty.NumberIntVal(0)},
				hcl.TraverseAttr{Name: "id"},
			},
		},
	}

	for _, c := range cases {
		got := canonicalizeTraversal(c.input)
		if !traversalsEqual(got, c.want) {
			t.Errorf("canonicalizeTraversal(%v) = %v; want %v", c.input, got, c.want)
		}
	}
}

func TestFilterPrefixTraversals(t *testing.T) {
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
	want := []hcl.Traversal{t2}
	got := filterPrefixTraversals(list)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("filterPrefixTraversals(%v) = %v; want %v", list, got, want)
	}
}

func TestStepEqual(t *testing.T) {
	cases := []struct {
		name   string
		stepA  hcl.Traverser
		stepB  hcl.Traverser
		expect bool
	}{
		{
			name:   "Same TraverseRoot",
			stepA:  hcl.TraverseRoot{Name: "aws_instance"},
			stepB:  hcl.TraverseRoot{Name: "aws_instance"},
			expect: true,
		},
		{
			name:   "Different TraverseRoot",
			stepA:  hcl.TraverseRoot{Name: "aws_instance"},
			stepB:  hcl.TraverseRoot{Name: "aws_s3_bucket"},
			expect: false,
		},
		{
			name:   "Same TraverseAttr",
			stepA:  hcl.TraverseAttr{Name: "id"},
			stepB:  hcl.TraverseAttr{Name: "id"},
			expect: true,
		},
		{
			name:   "Different TraverseAttr",
			stepA:  hcl.TraverseAttr{Name: "id"},
			stepB:  hcl.TraverseAttr{Name: "name"},
			expect: false,
		},
		{
			name:   "TraverseSplat match",
			stepA:  hcl.TraverseSplat{},
			stepB:  hcl.TraverseSplat{},
			expect: true,
		},
		{
			name:   "TraverseSplat vs TraverseAttr",
			stepA:  hcl.TraverseSplat{},
			stepB:  hcl.TraverseAttr{Name: "id"},
			expect: false,
		},
	}

	for _, c := range cases {
		got := stepEqual(c.stepA, c.stepB)
		if got != c.expect {
			t.Errorf("%s: stepEqual(%v, %v) = %v; want %v", c.name, c.stepA, c.stepB, got, c.expect)
		}
	}
}

// traversalsEqual compares two hcl.Traversal values step-by-step using stepEqual.
func traversalsEqual(a, b hcl.Traversal) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !stepEqual(a[i], b[i]) {
			return false
		}
	}
	return true
}

func TestIsFullResourceReference(t *testing.T) {
	rule := NewTerraformOutputResourceRule()

	cases := []struct {
		name     string
		trav     hcl.Traversal
		expected bool
	}{
		{
			name: "Single root",
			trav: hcl.Traversal{
				hcl.TraverseRoot{Name: "aws_instance"},
			},
			expected: false,
		},
		{
			name: "Resource without attribute",
			trav: hcl.Traversal{
				hcl.TraverseRoot{Name: "aws_instance"},
				hcl.TraverseAttr{Name: "example"},
			},
			expected: true,
		},
		{
			name: "Resource with attribute",
			trav: hcl.Traversal{
				hcl.TraverseRoot{Name: "aws_instance"},
				hcl.TraverseAttr{Name: "example"},
				hcl.TraverseAttr{Name: "id"},
			},
			expected: false,
		},
		{
			name: "Data resource without attribute",
			trav: hcl.Traversal{
				hcl.TraverseRoot{Name: "data"},
				hcl.TraverseAttr{Name: "aws_caller_identity"},
				hcl.TraverseAttr{Name: "current"},
			},
			expected: true,
		},
		{
			name: "Data resource with attribute",
			trav: hcl.Traversal{
				hcl.TraverseRoot{Name: "data"},
				hcl.TraverseAttr{Name: "aws_caller_identity"},
				hcl.TraverseAttr{Name: "current"},
				hcl.TraverseAttr{Name: "account_id"},
			},
			expected: false,
		},
		{
			name: "Variable reference",
			trav: hcl.Traversal{
				hcl.TraverseRoot{Name: "var"},
				hcl.TraverseAttr{Name: "example"},
			},
			expected: false,
		},
		{
			name: "Local reference",
			trav: hcl.Traversal{
				hcl.TraverseRoot{Name: "local"},
				hcl.TraverseAttr{Name: "example"},
			},
			expected: false,
		},
		{
			name: "Resource with index",
			trav: hcl.Traversal{
				hcl.TraverseRoot{Name: "aws_instance"},
				hcl.TraverseAttr{Name: "example"},
				hcl.TraverseIndex{Key: cty.NumberIntVal(0)},
			},
			expected: true,
		},
		{
			name: "Resource with splat",
			trav: hcl.Traversal{
				hcl.TraverseRoot{Name: "aws_instance"},
				hcl.TraverseAttr{Name: "example"},
				hcl.TraverseSplat{},
			},
			expected: true,
		},
	}

	for _, c := range cases {
		got := rule.isFullResourceReference(c.trav)
		if got != c.expected {
			t.Errorf("%s: isFullResourceReference(%v) = %v; want %v", c.name, c.trav, got, c.expected)
		}
	}
}
