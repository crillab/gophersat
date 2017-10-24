package solver

import (
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
		t.Fatalf("Invalid result for %q: expected %v, got %v", test.path, test.expected, status)
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
	{"testcnf/275.cnf", Sat},
	{"testcnf/300.cnf", Sat},
	{"testcnf/hoons-vbmc-lucky7.cnf", Unsat},
	// Commented out because solving this problem takes a very long times and makes go test crash because of the timeout
	/*{"testcnf/325.cnf", Sat},*/
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
	pb, err := ParseSlice(cnf)
	if err != nil {
		t.Fatalf("could not parse cnf %v: %v", cnf, err)
	}
	s := New(pb)
	if status := s.Solve(); status != Unsat {
		t.Fatalf("expected unsat for problem %v, got %v", cnf, status)
	}
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
