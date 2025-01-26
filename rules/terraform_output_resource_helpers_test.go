package rules

import (
	"reflect"
	"strings"
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
			Input:  "name_with[brackets]",
			Expect: []string{"name_with[brackets]"},
		},
		{
			Input:  "name.with.dots[and][brackets]",
			Expect: []string{"name", "with", "dots", "[and]", "[brackets]"},
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
			Name:   "Splat vs Splat with extra attribute => false",
			StepA:  hcl.TraverseAttr{Name: "multiple[*]"},
			StepB:  hcl.TraverseAttr{Name: "multiple[*].id"},
			Expect: false,
		},
		{
			Name:   "Plain splat steps => true",
			StepA:  hcl.TraverseSplat{},
			StepB:  hcl.TraverseSplat{},
			Expect: true,
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

// Helper functions for testing purposes

func splitAttrName(name string) []string {
	var parts []string
	var buf strings.Builder

	for i := 0; i < len(name); i++ {
		c := name[i]

		switch c {
		case '.':
			if buf.Len() > 0 {
				parts = append(parts, buf.String())
				buf.Reset()
			}
		case '[':
			if buf.Len() > 0 {
				parts = append(parts, buf.String())
				buf.Reset()
			}
			j := i + 1
			for j < len(name) && name[j] != ']' {
				j++
			}
			if j < len(name) && name[j] == ']' {
				parts = append(parts, name[i:j+1])
				i = j
			} else {
				buf.WriteByte(c)
			}
		default:
			buf.WriteByte(c)
		}
	}

	if buf.Len() > 0 {
		parts = append(parts, buf.String())
	}
	return parts
}

func canonicalizeTraversal(trav hcl.Traversal) hcl.Traversal {
	var result hcl.Traversal

	for _, step := range trav {
		switch s := step.(type) {
		case hcl.TraverseAttr:
			subSteps := splitAttrName(s.Name)
			for _, sub := range subSteps {
				if sub == "[*]" {
					result = append(result, hcl.TraverseSplat{})
				} else if strings.HasPrefix(sub, "[") && strings.HasSuffix(sub, "]") {
					indexKey := strings.Trim(sub, "[]")
					result = append(result, makeIndexStep(indexKey))
				} else {
					result = append(result, hcl.TraverseAttr{Name: sub})
				}
			}
		default:
			result = append(result, step)
		}
	}

	return result
}

func filterPrefixTraversals(all []hcl.Traversal) []hcl.Traversal {
	var result []hcl.Traversal

outer:
	for i, t1 := range all {
		for j, t2 := range all {
			if i == j {
				continue
			}
			if isPrefix(t1, t2) {
				continue outer
			}
		}
		result = append(result, t1)
	}
	return result
}

func isPrefix(t1, t2 hcl.Traversal) bool {
	if len(t1) >= len(t2) {
		return false
	}
	for i := range t1 {
		if !stepEqual(t1[i], t2[i]) {
			return false
		}
	}
	return true
}

func stepEqual(a, b hcl.Traverser) bool {
	switch aTyped := a.(type) {
	case hcl.TraverseRoot:
		if bTyped, ok := b.(hcl.TraverseRoot); ok {
			return aTyped.Name == bTyped.Name