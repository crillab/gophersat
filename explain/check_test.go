package explain

import (
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"
	"testing"

	"../solver"
)

func TestUnsat(t *testing.T) {
	const cnf = `p cnf 4 8
	c This is a simple, UNSAT problem

	 1  2 -3 0
	-1 -2  3 0
	 2  3 -4 0
	-2 -3  4 0
	 1  3  4 0
	-1 -3 -4 0
	-1  2  4 0
	 1 -2 -4 0`
	const cert = `
	c This is a certificate that proves the problem is UNSAT
	1 2 0
	1 0
	2 0
	0`
	const cert2 = `
	c This certificate does NOT prove the problem is UNSAT, even though the problem is
	-1 -2 0
	0`
	pb, err := ParseCNF(strings.NewReader(cnf))
	if err != nil {
		t.Fatalf("could not parse cnf: %v", err)
	}
	ok, err := pb.Unsat(strings.NewReader(cert))
	if err != nil {
		t.Errorf("%v", err)
	} else if !ok {
		t.Errorf("certificate proof failed")
	}
	ok, err = pb.Unsat(strings.NewReader(cert2))
	if err != nil {
		t.Errorf("%v", err)
	} else if ok {
		t.Errorf("invalid certificate proof succeeded")
	}
	cnf2, err := os.Open("testcnf/125.cnf")
	if err != nil {
		t.Errorf("could not read CNF file: %v", err)
		return
	}
	defer cnf2.Close()
	pb, err = ParseCNF(cnf2)
	if err != nil {
		t.Fatalf("could not parse cnf: %v", err)
	}
	cert3, err := os.Open("testcnf/125_cert.out")
	if err != nil {
		t.Errorf("could not read certificate file: %v", err)
		return
	}
	defer cert3.Close()
	ok, err = pb.Unsat(cert3)
	if err != nil {
		t.Errorf("%v", err)
	} else if !ok {
		t.Errorf("certificate proof failed")
	}
}

func TestUnsatChan(t *testing.T) {
	const cnf = `p cnf 4 8
	c This is a simple, UNSAT problem

	 1  2 -3 0
	-1 -2  3 0
	 2  3 -4 0
	-2 -3  4 0
	 1  3  4 0
	-1 -3 -4 0
	-1  2  4 0
	 1 -2 -4 0`
	const cert = `
	c This is a certificate that proves the problem is UNSAT
	1 2 0
	1 0
	2 0
	0`
	pb, err := ParseCNF(strings.NewReader(cnf))
	if err != nil {
		t.Fatalf("could not parse cnf: %v", err)
	}
	ch := make(chan string)
	go func() {
		defer close(ch)
		for _, line := range strings.Split(cert, "\n") {
			ch <- line
		}
	}()
	ok, err := pb.UnsatChan(ch)
	if err != nil {
		t.Errorf("%v", err)
	} else if !ok {
		t.Errorf("certificate proof failed")
	}
}

func TestUnsatSubset(t *testing.T) {
	const cnf = `p cnf 4 8
	c This is a simple, UNSAT problem

	 1  2 -3 0
	-1 -2  3 0
	 2  3 -4 0
	-2 -3  4 0
	 1  3  4 0
	-1 -3 -4 0
	-1  2  4 0
	 1 -2 -4 0`
	const cert = `
	c This is a certificate that proves the problem is UNSAT
	1 2 0
	1 0
	2 0
	0`
	pb, err := ParseCNF(strings.NewReader(cnf))
	if err != nil {
		t.Fatalf("could not parse cnf: %v", err)
	}
	subset, err := pb.UnsatSubset()
	if err != nil {
		t.Fatalf("could not extract subset: %v", err)
	}
	s := solver.New(solver.ParseSlice(subset.Clauses))
	if s.Solve() != solver.Unsat {
		t.Errorf("unsat subset was satisfiable")
	}
}

