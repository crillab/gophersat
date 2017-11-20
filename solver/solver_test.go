package solver

import (
	"fmt"
	"os"
	"testing"
)

// A test associates a path with an expected output.
type test struct {
	path     string
	expected Status
}

func runTest(test test, t *testing.T) {
	f, err := os.Open(test.path)
	if err != nil {
		t.Error(err.Error())
	}
	defer func() { _ = f.Close() }()
	pb, err := ParseCNF(f)
	if err != nil {
		t.Error(err.Error())
	}
	s := New(pb)
	if status := s.Solve(); status != test.expected {
		t.Errorf("Invalid result for %q: expected %v, got %v", test.path, test.expected, status)
	}
}

var tests = []test{
	{"testcnf/25.cnf", Sat},
	{"testcnf/50.cnf", Sat},
	{"testcnf/75.cnf", Sat},
	{"testcnf/100.cnf", Sat},
	{"testcnf/125.cnf", Unsat},
	{"testcnf/150.cnf", Unsat},
	{"testcnf/175.cnf", Unsat},
	{"testcnf/200.cnf", Unsat},
	{"testcnf/225.cnf", Sat},
	{"testcnf/250.cnf", Unsat},
	{"testcnf/hoons-vbmc-lucky7.cnf", Unsat},
}

func TestSolver(t *testing.T) {
	for _, test := range tests {
		runTest(test, t)
	}
}

func runBench(path string, b *testing.B) {
	f, err := os.Open(path)
	if err != nil {
		b.Fatal(err.Error())
	}
	defer func() { _ = f.Close() }()
	for i := 0; i < b.N; i++ {
		pb, err := ParseCNF(f)
		if err != nil {
			b.Fatal(err.Error())
		}
		s := New(pb)
		s.Solve()
	}
}

func TestParseSlice(t *testing.T) {
	cnf := [][]int{{1, 2, 3}, {-1}, {-2}, {-3}}
	pb := ParseSlice(cnf)
	s := New(pb)
	if status := s.Solve(); status != Unsat {
		t.Fatalf("expected unsat for problem %v, got %v", cnf, status)
	}
}

func TestParseSliceSat(t *testing.T) {
	cnf := [][]int{{1}, {-2, 3}, {-2, 4}, {-5, 3}, {-5, 6}, {-7, 3}, {-7, 8}, {-9, 10}, {-9, 4}, {-1, 10}, {-1, 6}, {3, 10}, {-3, -10}, {4, 6, 8}}
	pb := ParseSlice(cnf)
	s := New(pb)
	if status := s.Solve(); status != Sat {
		t.Fatalf("expected sat for problem %v, got %v", cnf, status)
	}
}

func TestParseSliceTrivial(t *testing.T) {
	cnf := [][]int{{1}, {-1}}
	pb := ParseSlice(cnf)
	s := New(pb)
	if status := s.Solve(); status != Unsat {
		t.Fatalf("expected unsat for problem %v, got %v", cnf, status)
	}
}

func TestParseCardConstrs(t *testing.T) {
	clauses := []CardConstr{
		{Lits: []int{1, 2, 3}, AtLeast: 3},
		{Lits: []int{-1, -2}, AtLeast: 0},
		{Lits: []int{2, 3, -4}, AtLeast: 2},
		AtLeast1(-1, -4),
	}
	pb := ParseCardConstrs(clauses)
	s := New(pb)
	if status := s.Solve(); status != Sat {
		t.Fatalf("expected sat for cardinality problem %v, got %v", clauses, status)
	}
	model, err := s.Model()
	if err != nil {
		t.Fatalf("could not get model: %v", err)
	}
	if !model[IntToVar(1)] || !model[IntToVar(2)] || !model[IntToVar(3)] || model[IntToVar(4)] {
		t.Fatalf("expected model 1, 2, 3, -4, got: %v", model)
	}
}

func TestAtMost1(t *testing.T) {
	c := AtMost1(1, -2, 3)
	if c.Lits[0] != -1 || c.Lits[1] != 2 || c.Lits[2] != -3 {
		t.Errorf("invalid constraint: expected [[-1 2 -3] 2], got %v", c)
	}
	if c.AtLeast != 2 {
		t.Errorf("invalid cardinality: expected 2, got %d", c.AtLeast)
	}
}

