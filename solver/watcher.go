package solver

import (
	"fmt"
	"sort"
)

type watcher struct {
	other  Lit // Another lit from the clause
	clause *Clause
}

// A watcherList is a structure used to store clauses and propagate unit literals efficiently.
type watcherList struct {
	nbMax     int         // Max # of learned clauses at current moment
	idxReduce int         // # of calls to reduce + 1
	wlistBin  [][]watcher // For each literal, a list of binary clauses where its negation appears
	wlist     [][]*Clause // For each literal, a list of non-binary clauses where its negation appears at a position <= (cardinality + 1)
	pbClauses []*Clause   // All the problem clauses.
	learned   []*Clause
}

// initWatcherList makes a new watcherList for the solver.
func (s *Solver) initWatcherList(clauses []*Clause) {
	nbMax := initNbMaxClauses
	newClauses := make([]*Clause, len(clauses))
	copy(newClauses, clauses)
	s.wl = watcherList{
		nbMax:     nbMax,
		idxReduce: 1,
		wlistBin:  make([][]watcher, s.nbVars*2),
		wlist:     make([][]*Clause, s.nbVars*2),
		pbClauses: newClauses,
	}
	for _, c := range clauses {
		s.watchClause(c)
	}
}

// appendClause appends the clause without checking whether the clause is already satisfiable, unit, or unsatisfiable.
// To perform those checks, call s.AppendClause.
// clause is supposed to be a problem clause, not a learned one.
func (s *Solver) appendClause(clause *Clause) {
	s.wl.pbClauses = append(s.wl.pbClauses, clause)
	s.watchClause(clause)
}

// bumpNbMax increases the max nb of clauses used.
// It is typically called after a restart.
func (s *Solver) bumpNbMax() {
	s.wl.nbMax += incrNbMaxClauses
}

// postponeNbMax increases the max nb of clauses used.
// It is typically called when too many good clauses were learned and a cleaning was expected.
func (s *Solver) postponeNbMax() {
	s.wl.nbMax += incrPostponeNbMax
}

// Utilities for sorting according to clauses' LBD and activities.
func (wl *watcherList) Len() int      { return len(wl.learned) }
func (wl *watcherList) Swap(i, j int) { wl.learned[i], wl.learned[j] = wl.learned[j], wl.learned[i] }

func (wl *watcherList) Less(i, j int) bool {
	lbdI := wl.learned[i].lbd()
	lbdJ := wl.learned[j].lbd()
	// Sort by lbd, break ties by activity
	return lbdI > lbdJ || (lbdI == lbdJ && wl.learned[i].activity < wl.learned[j].activity)
}

// Watches the provided clause.
func (s *Solver) watchClause(c *Clause) {
	if c.Len() == 2 {
		first := c.First()
		second := c.Second()
		neg0 := first.Negation()
		neg1 := second.Negation()
		s.wl.wlistBin[neg0] = append(s.wl.wlistBin[neg0], watcher{clause: c, other: second})
		s.wl.wlistBin[neg1] = append(s.wl.wlistBin[neg1], watcher{clause: c, other: first})
	} else {
		if c.PseudoBoolean() {
			w := 0
			i := 0
			for i < 2 || w <= c.Cardinality() {
				lit := c.Get(i)
				neg := lit.Negation()
				s.wl.wlist[neg] = append(s.wl.wlist[neg], c)
				c.watched[i] = true
				w += c.Weight(i)
				i++
			}
		} else {
			for i := 0; i < c.Cardinality()+1; i++ {
				lit := c.Get(i)
				neg := lit.Negation()
				s.wl.wlist[neg] = append(s.wl.wlist[neg], c)
			}
		}
	}
}