func TestMUS(t *testing.T) {
	cnf, err := os.Open("testcnf/50.cnf")
	if err != nil {
		t.Errorf("could not read CNF file: %v", err)
		return
	}
	defer cnf.Close()
	pb, err := ParseCNF(cnf)
	if err != nil {
		t.Fatalf("could not parse cnf: %v", err)
	}
	mus, err := pb.MUS()
	if err != nil {
		t.Fatalf("could not extract subset: %v", err)
	}
	s := solver.New(solver.ParseSlice(mus.Clauses))
	if s.Solve() != solver.Unsat {
		t.Errorf("mus was satisfiable")
	}
}
func TestMUSMaxSat(t *testing.T) {
	cnf, err := os.Open("testcnf/50.cnf")
	if err != nil {
		t.Errorf("could not read CNF file: %v", err)
		return
	}
	defer cnf.Close()
	pb, err := ParseCNF(cnf)
	if err != nil {
		t.Fatalf("could not parse cnf: %v", err)
	}
	mus, err := pb.MUSMaxSat()
	if err != nil {
		t.Fatalf("could not extract subset: %v", err)
	}
	s := solver.New(solver.ParseSlice(mus.Clauses))
	if s.Solve() != solver.Unsat {
		t.Errorf("mus was satisfiable")
	}
}

func TestMUSInsertion(t *testing.T) {
	cnf, err := os.Open("testcnf/50.cnf")
	if err != nil {
		t.Errorf("could not read CNF file: %v", err)
		return
	}
	defer cnf.Close()
	pb, err := ParseCNF(cnf)
	if err != nil {
		t.Fatalf("could not parse cnf: %v", err)
	}
	mus, err := pb.MUSInsertion()
	if err != nil {
		t.Fatalf("could not extract subset: %v", err)
	}
	s := solver.New(solver.ParseSlice(mus.Clauses))
	if s.Solve() != solver.Unsat {
		t.Errorf("mus was satisfiable")
	}
}

func TestMUSDeletion(t *testing.T) {
	cnf, err := os.Open("testcnf/50.cnf")
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
		t.Fatalf("could not extract subset: %v", err)
	}
	s := solver.New(solver.ParseSlice(mus.Clauses))
	if s.Solve() != solver.Unsat {
		t.Errorf("mus was satisfiable")
	}
}

func ExampleProblem_CNF() {
	const cnf = `p cnf 3 3
	c This is a simple problem

	 1  2 -3 0
	-1 -2  3 0
	2 0`
	pb, err := ParseCNF(strings.NewReader(cnf))
	if err != nil {
		fmt.Printf("could not parse problem: %v", err)
	} else {
		fmt.Println(pb.CNF())
	}
	// Output:
	// p cnf 3 3
	// 1 2 -3 0
	// -1 -2 3 0
	// 2 0
}

func ExampleProblem_MUS() {
	const cnf = `p cnf 6 9
	c This is a simple problem

	1  2 -3 0
	-1 -2  3 0
	2 5 0
	6 0
	2 -5 0
	3 4 5 0
	-1 -2 0
	1 3 0
	1 -3 0`
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
	// p cnf 6 5
	// -1 -2 0
	// 1 -3 0
	// 1 3 0
	// 2 -5 0
	// 2 5 0
}

func BenchmarkMUSMaxSat(b *testing.B) {
	content, err := ioutil.ReadFile("testcnf/50.cnf")
	if err != nil {
		b.Errorf("could not read CNF file: %v", err)
		return
	}
	for i := 0; i < b.N; i++ {
		cnf := strings.NewReader(string(content))
		pb, err := ParseCNF(cnf)
		if err != nil {
			b.Fatalf("could not parse cnf: %v", err)
		}
		pb.MUSMaxSat()
	}
}

func BenchmarkMUSInsertion(b *testing.B) {
	content, err := ioutil.ReadFile("testcnf/50.cnf")
	if err != nil {
		b.Errorf("could not read CNF file: %v", err)
		return
	}
	for i := 0; i < b.N; i++ {
		cnf := strings.NewReader(string(content))
		pb, err := ParseCNF(cnf)
		if err != nil {
			b.Fatalf("could not parse cnf: %v", err)
		}
		pb.MUSInsertion()
	}
}

func BenchmarkMUSDeletion(b *testing.B) {
	content, err := ioutil.ReadFile("testcnf/50.cnf")
	if err != nil {
		b.Errorf("could not read CNF file: %v", err)
		return
	}
	for i := 0; i < b.N; i++ {
		cnf := strings.NewReader(string(content))
		pb, err := ParseCNF(cnf)
		if err != nil {
			b.Fatalf("could not parse cnf: %v", err)
		}
		pb.MUSDeletion()
	}
}