func TestParseCardConstrsTrivial(t *testing.T) {
	pb := ParseCardConstrs([]CardConstr{{Lits: []int{1, 2}, AtLeast: 3}})
	s := New(pb)
	if status := s.Solve(); status != Unsat {
		t.Errorf("expected unsat, got %v", status)
	}
	pb = ParseCardConstrs([]CardConstr{{Lits: []int{1, 2, 3}, AtLeast: 3}})
	s = New(pb)
	if status := s.Solve(); status != Sat {
		t.Errorf("expected sat, got %v", status)
	} else if model, err := s.Model(); err != nil {
		t.Errorf("could not get model: %v", err)
	} else if !model[IntToVar(1)] || !model[IntToVar(2)] || !model[IntToVar(3)] {
		t.Errorf("invalid model, expected all true bindings, got %v", model)
	}
	pb = ParseCardConstrs([]CardConstr{{Lits: []int{1, -2, 3}, AtLeast: 2}, AtLeast1(2)})
	s = New(pb)
	if status := s.Solve(); status != Sat {
		t.Errorf("expected sat, got %v", status)
	} else if model, err := s.Model(); err != nil {
		t.Errorf("could not get model: %v", err)
	} else if !model[IntToVar(1)] || !model[IntToVar(2)] || !model[IntToVar(3)] {
		t.Errorf("invalid model, expected all true bindings, got %v", model)
	}
}

func TestPigeonCard(t *testing.T) {
	pb := ParseCardConstrs([]CardConstr{
		AtLeast1(1, 2, 3),
		AtMost1(1, 2, 3),
		AtLeast1(4, 5, 6),
		AtMost1(4, 5, 6),
		AtLeast1(7, 8, 9),
		AtMost1(7, 8, 9),
		AtLeast1(10, 11, 12),
		AtMost1(10, 11, 12),
		AtMost1(1, 4, 7, 10),
		AtMost1(2, 5, 8, 11),
		AtMost1(3, 6, 9, 12),
	})
	s := New(pb)
	if status := s.Solve(); status == Sat {
		model, _ := s.Model()
		t.Errorf("model found for pigeon problem: %v", model)
	}
}

func ExampleParseCardConstrs() {
	clauses := []CardConstr{
		{Lits: []int{1, 2, 3}, AtLeast: 3},
		{Lits: []int{2, 3, -4}, AtLeast: 2},
		AtLeast1(-1, -4),
		AtLeast1(-2, -3, 4),
	}
	pb := ParseCardConstrs(clauses)
	s := New(pb)
	if status := s.Solve(); status == Unsat {
		fmt.Println("Problem is not satisfiable")
	} else {
		model, err := s.Model()
		if err != nil {
			fmt.Printf("Could not get model: %v", err)
		} else {
			fmt.Printf("Model found: %v\n", model)
		}
	}
	// Output:
	// Problem is not satisfiable
}

func BenchmarkSolver125(b *testing.B) {
	runBench("testcnf/125.cnf", b)
}

func BenchmarkSolver150(b *testing.B) {
	runBench("testcnf/150.cnf", b)
}

func BenchmarkSolver175(b *testing.B) {
	runBench("testcnf/175.cnf", b)
}

func BenchmarkSolver200(b *testing.B) {
	runBench("testcnf/200.cnf", b)
}

func BenchmarkSolver225(b *testing.B) {
	runBench("testcnf/225.cnf", b)
}

func BenchmarkSolver250(b *testing.B) {
	runBench("testcnf/250.cnf", b)
}

func BenchmarkSolver275(b *testing.B) {
	runBench("testcnf/275.cnf", b)
}

func BenchmarkSolver300(b *testing.B) {
	runBench("testcnf/300.cnf", b)
}

func BenchmarkSolverIndus(b *testing.B) {
	runBench("testcnf/hoons-vbmc-lucky7.cnf", b)
}
