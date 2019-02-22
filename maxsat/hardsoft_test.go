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
