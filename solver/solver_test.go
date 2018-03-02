package solver

import (
	"fmt"
	"os"
	"strings"
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
		return
	}
	defer func() { _ = f.Close() }()
	var pb *Problem
	if strings.HasSuffix(test.path, "cnf") {
		pb, err = ParseCNF(f)
	} else {
		pb, err = ParseOPB(f)
	}
	if err != nil {
		t.Error(err.Error())
		return
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
	{"testcnf/hoons-vbmc-lucky7.cnf", Unsat},
	{"testcnf/3col-almost3reg-l010-r009-n1.opb", Unsat},
	{"testcnf/simple.opb", Sat},
	{"testcnf/fixed-bandwidth-10.cnf.gz-extracted.pb", Unsat},
	{"testcnf/ex1.opb", Unsat},
	{"testcnf/lo_8x8_009.opb", Sat},
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
	model := s.Model()
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
	} else {
		model := s.Model()
		if !model[IntToVar(1)] || !model[IntToVar(2)] || !model[IntToVar(3)] {
			t.Errorf("invalid model, expected all true bindings, got %v", model)
		}
	}
	pb = ParseCardConstrs([]CardConstr{{Lits: []int{1, -2, 3}, AtLeast: 2}, AtLeast1(2)})
	s = New(pb)
	if status := s.Solve(); status != Sat {
		t.Errorf("expected sat, got %v", status)
	} else {
		model := s.Model()
		if !model[IntToVar(1)] || !model[IntToVar(2)] || !model[IntToVar(3)] {
			t.Errorf("invalid model, expected all true bindings, got %v", model)
		}
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
		model := s.Model()
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
		fmt.Printf("Model found: %v\n", s.Model())
	}
	// Output:
	// Problem is not satisfiable
}

func TestCountModel(t *testing.T) {
	clauses := []CardConstr{
		AtLeast1(1, 2, 3),
		AtLeast1(-1, -2, -3),
		AtLeast1(2, 3, 4),
		AtLeast1(2, 3, 5),
		AtLeast1(3, 4, 5),
		AtLeast1(2, 4, 5),
	}
	pb := ParseCardConstrs(clauses)
	s := New(pb)
	if nb := s.CountModels(); nb != 17 {
		t.Errorf("Invalid #models: expected %d, got %d", 17, nb)
	}
}

func TestEnumerate(t *testing.T) {
	clauses := []CardConstr{
		AtLeast1(1, 2, 3),
		AtLeast1(-1, -2, -3),
		AtLeast1(2, 3, 4),
		AtLeast1(2, 3, 5),
		AtLeast1(3, 4, 5),
		AtLeast1(2, 4, 5),
	}
	pb := ParseCardConstrs(clauses)
	s := New(pb)
	if nb := s.Enumerate(nil, nil); nb != 17 {
		t.Errorf("Invalid #models returned: expected %d, got %d", 17, nb)
	}
	models := make(chan []bool)
	pb = ParseCardConstrs(clauses)
	s = New(pb)
	go s.Enumerate(models, nil)
	nb := 0
	for range models {
		nb++
	}
	if nb != 17 {
		t.Errorf("Invalid #models on chan models: expected %d, got %d", 17, nb)
	}

}

func BenchmarkCountModels(b *testing.B) {
	clauses := []CardConstr{
		AtLeast1(1, 2, 3),
		AtLeast1(-1, -2, -3),
		AtLeast1(2, 3, 4),
		AtLeast1(2, 3, 5),
		AtLeast1(3, 4, 5),
		AtLeast1(2, 4, 5),
		AtLeast1(-2, -3, -6),
		AtLeast1(4, 5, 6),
		AtLeast1(1, 2, 3, 4, 5, 6, 7, 8, 9, 10),
		AtLeast1(-7, -10),
		//AtLeast1(11, 12, 13, 14, 15, 16, 17, 18, 19, 20),
		//AtLeast1(21, 22, 23, 24, 25, 26, 27, 28, 29, 30),
		//AtLeast1(50, 100),
	}
	for i := 0; i < b.N; i++ {
		pb := ParseCardConstrs(clauses)
		s := New(pb)
		s.CountModels()
	}
}

func BenchmarkSolverHoons(b *testing.B) {
	runBench("testcnf/hoons-vbmc-lucky7.cnf", b)
}

func BenchmarkSolverXinetd(b *testing.B) {
	runBench("testcnf/xinetd_vc56703.cnf", b)
}

func BenchmarkSolverSmulo(b *testing.B) {
	runBench("testcnf/smulo016.cnf", b)
}

func BenchmarkSolverVMPC(b *testing.B) {
	runBench("testcnf/vmpc_24.cnf", b)
}

func BenchmarkSolverACG(b *testing.B) {
	runBench("testcnf/ACG-10-5p0.cnf", b)
}

func BenchmarkSolverGSS(b *testing.B) {
	runBench("testcnf/gss-13-s100.cnf", b)
}

func BenchmarkSolverGUS(b *testing.B) {
	runBench("testcnf/gus-md5-04.cnf", b)
}

func BenchmarkSolverHSAT(b *testing.B) {
	runBench("testcnf/hsat_vc11803.cnf", b)
}

func BenchmarkSolverEqAtree(b *testing.B) {
	runBench("testcnf/eq.atree.braun.9.unsat.cnf", b)
}

func BenchmarkSolverManolPipe(b *testing.B) {
	runBench("testcnf/manol-pipe-c9.cnf", b)
}
