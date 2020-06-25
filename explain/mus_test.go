package explain

import (
	"fmt"
	"sort"
	"strings"
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
