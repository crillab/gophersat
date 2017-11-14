package bf

import (
	"fmt"
	"strings"
	"testing"
)

// To each formula, associate an expected string input.
// An empty string means an error is expected.
var exprToFormula = map[string]string{
	"foo":                  "foo",
	"^foo":                 "not(foo)",
	"^^foo":                "not(not(foo))",
	"(foo)":                "foo",
	"a | b":                "or(a, b)",
	"a & b":                "and(a, b)",
	"a -> b":               "or(not(a), b)",
	"a = b":                "and(or(not(a), b), or(a, not(b)))",
	"^(a|  b)":             "not(or(a, b))",
	"a & b & c":            "and(a, and(b, c))",
	"a & (b & c) & d":      "and(a, and(and(b, c), d))",
	"a = b |c -> ^(d&e)":   "and(or(not(a), or(not(or(b, c)), not(and(d, e)))), or(a, not(or(not(or(b, c)), not(and(d, e))))))",
	"(a|^b|c) & ^(a|^b|c)": "and(or(a, or(not(b), c)), not(or(a, or(not(b), c))))",
	"{a, b, c}":            "and(or(a, b, c), or(not(a), not(b)), or(not(a), not(c)), or(not(b), not(c)))",
	"a | b; ^a | ^b":       "and(or(a, b), or(not(a), not(b)))",
}

func TestParse(t *testing.T) {
	for expr, expected := range exprToFormula {
		r := strings.NewReader(expr)
		f, err := Parse(r)
		if err != nil {
			t.Errorf("Could not parse expression %q: %v", expr, err)
		} else if f.String() != expected {
			t.Errorf("For expression %q, expected formula %q, got %q", expr, expected, f.String())
		}
	}
}

func ExampleParse() {
	expr := "a & ^(b -> c) & (c = d | ^a)"
	f, err := Parse(strings.NewReader(expr))
	if err != nil {
		fmt.Printf("Could not parse expression %q: %v", expr, err)
	} else {
		sat, model, err := Solve(f)
		if err != nil {
			fmt.Printf("Could not solve %q: %v", expr, err)
		} else if !sat {
			fmt.Printf("Problem is unsatisfiable")
		} else {
			fmt.Printf("Problem is satisfiable, model: a=%t, b=%t, c=%t, d=%t", model["a"], model["b"], model["c"], model["d"])
		}
	}
	// Output:
	// Problem is satisfiable, model: a=true, b=true, c=false, d=false
}

func ExampleParse_unsatisfiable() {
	expr := "(a|^b|c) & ^(a|^b|c)"
	f, err := Parse(strings.NewReader(expr))
	if err != nil {
		fmt.Printf("Could not parse expression %q: %v", expr, err)
	} else {
		sat, model, err := Solve(f)
		if err != nil {
			fmt.Printf("Could not solve %q: %v", expr, err)
		} else if !sat {
			fmt.Printf("Problem is unsatisfiable")
		} else {
			fmt.Printf("Problem is satisfiable, model: a=%t, b=%t, c=%t", model["a"], model["b"], model["c"])
		}
	}
	// Output:
	// Problem is unsatisfiable
}

func ExampleParse_unique() {
	expr := "a & {a, b, c}"
	f, err := Parse(strings.NewReader(expr))
	if err != nil {
		fmt.Printf("Could not parse expression %q: %v", expr, err)
	} else {
		sat, model, err := Solve(f)
		if err != nil {
			fmt.Printf("Could not solve %q: %v", expr, err)
		} else if !sat {
			fmt.Printf("Problem is unsatisfiable")
		} else {
			fmt.Printf("Problem is satisfiable, model: a=%t, b=%t, c=%t", model["a"], model["b"], model["c"])
		}
	}
	// Output:
	// Problem is satisfiable, model: a=true, b=false, c=false
}
