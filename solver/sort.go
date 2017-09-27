package solver

import "sort"

// clauseSorter is a structure to facilitate the sorting of lits in a learned clause
// according to their respective decision levels.
type clauseSorter struct {
	lits  []Lit
	model Model
}

func (cs *clauseSorter) Len() int { return len(cs.lits) }
func (cs *clauseSorter) Less(i, j int) bool {
	return abs(cs.model[cs.lits[i].Var()]) > abs(cs.model[cs.lits[j].Var()])
}
func (cs *clauseSorter) Swap(i, j int) { cs.lits[i], cs.lits[j] = cs.lits[j], cs.lits[i] }

// sortLiterals sorts the literals depending on the decision level they were bound
// i.e. abs(model[lits[i]]) <= abs(model[lits[i+1]]).
func sortLiterals(lits []Lit, model []decLevel) {
	cs := &clauseSorter{lits, model}
	sort.Sort(cs)
}
