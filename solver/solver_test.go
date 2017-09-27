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
	{"/Users/delorme/prog/sat/25.cnf", Sat},
	{"/Users/delorme/prog/sat/50.cnf", Sat},
	{"/Users/delorme/prog/sat/75.cnf", Sat},
	{"/Users/delorme/prog/sat/100.cnf", Sat},
	{"/Users/delorme/prog/sat/125.cnf", Unsat},
	{"/Users/delorme/prog/sat/150.cnf", Unsat},
	{"/Users/delorme/prog/sat/175.cnf", Unsat},
	{"/Users/delorme/prog/sat/200.cnf", Unsat},
	{"/Users/delorme/prog/sat/225.cnf", Sat},
	{"/Users/delorme/prog/sat/250.cnf", Unsat},
	{"/Users/delorme/prog/sat/275.cnf", Sat},
	{"/Users/delorme/prog/sat/300.cnf", Sat},
	{"/Users/delorme/prog/sat/325.cnf", Sat},
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

func BenchmarkSolver125(b *testing.B) {
	runBench("/Users/delorme/prog/sat/125.cnf", b)
}

func BenchmarkSolver150(b *testing.B) {
	runBench("/Users/delorme/prog/sat/150.cnf", b)
}

func BenchmarkSolver175(b *testing.B) {
	runBench("/Users/delorme/prog/sat/175.cnf", b)
}

func BenchmarkSolver200(b *testing.B) {
	runBench("/Users/delorme/prog/sat/200.cnf", b)
}

func BenchmarkSolver225(b *testing.B) {
	runBench("/Users/delorme/prog/sat/225.cnf", b)
}

func BenchmarkSolver250(b *testing.B) {
	runBench("/Users/delorme/prog/sat/250.cnf", b)
}

func BenchmarkSolver275(b *testing.B) {
	runBench("/Users/delorme/prog/sat/275.cnf", b)
}

func BenchmarkSolver300(b *testing.B) {
	runBench("/Users/delorme/prog/sat/300.cnf", b)
}

func BenchmarkSolver325(b *testing.B) {
	runBench("/Users/delorme/prog/sat/325.cnf", b)
}

func BenchmarkSolverHuge(b *testing.B) {
	runBench("/Users/delorme/prog/sat/easy-huge.cnf", b)
}
