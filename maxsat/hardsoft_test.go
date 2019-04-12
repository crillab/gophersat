package maxsat

import (
	"fmt"
	"testing"
)

func TestHardSoft(t *testing.T) {
	x := Var("x")
	//y := Var("y")
	hard := HardClause(x)
	soft := SoftClause(x.Negation())

	problem := New(hard, soft)
	problem.Solve()
}

func TestDoubleUnits(t *testing.T) {
	x := Var("x")
	//y := maxsat.Var("y")
	hard := HardClause(x)
	hard2 := HardClause(x)

	problem := New(hard, hard2)
	problem.Solve()
}

func TestBugInfiniteLoop(t *testing.T) {
	cs := Var("cs")
	p := Var("p")
	d1 := Var("d1")
	d2 := Var("d2")
	t1 := Var("t1")
	t2 := Var("t2")
	c1 := Var("c1")
	c2 := Var("c2")

	clauses := []Constr{
		HardClause(cs),
		HardClause(cs.Negation(), p),

		SoftClause(p.Negation(), d1, d2),
		SoftClause(d1.Negation(), t1),
		SoftClause(d1.Negation(), t2),
		SoftClause(d2.Negation()),
		SoftClause(t1.Negation(), c1),
		SoftClause(t2.Negation(), c2),

		HardClause(d1.Negation(), d2.Negation()),
		HardClause(c1.Negation(), c2.Negation()),
	}
	problem := New(clauses...)

	model, cost := problem.Solve()
	fmt.Println(cost)
	fmt.Println(model)

}
