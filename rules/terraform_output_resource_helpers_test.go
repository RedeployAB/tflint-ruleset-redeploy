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
			Input:  "one[0].two",
			Expect: []string{"one", "[0]", "two"},
		},
		{
			Input:  "no.brackets.or.index",
			Expect: []string{"no", "brackets", "or", "index"},
		},
	}

	for _, c := range cases {
		actual := splitAttrName(c.Input)
		if !reflect.DeepEqual(actual, c.Expect) {
			t.Errorf("splitAttrName(%q) => %v, want %v", c.Input, actual, c.Expect)
		}
	}
}

func TestMakeIndexStep(t *testing.T) {
	cases := []struct {
		Input       string
		ExpectSplat bool
		ExpectIndex interface{} // can be an int or string
	}{
		{Input: "*", ExpectSplat: true},
		{Input: "0", ExpectSplat: false, ExpectIndex: 0},
		{Input: "123", ExpectSplat: false, ExpectIndex: 123},
		{Input: "abc", ExpectSplat: false, ExpectIndex: "abc"},
	}

	for _, c := range cases {
		step := makeIndexStep(c.Input)
		switch s := step.(type) {
		case hcl.TraverseSplat:
			if !c.ExpectSplat {
				t.Errorf("makeIndexStep(%q) => got splat, want index %v", c.Input, c.ExpectIndex)
			}
		case hcl.TraverseIndex:
			keyVal := s.Key
			switch {
			case keyVal.Type() == cty.Number:
				num, _ := keyVal.AsBigFloat().Int64()
				expectedNum, ok := c.ExpectIndex.(int)
				if !ok || int64(expectedNum) != num {
					t.Errorf("makeIndexStep(%q) => numeric index %d, want %v", c.Input, num, c.ExpectIndex)
				}
			case keyVal.Type() == cty.String:
				str := keyVal.AsString()
				expectedStr, ok := c.ExpectIndex.(string)
				if !ok || str != expectedStr {
					t.Errorf("makeIndexStep(%q) => string index %q, want %q", c.Input, str, c.ExpectIndex)
				}
			default:
				t.Errorf("makeIndexStep(%q) => unknown key type %s", c.Input, keyVal.Type().FriendlyName())
			}
		default:
			t.Errorf("makeIndexStep(%q) => unexpected type %T", c.Input, s)
		}
	}
}

func TestCanonicalizeTraversal(t *testing.T) {
	input := hcl.Traversal{
		hcl.TraverseRoot{Name: "aws_instance"},
		hcl.TraverseAttr{Name: "multiple[*].id"},
	}
	got := canonicalizeTraversal(input)
	want := hcl.Traversal{
		hcl.TraverseRoot{Name: "aws_instance"},
		hcl.TraverseAttr{Name: "multiple"},
		hcl.TraverseSplat{},
		hcl.TraverseAttr{Name: "id"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("canonicalizeTraversal(%v) => %v, want %v", input, got, want)
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
	t3 := hcl.Traversal{
		hcl.TraverseRoot{Name: "aws_instance"},
		hcl.TraverseAttr{Name: "example"},
		hcl.TraverseAttr{Name: "name"},
	}
	list := []hcl.Traversal{t1, t2, t3}
	got := filterPrefixTraversals(list)
	want := []hcl.Traversal{t2, t3} // t1 should be filtered out
	if !reflect.DeepEqual(got, want) {
		t.Errorf("filterPrefixTraversals(%v) => %v, want %v", list, got, want)
	}
}

func TestIsResourceRootTraversal(t *testing.T) {
	tr := hcl.Traversal{
		hcl.TraverseRoot{Name: "data"},
		hcl.TraverseAttr{Name: "aws_caller_identity"},
	}
	if !isResourceRootTraversal(tr) {
		t.Errorf("Expected isResourceRootTraversal to be true for %v", tr)
	}

	trVar := hcl.Traversal{
		hcl.TraverseRoot{Name: "var"},
		hcl.TraverseAttr{Name: "foo"},
	}
	if isResourceRootTraversal(trVar) {
		t.Errorf("Expected isResourceRootTraversal to be false for var %v", trVar)
	}
}

func TestEndsWithAttribute(t *testing.T) {
	tr := hcl.Traversal{
		hcl.TraverseRoot{Name: "aws_instance"},
		hcl.TraverseAttr{Name: "example"},
		hcl.TraverseAttr{Name: "id"},
	}
	if !endsWithAttribute(tr) {
		t.Errorf("endsWithAttribute => false, want true for %v", tr)
	}

	tr2 := canonicalizeTraversal(hcl.Traversal{
		hcl.TraverseRoot{Name: "aws_instance"},
		hcl.TraverseAttr{Name: "example[*].id"},
	})
	if !endsWithAttribute(tr2) {
		t.Errorf("endsWithAttribute => false, want true for %v", tr2)
	}
}
