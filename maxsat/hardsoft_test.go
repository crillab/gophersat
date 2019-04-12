package maxsat

import (
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

func TestSecondBug(t *testing.T) {
	x := Var("x")
	//y := maxsat.Var("y")
	hard := HardClause(x)
	hard2 := HardClause(x)

	problem := New(hard, hard2)
	problem.Solve()
}
