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

func runMinTest(test optimTest, t *testing.T) {
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
	if cost := s.Minimize(); cost != test.cost {
		t.Errorf("Invalid result while minimizing %q: expected cost %d, got %d", test.path, test.cost, cost)
	}
}

func runOptimTest(test optimTest, results chan Result, t *testing.T) {
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
	cost := -1
	if results == nil {
		res := s.Optimal(nil, nil)
		cost = res.Weight
		if res.Status != Sat {
			cost = -1
		}
	} else {
		go s.Optimal(results, nil)
		for res := range results {
			cost = res.Weight
			if res.Status != Sat {
				cost = -1
			}
		}
	}
	if cost != test.cost {
		t.Errorf("Invalid result while minimizing %q with chan results = %v: expected cost %d, got %d", test.path, results, test.cost, cost)
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
		runMinTest(test, t)
	}
}

func TestOptimal(t *testing.T) {
	for _, test := range optimTests {
		runOptimTest(test, nil, t)
		runOptimTest(test, make(chan Result), t)
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
