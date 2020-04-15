package explain

import (
	"fmt"
	"strings"
)

// A Problem is a conjunction of Clauses.
// This package does not use solver's representation.
// We want this code to be as simple as possible to be easy to audit.
// On the other hand, solver's code must be as efficient as possible.
type Problem struct {
	Clauses   [][]int
	NbVars    int
	nbClauses int
	units     []int // For each var, 0 if the var is unbound, 1 if true, -1 if false
	Options   Options
	tagged    []bool // List of claused used whil proving the problem is unsat. Initialized lazily
}

func (pb *Problem) initTagged() {
	pb.tagged = make([]bool, pb.nbClauses)
	for i, clause := range pb.Clauses {
		// Unit clauses are tagged as they will probably be used during resolution
		pb.tagged[i] = len(clause) == 1
	}
}

func (pb *Problem) clone() *Problem {
	pb2 := &Problem{
		Clauses:   make([][]int, pb.nbClauses),
		NbVars:    pb.NbVars,
		nbClauses: pb.nbClauses,
		units:     make([]int, pb.NbVars),
	}
	copy(pb2.units, pb.units)
	for i, clause := range pb.Clauses {
		pb2.Clauses[i] = make([]int, len(clause))
		copy(pb2.Clauses[i], clause)
	}
	return pb2
}

// restore removes all learned clauses, if any.
func (pb *Problem) restore() {
	pb.Clauses = pb.Clauses[:pb.nbClauses]
}

// unsat will be true iff the problem can be proven unsat through unit propagation.
// This methods modifies pb.units.
func (pb *Problem) unsat() bool {
	done := make([]bool, len(pb.Clauses)) // clauses that were deemed sat during propagation
	modified := true
	for modified {
		modified = false
		for i, clause := range pb.Clauses {
			if done[i] { // That clause was already proved true
				continue
			}
			unbound := 0
			var unit int // An unbound literal, if any
			sat := false
			for _, lit := range clause {
				v := lit
				if v < 0 {
					v = -v
				}
				binding := pb.units[v-1]
				if binding == 0 {
					unbound++
					if unbound == 1 {
						unit = lit
					} else {
						break
					}
				} else if binding*lit == v { // (binding == -1 && lit < 0) || (binding == 1 && lit > 0) {
					sat = true
					break
				}
			}
			if sat {
				done[i] = true
				continue
			}
			if unbound == 0 {
				// All lits are false: problem is UNSAT
				if i < pb.nbClauses {
					pb.tagged[i] = true
				}
				return true
			}
			if unbound == 1 {
				if unit < 0 {
					pb.units[-unit-1] = -1
				} else {
					pb.units[unit-1] = 1
				}
				done[i] = true
				if i < pb.nbClauses {
					pb.tagged[i] = true
				}
				modified = true
			}
		}
	}
	// Problem is either sat or could not be proven unsat through unit propagation
	return false
}

// CNF returns a representation of the problem using the Dimacs syntax.
func (pb *Problem) CNF() string {
	lines := make([]string, 1, pb.nbClauses+1)
	lines[0] = fmt.Sprintf("p cnf %d %d", pb.NbVars, pb.nbClauses)
	for i := 0; i < pb.nbClauses; i++ {
		clause := pb.Clauses[i]
		strClause := make([]string, len(clause)+1)
		for i, lit := range clause {
			strClause[i] = fmt.Sprintf("%d", lit)
		}
		strClause[len(clause)] = "0"
		line := strings.Join(strClause, " ")
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}
