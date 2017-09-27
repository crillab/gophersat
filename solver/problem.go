package solver

import "fmt"

// A Problem is a list of clauses & a nb of vars.
type Problem struct {
	NbVars  int        // Total nb of vars
	Clauses []*Clause  // List of non-empty, non-unit clauses
	Status  Status     // Status of the problem. Can be trivially UNSAT (if empty clause was met or inferred by UP) or Indet.
	Units   []Lit      // List of unit literal found in the problem.
	Model   []decLevel // For each var, its inferred binding. 0 means unbound, 1 means bound to true, -1 means bound to false.
}

// CNF returns a DIMACS CNF representation of the problem.
func (pb *Problem) CNF() string {
	res := fmt.Sprintf("p cnf %d %d\n", pb.NbVars, len(pb.Clauses))
	for _, clause := range pb.Clauses {
		res += fmt.Sprintf("%s\n", clause.CNF())
	}
	return res
}

// simplify simplifies the problem, i.e run unit propagation if possible.
func (pb *Problem) simplify() {
	nbClauses := len(pb.Clauses)
	i := 0
	for i < nbClauses {
		c := pb.Clauses[i]
		nbLits := c.Len()
		clauseSat := false
		j := 0
		for j < nbLits {
			lit := c.Get(j)
			if pb.Model[lit.Var()] == 0 {
				j++
			} else if (pb.Model[lit.Var()] == 1) == lit.IsPositive() {
				clauseSat = true
				break
			} else {
				nbLits--
				c.Set(j, c.Get(nbLits))
			}
		}
		if clauseSat {
			nbClauses--
			pb.Clauses[i] = pb.Clauses[nbClauses]
		} else if nbLits == 0 {
			pb.Status = Unsat
			return
		} else if nbLits == 1 { // UP
			lit := c.First()
			if lit.IsPositive() {
				if pb.Model[lit.Var()] == -1 {
					pb.Status = Unsat
					return
				}
				pb.Model[lit.Var()] = 1
			} else {
				if pb.Model[lit.Var()] == 1 {
					pb.Status = Unsat
					return
				}
				pb.Model[lit.Var()] = -1
			}
			pb.Units = append(pb.Units, lit)
			nbClauses--
			pb.Clauses[i] = pb.Clauses[nbClauses]
			i = 0 // Must restart, since this lit might have made one more clause Unit or SAT.
		} else { // 2 or more lits unbound
			if c.Len() != nbLits {
				c.Shrink(nbLits)
			}
			i++
		}
	}
	pb.Clauses = pb.Clauses[:nbClauses]
	if pb.Status == Indet && nbClauses == 0 {
		pb.Status = Sat
	}
}
