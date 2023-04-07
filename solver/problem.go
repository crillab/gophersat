package solver

import (
	"fmt"
)

// A Problem is a list of clauses & a nb of vars.
type Problem struct {
	NbVars     int        // Total nb of vars
	Clauses    []*Clause  // List of non-empty, non-unit clauses
	Status     Status     // Status of the problem. Can be trivially UNSAT (if empty clause was met or inferred by UP) or Indet.
	Units      []Lit      // List of unit literal found in the problem.
	Model      []decLevel // For each var, its inferred binding. 0 means unbound, 1 means bound to true, -1 means bound to false.
	minLits    []Lit      // For an optimisation problem, the list of lits whose sum must be minimized
	minWeights []int      // For an optimisation problem, the weight of each lit.
}

// Optim returns true iff pb is an optimisation problem, ie
// a problem for which we not only want to find a model, but also
// the best possible model according to an optimization constraint.
func (pb *Problem) Optim() bool {
	return pb.minLits != nil
}

// CNF returns a DIMACS CNF representation of the problem.
func (pb *Problem) CNF() string {
	res := fmt.Sprintf("p cnf %d %d\n", pb.NbVars, len(pb.Clauses)+len(pb.Units))
	for _, unit := range pb.Units {
		res += fmt.Sprintf("%d 0\n", unit.Int())
	}
	for _, clause := range pb.Clauses {
		res += fmt.Sprintf("%s\n", clause.CNF())
	}
	return res
}

// PBString returns a representation of the problem as a pseudo-boolean problem.
func (pb *Problem) PBString() string {
	res := pb.costFuncString()
	for _, unit := range pb.Units {
		sign := ""
		if !unit.IsPositive() {
			sign = "~"
			unit = unit.Negation()
		}
		res += fmt.Sprintf("1 %sx%d = 1 ;\n", sign, unit.Int())
	}
	for _, clause := range pb.Clauses {
		res += fmt.Sprintf("%s\n", clause.PBString())
	}
	return res
}

// SetCostFunc sets the function to minimize when optimizing the problem.
// If all weights are 1, weights can be nil.
// In all other cases, len(lits) must be the same as len(weights).
func (pb *Problem) SetCostFunc(lits []Lit, weights []int) {
	if weights != nil && len(lits) != len(weights) {
		panic("length of lits and of weights don't match")
	}
	pb.minLits = lits
	pb.minWeights = weights
}

// costFuncString returns a string representation of the cost function of the problem, if any, followed by a \n.
// If there is no cost function, the empty string will be returned.
func (pb *Problem) costFuncString() string {
	if pb.minLits == nil {
		return ""
	}
	res := "min: "
	for i, lit := range pb.minLits {
		w := 1
		if pb.minWeights != nil {
			w = pb.minWeights[i]
		}
		sign := ""
		if w >= 0 && i != 0 { // No plus sign for the first term or for negative terms.
			sign = "+"
		}
		val := lit.Int()
		neg := ""
		if val < 0 {
			val = -val
			neg = "~"
		}
		res += fmt.Sprintf("%s%d %sx%d", sign, w, neg, val)
	}
	res += " ;\n"
	return res
}

func (pb *Problem) updateStatus(nbClauses int) {
	pb.Clauses = pb.Clauses[:nbClauses]
	if pb.Status == Indet && nbClauses == 0 {
		pb.Status = Sat
	}
}

// simplify simplifies the pure SAT problem, i.e runs unit propagation if possible.
func (pb *Problem) simplify2() {
	nbClauses := len(pb.Clauses)
	restart := true
	for restart {
		restart = false
		i := 0
		for i < nbClauses {
			c := pb.Clauses[i]
			nbLits := c.Len()
			clauseSat := false
			j := 0
			for j < nbLits {
				lit := c.Get(j)
				k := j + 1
				for k < nbLits {
					lit2 := c.Get(k)
					if lit2 == lit.Negation() {
						clauseSat = true
						break
					}
					if lit2 == lit { // duplicate lit
						nbLits--
						c.Set(k, c.Get(nbLits))
					} else {
						k++
					}
				}
				if clauseSat {
					clauseSat = true
					break
				}
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
				pb.addUnit(c.First())
				if pb.Status == Unsat {
					return
				}
				nbClauses--
				pb.Clauses[i] = pb.Clauses[nbClauses]
				restart = true // Must restart, since this lit might have made one more clause Unit or SAT.
			} else { // nb lits unbound > cardinality
				if c.Len() != nbLits {
					c.Shrink(nbLits)
				}
				i++
			}
		}
	}
	pb.updateStatus(nbClauses)
}

