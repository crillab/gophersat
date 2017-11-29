package solver

import (
	"os"
	"strings"
	"testing"
)

// An optimTest associates a path with an expected minimization cost, which can be -1 for UNSAT,
// 0 for SAT with no optimization function, or any value for a real optimization problem.
type optimTest struct {
	path string
	cost int
}

func runOptimTest(test optimTest, t *testing.T) {
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
	if cost, _ := s.Minimize(); cost != test.cost {
		t.Errorf("Invalid result while minimizing %q: expected cost %d, got %d", test.path, test.cost, cost)
	}
}

var optimTests = []optimTest{
	{"testcnf/100.cnf", 0},
	{"testcnf/125.cnf", -1},
	{"testcnf/150.cnf", -1},
	{"testcnf/200.cnf", -1},
	{"testcnf/225.cnf", 0},
	{"testcnf/hoons-vbmc-lucky7.cnf", -1},
	{"testcnf/lo_8x8_009.opb", 27},
}

func TestMinimize(t *testing.T) {
	for _, test := range optimTests {
		runOptimTest(test, t)
	}
}

func runOptimBench(path string, b *testing.B) {
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
		s.Minimize()
	}
}

func BenchmarkLo88(b *testing.B) {
	runOptimBench("testcnf/lo_8x8_009.opb", b)
}
