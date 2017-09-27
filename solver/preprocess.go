package solver

//
// import "log"
//
// // Subsumes returns true iff c subsumes c2.
// func (c *Clause) Subsumes(c2 *Clause) bool {
// 	if c.Len() > c2.Len() {
// 		return false
// 	}
// 	for _, lit := range c.lits {
// 		found := false
// 		for _, lit2 := range c2.lits {
// 			if lit == lit2 {
// 				found = true
// 				break
// 			}
// 			if lit2 > lit {
// 				return false
// 			}
// 		}
// 		if !found {
// 			return false
// 		}
// 	}
// 	return true
// }
//
// // SelfSubsumes returns true iff c self-subsumes c2.
// func (c *Clause) SelfSubsumes(c2 *Clause) bool {
// 	oneNeg := false
// 	for _, lit := range c.lits {
// 		found := false
// 		for _, lit2 := range c2.lits {
// 			if lit == lit2 {
// 				found = true
// 				break
// 			}
// 			if lit == lit2.Negation() {
// 				if oneNeg { // We want exactly one, but this is the second
// 					return false
// 				}
// 				oneNeg = true
// 				found = true
// 				break
// 			}
// 			if lit2 > lit { // We won't find it anymore
// 				return false
// 			}
// 		}
// 		if !found {
// 			return false
// 		}
// 	}
// 	return oneNeg
// }
//
// // Simplify simplifies the given clause by removing redundant lits.
// // If the clause is trivially satisfied (i.e contains both a lit and its negation),
// // true is returned. Otherwise, false is returned.
// func (c *Clause) Simplify() (isSat bool) {
// 	c.Sort()
// 	lits := make([]Lit, 0, len(c.lits))
// 	i := 0
// 	for i < len(c.lits) {
// 		if i < len(c.lits)-1 && c.lits[i] == c.lits[i+1].Negation() {
// 			return true
// 		}
// 		lit := c.lits[i]
// 		lits = append(lits, lit)
// 		i++
// 		for i < len(c.lits) && c.lits[i] == lit {
// 			i++
// 		}
// 	}
// 	if len(lits) < len(c.lits) {
// 		c.lits = lits
// 	}
// 	return false
// }
//
// // Generate returns a subsumed clause from c and c2, by removing v.
// func (c *Clause) Generate(c2 *Clause, v Var) *Clause {
// 	c3 := &Clause{lits: make([]Lit, 0, len(c.lits)+len(c2.lits)-2)}
// 	for _, lit := range c.lits {
// 		if lit.Var() != v {
// 			c3.lits = append(c3.lits, lit)
// 		}
// 	}
// 	for _, lit2 := range c2.lits {
// 		if lit2.Var() != v {
// 			c3.lits = append(c3.lits, lit2)
// 		}
// 	}
// 	return c3
// }
//
// func (pb *Problem) preprocess() {
// 	log.Printf("Preprocessing... %d clauses currently", len(pb.Clauses))
// 	occurs := make([][]int, pb.NbVars*2)
// 	for i, c := range pb.Clauses {
// 		for j := 0; j < c.Len(); j++ {
// 			occurs[c.Get(j)] = append(occurs[c.Get(j)], i)
// 		}
// 	}
// 	modified := true
// 	neverModified := true
// 	for modified {
// 		modified = false
// 		for i := 0; i < pb.NbVars; i++ {
// 			if pb.Model[i] != 0 {
// 				continue
// 			}
// 			v := Var(i)
// 			lit := v.Lit()
// 			nbLit := len(occurs[lit])
// 			nbLit2 := len(occurs[lit.Negation()])
// 			if (nbLit < 10 || nbLit2 < 10) && (nbLit != 0 || nbLit2 != 0) {
// 				modified = true
// 				neverModified = false
// 				// pb.deleted[v] = true
// 				log.Printf("%d can be removed: %d and %d", lit.Int(), len(occurs[lit]), len(occurs[lit.Negation()]))
// 				for _, idx1 := range occurs[lit] {
// 					for _, idx2 := range occurs[lit.Negation()] {
// 						c1 := pb.Clauses[idx1]
// 						c2 := pb.Clauses[idx2]
// 						newC := c1.Generate(c2, v)
// 						if !newC.Simplify() {
// 							switch newC.Len() {
// 							case 0:
// 								log.Printf("Inferred UNSAT")
// 								pb.Status = Unsat
// 								return
// 							case 1:
// 								log.Printf("Unit %d", newC.First().Int())
// 								lit2 := newC.First()
// 								if lit2.IsPositive() {
// 									if pb.Model[lit2.Var()] == -1 {
// 										pb.Status = Unsat
// 										return
// 									}
// 									pb.Model[lit2.Var()] = 1
// 								} else {
// 									if pb.Model[lit2.Var()] == 1 {
// 										pb.Status = Unsat
// 										return
// 									}
// 									pb.Model[lit2.Var()] = -1
// 								}
// 								pb.Units = append(pb.Units, lit2)
// 							default:
// 								pb.Clauses = append(pb.Clauses, newC)
// 							}
// 						}
// 					}
// 				}
// 				nbRemoved := 0
// 				for _, idx := range occurs[lit] {
// 					pb.Clauses[idx] = pb.Clauses[len(pb.Clauses)-nbRemoved-1]
// 					nbRemoved++
// 				}
// 				for _, idx := range occurs[lit.Negation()] {
// 					pb.Clauses[idx] = pb.Clauses[len(pb.Clauses)-nbRemoved-1]
// 					nbRemoved++
// 				}
// 				pb.Clauses = pb.Clauses[:len(pb.Clauses)-nbRemoved]
// 				log.Printf("clauses=%s", pb.CNF())
// 				// Redo occurs
// 				occurs = make([][]int, pb.NbVars*2)
// 				for i, c := range pb.Clauses {
// 					for j := 0; j < c.Len(); j++ {
// 						occurs[c.Get(j)] = append(occurs[c.Get(j)], i)
// 					}
// 				}
// 				continue
// 			}
// 		}
// 	}
// 	if !neverModified {
// 		pb.simplify()
// 	}
// 	log.Printf("Done. %d clauses now", len(pb.Clauses))
// }