// unwatch the given learned clause.
// NOTE: since it is only called when c.lbd() > 2, we know for sure
// that c is not a binary clause.
// We also know for sure this is a propositional clause, since only those are learned.
func (s *Solver) unwatchClause(c *Clause) {
	for i := 0; i < 2; i++ { // 2, not Cardinality + 1: learned clauses always have Cardinality == 1.
		neg := c.Get(i).Negation()
		j := 0
		length := len(s.wl.wlist[neg])
		// We're looking for the index of the clause.
		// This will panic if c is not in wlist[neg], but this shouldn't happen.
		for s.wl.wlist[neg][j] != c {
			j++
		}
		s.wl.wlist[neg][j] = s.wl.wlist[neg][length-1]
		s.wl.wlist[neg] = s.wl.wlist[neg][:length-1]
	}
}

func (s *Solver) tryUnwatchClause(c *Clause) {
	for i := 0; i < 2; i++ {
		lit := c.Get(i)
		if s.model[lit.Var()] != 0 {
			continue
		}
		neg := lit.Negation()
		j := 0
		length := len(s.wl.wlist[neg])
		// We're looking for the index of the clause.
		// This will panic if c is not in wlist[neg], but this shouldn't happen.
		for s.wl.wlist[neg][j] != c {
			j++
		}
		s.wl.wlist[neg][j] = s.wl.wlist[neg][length-1]
		s.wl.wlist[neg] = s.wl.wlist[neg][:length-1]
	}
}

// reduceLearned removes a few learned clauses that are deemed useless.
func (s *Solver) reduceLearned() {
	sort.Sort(&s.wl)
	nbLearned := len(s.wl.learned)
	length := nbLearned / 2
	if s.wl.learned[length].lbd() <= 3 { // Lots of good clauses, postpone reduction
		s.postponeNbMax()
	}
	nbRemoved := 0
	for i := 0; i < length; i++ {
		c := s.wl.learned[i]
		if c.lbd() <= 2 || c.isLocked() {
			continue
		}
		nbRemoved++
		s.Stats.NbDeleted++
		s.wl.learned[i] = s.wl.learned[nbLearned-nbRemoved]
		s.unwatchClause(c)
	}
	nbLearned -= nbRemoved
	s.wl.learned = s.wl.learned[:nbLearned]
}

// Adds the given learned clause and updates watchers.
// If too many clauses have been learned yet, one will be removed.
func (s *Solver) addLearned(c *Clause) {
	s.wl.learned = append(s.wl.learned, c)
	s.watchClause(c)
	s.clauseBumpActivity(c)
}

// If l is negative, -lvl is returned. Else, lvl is returned.
func lvlToSignedLvl(l Lit, lvl decLevel) decLevel {
	if l.IsPositive() {
		return lvl
	}
	return -lvl
}

// Removes the first occurrence of c from lst.
// The element *must* be present into lst.
func removeFrom(lst []*Clause, c *Clause) []*Clause {
	i := 0
	for lst[i] != c {
		i++
	}
	last := len(lst) - 1
	lst[i] = lst[last]
	return lst[:last]
}

// Unifies the given literal and returns a conflict clause, or nil if no conflict arose.
func (s *Solver) unifyLiteral(lit Lit, lvl decLevel) *Clause {
	s.model[lit.Var()] = lvlToSignedLvl(lit, lvl)
	ptr := len(s.trail)
	s.trail = append(s.trail, lit)
	for ptr < len(s.trail) {
		lit := s.trail[ptr]
		for _, w := range s.wl.wlistBin[lit] {
			v2 := w.other.Var()
			if assign := s.model[v2]; assign == 0 { // Other was unbounded: propagate
				s.reason[v2] = w.clause
				w.clause.lock()
				s.model[v2] = lvlToSignedLvl(w.other, lvl)
				s.trail = append(s.trail, w.other)
			} else if (assign > 0) != w.other.IsPositive() { // Conflict here
				return w.clause
			}
		}
		for _, c := range s.wl.wlist[lit] {
			if c.Cardinality() == 1 {
				res, unit := s.simplifyClause(c)
				if res == Unsat {
					return c
				} else if res == Unit {
					s.propagateUnit(c, lvl, unit)
				}
			} else if c.PseudoBoolean() {
				if !s.simplifyPseudoBool(c, lvl) {
					return c
				}
			} else {
				if !s.simplifyCardClause(c, lvl) {
					return c
				}
			}
		}
		ptr++
	}
	// No unsat clause was met
	return nil
}

