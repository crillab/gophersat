package solver

import (
	"os"
	"testing"
)

func TestPropClause(t *testing.T) {
	clauses := []PBConstr{
		PropClause(1, 2, 3),
		PropClause(-1, -2),
		PropClause(-2, -3),
		PropClause(-1, -3),
		PropClause(2),
	}
	pb := ParsePBConstrs(clauses)
	s := New(pb)
	status := s.Solve()
	if status != Sat {
		t.Errorf("problem should be sat")
	} else {
		model := s.Model()
		if model[IntToVar(1)] || !model[IntToVar(2)] || model[IntToVar(3)] {
			t.Errorf("invalid model: got %v", model)
		}
	}
}

func TestAtMostAtLeast(t *testing.T) {
	clauses := []PBConstr{
		PropClause(1, 2, 3),
		AtMost([]int{1, 2, 3}, 1),
		AtLeast([]int{-1, -3}, 2),
	}
	pb := ParsePBConstrs(clauses)
	s := New(pb)
	status := s.Solve()
	if status != Sat {
		t.Errorf("problem should be sat")
	} else {
		model := s.Model()
		if model[IntToVar(1)] || !model[IntToVar(2)] || model[IntToVar(3)] {
			t.Errorf("invalid model: got %v", model)
		}
	}
}

func TestLtEq(t *testing.T) {
	pc := LtEq([]int{1, 2, 3, 4}, []int{2, 1, 1, 1}, 3)
	if pc.Lits[0] != -1 || pc.Lits[1] != -2 || pc.Lits[2] != -3 || pc.Lits[3] != -4 {
		t.Errorf("incorrect literals: %v", pc.Lits)
	}
	if pc.AtLeast != 2 {
		t.Errorf("incorrect cardinality: %d", pc.AtLeast)
	}
}

func TestLtEqZeroWtFirst(t *testing.T) {
	pc := LtEq([]int{2, 1, 3, 4}, []int{0, 2, 1, 1}, 3)
	if pc.Weights[0] != 2 || pc.Weights[1] != 1 || pc.Weights[2] != 1 {
		t.Errorf("incorrect weights: %v", pc.Weights)
	}
	if pc.Lits[0] != -1 || pc.Lits[1] != -3 || pc.Lits[2] != -4 {
		t.Errorf("incorrect literals: %v", pc.Lits)
	}
	if pc.AtLeast != 1 {
		t.Errorf("incorrect cardinality: %d", pc.AtLeast)
	}
}

func TestLtEqZeroWtLoast(t *testing.T) {
	pc := LtEq([]int{4, 1, 3, 2}, []int{1, 2, 1, 0}, 3)
	if pc.Weights[0] != 1 || pc.Weights[1] != 2 || pc.Weights[2] != 1 {
		t.Errorf("incorrect weights: %v", pc.Weights)
	}
	if pc.Lits[0] != -4 || pc.Lits[1] != -1 || pc.Lits[2] != -3 {
		t.Errorf("incorrect literals: %v", pc.Lits)
	}
	if pc.AtLeast != 1 {
		t.Errorf("incorrect cardinality: %d", pc.AtLeast)
	}
}

func TestLtEqZeroWt(t *testing.T) {
	pc := LtEq([]int{1, 2, 3, 4}, []int{2, 0, 1, 1}, 3)
	if pc.Weights[0] != 2 || pc.Weights[1] != 1 || pc.Weights[2] != 1 {
		t.Errorf("incorrect weights: %v", pc.Weights)
	}
	if pc.Lits[0] != -1 || pc.Lits[1] != -3 || pc.Lits[2] != -4 {
		t.Errorf("incorrect literals: %v", pc.Lits)
	}
	if pc.AtLeast != 1 {
		t.Errorf("incorrect cardinality: %d", pc.AtLeast)
	}
}

