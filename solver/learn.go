package solver

// computeLbd computes and sets c's LBD (Literal Block Distance).
func (c *Clause) computeLbd(model Model) {
	c.setLbd(1)
	curLvl := abs(model[c.Get(0).Var()])
	for i := 0; i < c.Len(); i++ {
		lit := c.Get(i)
		if lvl := abs(model[lit.Var()]); lvl != curLvl {
			curLvl = lvl
			c.incLbd()
		}
	}
}

// addClauseLits is a helper function for learnClause.
// It deals with lits from the conflict clause.
func (s *Solver) addClauseLits(confl *Clause, lvl decLevel, met, metLvl []bool, lits *[]Lit) int {
	nbLvl := 0
	for i := 0; i < confl.Len(); i++ {
		l := confl.Get(i)
		v := l.Var()
		if s.litStatus(l) != Unsat {
			// In clauses where cardinality > 1, some lits might be true in the conflict clause: ignore them
			continue
		}
		met[v] = true
		s.varBumpActivity(v)
		if abs(s.model[v]) == lvl {
			metLvl[v] = true
			nbLvl++
		} else if abs(s.model[v]) != 1 {
			*lits = append(*lits, l)
		}
	}
	return nbLvl
}

var bufLits = make([]Lit, 10000) // Buffer for lits in learnClause. Used to reduce allocations.

// learnClause creates a conflict clause and returns either:
// the clause itself, if its len is at least 2.
// a nil clause and a unit literal, if its len is exactly 1.
func (s *Solver) learnClause(confl *Clause, lvl decLevel) (learned *Clause, unit Lit) {
	s.clauseBumpActivity(confl)
	lits := bufLits[:1]             // Not 0: make room for asserting literal
	buf := make([]bool, s.nbVars*2) // Buffer for met and metLvl; reduces allocs/deallocs
	met := buf[:s.nbVars]           // List of all vars already met
	metLvl := buf[s.nbVars:]        // List of all vars from current level to deal with
	// nbLvl is the nb of vars in lvl currently used
	nbLvl := s.addClauseLits(confl, lvl, met, metLvl, &lits)
	ptr := len(s.trail) - 1 // Pointer in propagation trail
	for nbLvl > 1 {         // We will stop once we only have one lit from current level.
		for !metLvl[s.trail[ptr].Var()] {
			if abs(s.model[s.trail[ptr].Var()]) == lvl { // This var was deduced afterwards and was not a reason for the conflict
				met[s.trail[ptr].Var()] = true
			}
			ptr--
		}
		v := s.trail[ptr].Var()
		ptr--
		nbLvl--
		if reason := s.reason[v]; reason != nil {
			s.clauseBumpActivity(reason)
			for i := 0; i < reason.Len(); i++ {
				lit := reason.Get(i)
				if v2 := lit.Var(); !met[v2] {
					if s.litStatus(lit) != Unsat {
						continue
					}
					met[v2] = true
					s.varBumpActivity(v2)
					if abs(s.model[v2]) == lvl {
						metLvl[v2] = true
						nbLvl++
					} else if abs(s.model[v2]) != 1 {
						lits = append(lits, lit)
					}
				}
			}
		}
	}
	for _, l := range s.trail { // Look for last lit from lvl and use it as asserting lit
		if metLvl[l.Var()] {
			lits[0] = l.Negation()
			break
		}
	}
	s.varDecayActivity()
	s.clauseDecayActivity()
	sortLiterals(lits, s.model)
	sz := s.minimizeLearned(met, lits)
	if sz == 1 {
		return nil, lits[0]
	}
	learned = NewLearnedClause(alloc.newLits(lits[0:sz]...))
	learned.computeLbd(s.model)
	return learned, -1
}

// minimizeLearned reduces (if possible) the length of the learned clause and returns the size
// of the new list of lits.
func (s *Solver) minimizeLearned(met []bool, learned []Lit) int {
	sz := 1
	for i := 1; i < len(learned); i++ {
		if reason := s.reason[learned[i].Var()]; reason == nil {
			learned[sz] = learned[i]
			sz++
		} else {
			for k := 0; k < reason.Len(); k++ {
				lit := reason.Get(k)
				if !met[lit.Var()] && abs(s.model[lit.Var()]) > 1 {
					learned[sz] = learned[i]
					sz++
					break
				}
			}
		}
	}
	return sz
}