func (s *Solver) propagateUnit(c *Clause, lvl decLevel, unit Lit) {
	if s.litStatus(unit) != Indet {
		panic(fmt.Errorf("could not propagate %d at level %d: its binding is %d", unit.Int(), lvl, s.model[unit.Var()]))
	}
	v := unit.Var()
	s.reason[v] = c
	c.lock()
	s.model[v] = lvlToSignedLvl(unit, lvl)
	s.trail = append(s.trail, unit)
}

// simplifyClause simplifies the given clause according to current binding.
// It assumes the cardinality of the clause is 1.
// It returns a new status, and a potential unit literal.
func (s *Solver) simplifyClause(clause *Clause) (Status, Lit) {
	var freeIdx int // Index of the first free lit found, if any
	found := false
	len := clause.Len()
	for i := 0; i < len; i++ {
		lit := clause.Get(i)
		if status := s.litStatus(lit); status == Indet {
			if found {
				// 2 lits are known to be unbounded
				switch freeIdx {
				case 0: // c[0] is not removed, c[1] is
					n1 := &s.wl.wlist[clause.Second().Negation()]
					nf1 := &s.wl.wlist[clause.Get(i).Negation()]
					clause.swap(i, 1)
					*n1 = removeFrom(*n1, clause)
					*nf1 = append(*nf1, clause)
				case 1: // c[0] is removed, not c[1]
					n0 := &s.wl.wlist[clause.First().Negation()]
					nf1 := &s.wl.wlist[clause.Get(i).Negation()]
					clause.swap(i, 0)
					*n0 = removeFrom(*n0, clause)
					*nf1 = append(*nf1, clause)
				default: // Both c[0] & c[1] are removed
					n0 := &s.wl.wlist[clause.First().Negation()]
					n1 := &s.wl.wlist[clause.Second().Negation()]
					nf0 := &s.wl.wlist[clause.Get(freeIdx).Negation()]
					nf1 := &s.wl.wlist[clause.Get(i).Negation()]
					clause.swap(freeIdx, 0)
					clause.swap(i, 1)
					*n0 = removeFrom(*n0, clause)
					*n1 = removeFrom(*n1, clause)
					*nf0 = append(*nf0, clause)
					*nf1 = append(*nf1, clause)
				}
				return Many, -1
			}
			freeIdx = i
			found = true
		} else if status == Sat {
			return Sat, -1
		}
	}
	if !found {
		return Unsat, -1
	}
	return Unit, clause.Get(freeIdx)
}

// simplifyCardClauses simplifies a clause of cardinality > 1, but with all weights = 1.
// returns false iff the clause cannot be satisfied.
func (s *Solver) simplifyCardClause(clause *Clause, lvl decLevel) bool {
	length := clause.Len()
	card := clause.Cardinality()
	nbTrue := 0
	nbFalse := 0
	nbUnb := 0
	for i := 0; i < length; i++ {
		lit := clause.Get(i)
		switch s.litStatus(lit) {
		case Indet:
			nbUnb++
			if nbUnb+nbTrue > card {
				break
			}
		case Sat:
			nbTrue++
			if nbTrue == card {
				return true
			}
			if nbUnb+nbTrue > card {
				break
			}
		case Unsat:
			nbFalse++
			if length-nbFalse < card {
				return false
			}
		}
	}
	if nbTrue >= card {
		return true
	}
	if nbUnb+nbTrue == card {
		// All unbounded lits must be bound to make the clause true
		i := 0
		for nbUnb > 0 {
			lit := clause.Get(i)
			if s.model[lit.Var()] == 0 {
				s.propagateUnit(clause, lvl, lit)
				nbUnb--
			} else {
				i++
			}
		}
		return true
	}
	s.swapFalse(clause)
	return true
}

