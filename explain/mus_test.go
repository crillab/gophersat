package explain

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/crillab/gophersat/solver"
)

func ExampleInstanceIsAMUS() {
	const cnf = `p cnf 1 2
	c This is a simple problem
	1 0
	-1 0`
	pb, err := ParseCNF(strings.NewReader(cnf))
	if err != nil {
		fmt.Printf("could not parse problem: %v", err)
		return
	}
	mus, err := pb.MUS()
	if err != nil {
		fmt.Printf("could not compute MUS: %v", err)
		return
	}
	musCnf := mus.CNF()
	// Sort clauses so as to always have the same output
	lines := strings.Split(musCnf, "\n")
	sort.Sort(sort.StringSlice(lines[1:]))
	musCnf = strings.Join(lines, "\n")
	fmt.Println(musCnf)
	// Output:
	// p cnf 1 2
	// -1 0
	// 1 0
}

func TestTrivialMUS(t *testing.T) {
	cnf, err := os.Open("testcnf/trivial.cnf")
	if err != nil {
		t.Errorf("could not read CNF file: %v", err)
		return
	}
	defer cnf.Close()
	pb, err := ParseCNF(cnf)
	if err != nil {
		t.Fatalf("could not parse cnf: %v", err)
	}
	mus, err := pb.MUSDeletion()
	if err != nil {
		t.Fatalf("could not compute MUS: %v", err)
	}
	s := solver.New(solver.ParseSlice(mus.Clauses))
	if s.Solve() != solver.Unsat {
		t.Errorf("MUS was satisfiable")
	}
}

func TestMUSOnSatisfiableFormula(t *testing.T) {
	cnf, err := os.Open("testcnf/impossible.cnf")
	if err != nil {
		t.Errorf("could not read CNF file: %v", err)
		return
	}
	defer cnf.Close()

	pb, err := ParseCNF(cnf)
	if err != nil {
		t.Fatalf("could not parse cnf: %v", err)
	}

	mus, err := pb.MUSDeletion()
	if err == nil {
		t.Fatal("This function should return an error on satisfiable formula")
		s := solver.New(solver.ParseSlice(mus.Clauses))
		if s.Solve() != solver.Unsat {
			t.Errorf("MUS was satisfiable")
		}
	}
}