func TestEq(t *testing.T) {
	pc := Eq([]int{1, 2, 3}, []int{2, 1, 1}, 2)
	if len(pc) != 2 {
		t.Errorf("expected 2 constraints, got %d: %v", len(pc), pc)
	} else {
		c0 := pc[0]
		c1 := pc[1]
		if c0.Lits[0] != 1 || c0.Lits[1] != 2 || c0.Lits[2] != 3 {
			t.Errorf("invalid lits in first constr: %v", c0)
		}
		if c1.Lits[0] != -1 || c1.Lits[1] != -2 || c1.Lits[2] != -3 {
			t.Errorf("invalid lits in second constr: %v", c1)
		}
	}
}

func TestPBPigeons(t *testing.T) {
	clauses := []PBConstr{
		PropClause(1, 2, 3),
		PropClause(4, 5, 6),
		PropClause(7, 8, 9),
		PropClause(10, 11, 12),
		LtEq([]int{1, 4, 7, 10}, []int{1, 1, 1, 1}, 1),
		LtEq([]int{2, 5, 8, 11}, []int{1, 1, 1, 1}, 1),
		LtEq([]int{3, 6, 9, 12}, []int{1, 1, 1, 1}, 1),
	}
	pb := ParsePBConstrs(clauses)
	s := New(pb)
	status := s.Solve()
	if status != Unsat {
		t.Errorf("problem should be unsat")
	}
}

func TestPBTrivial(t *testing.T) {
	// Can be solved by unit propagation
	clauses := []PBConstr{
		GtEq([]int{1, 2, 3}, []int{2, 2, 1}, 4),
		GtEq([]int{3, 4, 5}, []int{2, 1, 1}, 3),
		GtEq([]int{-1, -2, -3}, []int{2, 2, 2}, 5),
	}
	pb := ParsePBConstrs(clauses)
	s := New(pb)
	status := s.Solve()
	if status != Unsat {
		t.Errorf("problem should be unsat")
	}
}

func TestPB(t *testing.T) {
	clauses := []PBConstr{
		GtEq([]int{1, 2, 3}, []int{2, 1, 1}, 2),
		GtEq([]int{-1, -2, -3, -4}, []int{3, 2, 2, 1}, 3),
		GtEq([]int{2, 3, 4}, []int{2, 2, 1}, 3),
	}
	pb := ParsePBConstrs(clauses)
	s := New(pb)
	status := s.Solve()
	if status != Sat {
		t.Errorf("problem should be sat")
	}
	clauses = []PBConstr{
		GtEq([]int{1, 2, 3, 4}, []int{3, 2, 1, 1}, 3),
		GtEq([]int{-1, -2, -3}, []int{2, 1, 1}, 2),
		PropClause(-1, -4),
		AtMost([]int{2, 3, 4}, 1),
	}
	pb = ParsePBConstrs(clauses)
	s = New(pb)
	status = s.Solve()
	if status != Sat {
		t.Errorf("problem should be sat")
	}
	clauses = []PBConstr{
		GtEq([]int{1, 2, 3}, []int{2, 1, 1}, 2),
		GtEq([]int{-1, -2, -3}, []int{2, 2, 1}, 3),
	}
	pb = ParsePBConstrs(clauses)
	s = New(pb)
	status = s.Solve()
	if status != Sat {
		t.Errorf("problem should be sat")
	}
}

func runPBBench(path string, b *testing.B) {
	f, err := os.Open(path)
	if err != nil {
		b.Fatal(err.Error())
	}
	defer func() { _ = f.Close() }()
	for i := 0; i < b.N; i++ {
		pb, err := ParseOPB(f)
		if err != nil {
			b.Fatal(err.Error())
		}
		s := New(pb)
		s.Solve()
	}
}

func BenchmarkSimple(b *testing.B) {
	runPBBench("testcnf/simple.opb", b)
}

func Benchmark3Col(b *testing.B) {
	runPBBench("testcnf/3col-almost3reg-l010-r009-n1.opb", b)
}

func BenchmarkBandwidth(b *testing.B) {
	runPBBench("testcnf/fixed-bandwidth-10.cnf.gz-extracted.pb", b)
}