// swapFalse swaps enough literals from the clause so that all watching literals are either true or unbounded lits.
// Must only be called when there a at least cardinality + 1 true and unbounded lits.
func (s *Solver) swapFalse(clause *Clause) {
	card := clause.Cardinality()
	i := 0
	j := card + 1
	for i < card+1 {
		lit := clause.Get(i)
		for s.litStatus(lit) != Unsat {
			i++
			if i == card+1 {
				return
			}
			lit = clause.Get(i)
		}
		lit = clause.Get(j)
		for s.litStatus(lit) == Unsat {
			j++
			lit = clause.Get(j)
		}
		ni := &s.wl.wlist[clause.Get(i).Negation()]
		nj := &s.wl.wlist[clause.Get(j).Negation()]
		clause.swap(i, j)
		*ni = removeFrom(*ni, clause)
		*nj = append(*nj, clause)
		i++
		j++
	}
}

// sumWeights returns the sum of weights of the given PB clause, if the clause is not SAT yet.
// If the clause is already SAT, it returns true and the given sum value
// is meaningless.
func (s *Solver) sumWeights(c *Clause) (sum int, sat bool) {
	sum = 0
	sumSat := 0
	card := c.Cardinality()
	for i, v := range c.weights {
		if status := s.litStatus(c.Get(i)); status == Indet {
			sum += v
		} else if status == Sat {
			sum += v
			sumSat += v
			if sumSat >= card {
				return sum, true
			}
		}
	}
	return sum, false
}

// propagateAll propagates all unbounded literals from c as unit literals
func (s *Solver) propagateAll(c *Clause, lvl decLevel) {
	for i := 0; i < c.Len(); i++ {
		if lit := c.Get(i); s.litStatus(lit) == Indet {
			s.propagateUnit(c, lvl, lit)
		}
	}
}

// simplifyPseudoBool simplifies a pseudo-boolean constraint.
// propagates unit literals, if any.
// returns false if the clause cannot be satisfied.
func (s *Solver) simplifyPseudoBool(clause *Clause, lvl decLevel) bool {
	card := clause.Cardinality()
	foundUnit := true
	for foundUnit {
		sumW, sat := s.sumWeights(clause)
		if sat {
			return true
		}
		if sumW < card {
			return false
		}
		if sumW == card {
			s.propagateAll(clause, lvl)
			return true
		}
		foundUnit = false
		for i := 0; i < clause.Len(); i++ {
			lit := clause.Get(i)
			if s.litStatus(lit) == Indet && sumW-clause.Weight(i) < card { // lit can't be falsified
				s.propagateUnit(clause, lvl, lit)
				foundUnit = true
			}
		}
	}
	s.updateWatchPB(clause)
	return true
}

func (s *Solver) updateWatchPB(clause *Clause) {
	weightWatched := 0
	i := 0
	card := clause.Cardinality()
	for weightWatched <= card && i < clause.Len() {
		lit := clause.Get(i)
		if s.litStatus(lit) == Unsat {
			if clause.watched[i] {
				ni := &s.wl.wlist[lit.Negation()]
				*ni = removeFrom(*ni, clause)
				clause.watched[i] = false
			}
		} else {
			weightWatched += clause.Weight(i)
			if !clause.watched[i] {
				ni := &s.wl.wlist[lit.Negation()]
				*ni = append(*ni, clause)
				clause.watched[i] = true
			}
		}
		i++
	}
	// If there are some more watched literals, they are now useless
	for i := i; i < clause.Len(); i++ {
		if clause.watched[i] {
			ni := &s.wl.wlist[clause.Get(i).Negation()]
			*ni = removeFrom(*ni, clause)
			clause.watched[i] = false
		}
	}
}