// DetectAtMostOne tries to detect AtMostOne constraints encoded using the pairwise encoding.
// It replaces those binary clauses by a single cardinality constraint.
// This should mostly be called using the CuttingPlanes option, as it can dramatically improve the resolution process in some cases.
func (pb *Problem) DetectAtMostOne() {
	considered := make([]bool, pb.NbVars*2)   // Has lit 1 already been detected in a clique?
	propagates := make([][]Lit, pb.NbVars*2)  // For each lit, the literals it propagates in a binary clause
	indexes := make([][]int, len(propagates)) // Indexes of binary clauses, to remove them efficiently
	toRemove := make([]int, 0, 1_000)         // Indexes of clauses that have to be removed afterwards
	for i, c := range pb.Clauses {
		if c.Len() == 2 {
			lit1 := c.First()
			neg1 := lit1.Negation()
			lit2 := c.Second()
			neg2 := lit2.Negation()
			propagates[neg1] = append(propagates[neg1], lit2)
			propagates[neg2] = append(propagates[neg2], lit1)
			indexes[neg1] = append(indexes[neg1], i)
			indexes[neg2] = append(indexes[neg2], i)
		}
	}
	for i := range propagates {
		if considered[i] {
			continue
		}
		lit := Lit(i)
		others := propagates[lit]
		if len(others) < 2 { // We won't find a cardinality constraint here
			continue
		}
		constr := []Lit{lit.Negation()}
		for j, other := range others {
			if considered[other] {
				continue
			}
			ok := true
			for j := 1; j < len(constr); j++ {
				lit2 := constr[j].Negation()
				found := false
				for _, lit3 := range propagates[lit2] {
					if lit3 == other {
						found = true
						break
					}
				}
				if !found {
					ok = false
					break
				}
			}
			if ok { // other was found in a binary clause with each literal in constr
				constr = append(constr, other)
				toRemove = append(toRemove, indexes[lit][j])
			}
		}
		if len(constr) > 2 { // We detected a stronger cardinality constraint
			for _, lit := range constr {
				considered[lit.Negation()] = true
			}
			pb.Clauses = append(pb.Clauses, NewCardClause(constr, len(constr)-1))
		}
	}
	pb.removeBinaries(toRemove)
}

// removeBinaries removes the binary clauses that were used to build the
// hidden cardinality constraint whose lits are given as a parameter.
func (pb *Problem) removeBinaries(toRemove []int) {
	if len(toRemove) == 0 {
		return
	}
	newClauses := make([]*Clause, 0, len(pb.Clauses)-len(toRemove))
	i := 0
	for j, c := range pb.Clauses {
		if j == toRemove[i] {
			i++
			if i == len(toRemove) {
				break
			}
		} else {
			newClauses = append(newClauses, c)
		}
	}
	pb.Clauses = newClauses
}

// simplifyCard simplifies the problem, i.e runs unit propagation if possible.
func (pb *Problem) simplifyCard() {
	nbClauses := len(pb.Clauses)
	restart := true
	for restart {
		restart = false
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
				restart = true // Must restart, since this lit might have made one more clause Unit or SAT.
			} else { // nb lits unbound > cardinality
				if c.Len() != nbLits {
					c.Shrink(nbLits)
				}
				i++
			}
		}
	}
	pb.updateStatus(nbClauses)
}

func (pb *Problem) simplifyPB() {
	pb.replicateUnits()
	modified := true
	for modified {
		modified = false
		i := 0
		for i < len(pb.Clauses) {
			c := pb.Clauses[i]
			j := 0
			card := c.Cardinality()
			wSum := c.WeightSum()
			for j < c.Len() {
				lit := c.Get(j)
				v := lit.Var()
				w := c.Weight(j)
				if pb.Model[v] == 0 { // Literal not assigned: is it unit?
					if wSum-w < card { // Lit must be true for the clause to be satisfiable
						pb.addUnit(lit)
						if pb.Status == Unsat {
							return
						}
						c.removeLit(j)
						card -= w
						c.updateCardinality(-w)
						wSum -= w
						modified = true
					} else {
						j++
					}
				} else { // Bound literal: remove it and update, if needed, cardinality
					wSum -= w
					if (pb.Model[v] == 1) == lit.IsPositive() {
						card -= w
						c.updateCardinality(-w)
					}
					c.removeLit(j)
					modified = true
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

func (pb *Problem) replicateUnits() {
	for _, unit := range pb.Units {
		v := unit.Var()
		if unit.IsPositive() {
			pb.Model[v] = 1
		} else {
			pb.Model[v] = -1
		}
	}
}
