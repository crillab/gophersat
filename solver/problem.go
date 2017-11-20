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

// PBString returns a representation of the problem as a pseudo-boolean problem.
func (pb *Problem) PBString() string {
	res := ""
	for _, clause := range pb.Clauses {
		res += fmt.Sprintf("%s\n", clause.PBString())
	}
	return res
}

func (pb *Problem) updateStatus(nbClauses int) {
	pb.Clauses = pb.Clauses[:nbClauses]
	if pb.Status == Indet && nbClauses == 0 {
		pb.Status = Sat
	}
}

// simplify simplifies the problem, i.e runs unit propagation if possible.
func (pb *Problem) simplify() {
	nbClauses := len(pb.Clauses)
	i := 0
	for i < nbClauses {
		c := pb.Clauses[i]
		nbLits := c.Len()
		card := c.Cardinality()
		clauseSat := false
		nbSat := 0
		j := 0
		for j < nbLits {
			lit := c.Get(j)
			if pb.Model[lit.Var()] == 0 {
				j++
			} else if (pb.Model[lit.Var()] == 1) == lit.IsPositive() {
				nbSat++
				if nbSat == card {
					clauseSat = true
					break
				}
			} else {
				nbLits--
				c.Set(j, c.Get(nbLits))
			}
		}
		if clauseSat {
			nbClauses--
			pb.Clauses[i] = pb.Clauses[nbClauses]
		} else if nbLits < card {
			pb.Status = Unsat
			return
		} else if nbLits == card { // UP
			pb.addUnits(c, nbLits)
			if pb.Status == Unsat {
				return
			}
			nbClauses--
			pb.Clauses[i] = pb.Clauses[nbClauses]
			i = 0 // Must restart, since this lit might have made one more clause Unit or SAT.
		} else { // nb lits unbound > cardinality
			if c.Len() != nbLits {
				c.Shrink(nbLits)
			}
			i++
		}
	}
	pb.updateStatus(nbClauses)
}

func (pb *Problem) simplifyPB() {
	modified := true
	for modified {
		modified = false
		i := 0
		for i < len(pb.Clauses) {
			c := pb.Clauses[i]
			//log.Printf("treating clause %s", c.PBString())
			j := 0
			card := c.Cardinality()
			wSum := c.WeightSum()
			for j < c.Len() {
				lit := c.Get(j)
				v := lit.Var()
				w := c.Weight(j)
				if pb.Model[v] == 0 {
					if wSum-w < card { // Lit must be true for the clause to be satisfiable
						pb.addUnit(c.Get(j))
						if pb.Status == Unsat {
							return
						}
						c.removeLit(j)
						card -= w
						wSum -= w
						modified = true
					} else {
						j++
					}
				} else {
					//log.Printf("found unit: lit is %d, binding is %d", lit.Int(), pb.Model[v])
					wSum -= w
					if (pb.Model[v] == 1) == lit.IsPositive() {
						card -= w
					}
					c.removeLit(j)
					modified = true
					//log.Printf("clause is now %s", c.PBString())
				}
			}
			if card <= 0 { // Clause is Sat
				pb.Clauses[i] = pb.Clauses[len(pb.Clauses)-1]
				pb.Clauses = pb.Clauses[:len(pb.Clauses)-1]
				modified = true
			} else if wSum < card {
				pb.Clauses = nil
				pb.Status = Unsat
				return
			} else {
				i++
			}
		}
	}
	if pb.Status == Indet && len(pb.Clauses) == 0 {
		pb.Status = Sat
	}
}

func (pb *Problem) addUnit(lit Lit) {
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
}

func (pb *Problem) addUnits(c *Clause, nbLits int) {
	for i := 0; i < nbLits; i++ {
		lit := c.Get(i)
		pb.addUnit(lit)
	}
}
