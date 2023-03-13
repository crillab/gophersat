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
	nbMax       int         // Max # of learned clauses at current moment
	idxReduce   int         // # of calls to reduce + 1
	wlistBin    [][]watcher // For each literal, a list of binary clauses where its negation appears
	wlist       [][]watcher // For each literal, a list of non-binary clauses where its negation appears atposition 1 or 2
	wlistPb     [][]*Clause // For each literal a list of PB or cardinality constraints.
	origClauses []*Clause   // All the problem clauses.
	learned     []*Clause
}

// initWatcherList makes a new watcherList for the solver.
func (s *Solver) initWatcherList(clauses []*Clause) {
	nbMax := initNbMaxClauses
	newClauses := make([]*Clause, len(clauses))
	copy(newClauses, clauses)
	s.wl = watcherList{
		nbMax:       nbMax,
		idxReduce:   1,
		wlistBin:    make([][]watcher, s.nbVars*2),
		wlist:       make([][]watcher, s.nbVars*2),
		wlistPb:     make([][]*Clause, s.nbVars*2),
		origClauses: newClauses,
	}
	for _, c := range clauses {
		s.watchClause(c)
	}
}

// Should be called when new vars are added to the problem (see Solver.newVar)
func (s *Solver) addVarWatcherList(v Var) {
	cnfVar := int(v.Int())
	for i := s.nbVars; i < cnfVar; i++ {
		s.wl.wlistBin = append(s.wl.wlistBin, nil, nil)
		s.wl.wlist = append(s.wl.wlist, nil, nil)
		s.wl.wlistPb = append(s.wl.wlistPb, nil, nil)
	}
}

// appendClause appends the clause without checking whether the clause is already satisfiable, unit, or unsatisfiable.
// To perform those checks, call s.AppendClause.
// clause is supposed to be a problem clause, not a learned one.
func (s *Solver) appendClause(clause *Clause) {
	s.wl.origClauses = append(s.wl.origClauses, clause)
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
	ci := wl.learned[i]
	cj := wl.learned[j]
	lbdI := ci.lbd()
	lbdJ := cj.lbd()
	// Sort by lbd, break ties by activity
	return lbdI > lbdJ || (lbdI == lbdJ && wl.learned[i].activity < wl.learned[j].activity)
}

// Watches the provided clause.
func (s *Solver) watchClause(c *Clause) {
	if c.PseudoBoolean() {
		s.watchPB(c)
	} else if c.Cardinality() > 1 {
		for i := 0; i < c.Cardinality()+1; i++ {
			lit := c.Get(i)
			neg := lit.Negation()
			s.wl.wlistPb[neg] = append(s.wl.wlistPb[neg], c)
		}
	} else if c.Len() == 2 {
		first := c.First()
		second := c.Second()
		neg0 := first.Negation()
		neg1 := second.Negation()
		s.wl.wlistBin[neg0] = append(s.wl.wlistBin[neg0], watcher{clause: c, other: second})
		s.wl.wlistBin[neg1] = append(s.wl.wlistBin[neg1], watcher{clause: c, other: first})
	} else { // Regular, propositional clause
		first := c.First()
		second := c.Second()
		neg0 := first.Negation()
		neg1 := second.Negation()
		s.wl.wlist[neg0] = append(s.wl.wlist[neg0], watcher{clause: c, other: second})
		s.wl.wlist[neg1] = append(s.wl.wlist[neg1], watcher{clause: c, other: first})
	}
}

func (s *Solver) watchPB(c *Clause) {
	goal := c.Weight(0) + c.Cardinality() // We'll keep watching vars until the max weightat least reaches this value
	sum := 0
	i := 0
	for sum < goal {
		lit := c.Get(i)
		neg := lit.Negation()
		s.wl.wlistPb[neg] = append(s.wl.wlistPb[neg], c)
		c.pbData.watched[i] = true
		sum += c.Weight(i)
		i++
	}
}

// unwatch the given learned clause.
// NOTE: since it is only called when c.lbd() > 2, we know for sure
// that c is not a binary clause.
// We also know for sure this is a propositional clause, since only those are learned.
func (s *Solver) unwatchClause(c *Clause) {
	for i := 0; i < 2; i++ {
		neg := c.Get(i).Negation()
		j := 0
		length := len(s.wl.wlist[neg])
		// We're looking for the index of the clause.
		// This will panic if c is not in wlist[neg], but this shouldn't happen.
		for s.wl.wlist[neg][j].clause != c {
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
	if s.Certified {
		if s.CertChan == nil {
			fmt.Printf("%s\n", c.CNF())
		} else {
			s.CertChan <- c.CNF()
		}
	}
}

// Adds the given unit literal to the model at the top level.
func (s *Solver) addLearnedUnit(unit Lit) {

	s.model[unit.Var()] = lvlToSignedLvl(unit, 1)
	if s.Certified {
		if s.CertChan == nil {
			fmt.Printf("%d 0\n", unit.Int())
		} else {
			s.CertChan <- fmt.Sprintf("%d 0", unit.Int())
		}
	}
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

// Propagates literals in the trail starting from the ptrth, and returns a conflict clause, or nil if none arose.
func (s *Solver) propagate(ptr int, lvl decLevel) *Clause {
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
		if confl := s.simplifyPropClauses(lit, lvl); confl != nil {
			return confl
		}
		for _, c := range s.wl.wlistPb[lit] {
			if c.PseudoBoolean() {
				if !s.simplifyPseudoBool(c, lvl) {
					return c
				}
			} else {
				if !s.simplifyCardConstr(c, lvl) {
					return c
				}
			}
		}
		ptr++
	}
	// No unsat clause was met
	return nil
}

// Unifies the given literal and returns a conflict clause, or nil if no conflict arose.
func (s *Solver) unifyLiteral(lit Lit, lvl decLevel) *Clause {
	s.model[lit.Var()] = lvlToSignedLvl(lit, lvl)
	s.trail = append(s.trail, lit)
	return s.propagate(len(s.trail)-1, lvl)
}

func (s *Solver) propagateUnit(c *Clause, lvl decLevel, unit Lit) {
	v := unit.Var()
	s.reason[v] = c
	c.lock()
	s.model[v] = lvlToSignedLvl(unit, lvl)
	s.trail = append(s.trail, unit)
}

func (s *Solver) simplifyPropClauses(lit Lit, lvl decLevel) *Clause {
	wl := s.wl.wlist[lit]
	j := 0
	for i, w := range wl {
		if s.litStatus(w.other) == Sat { // blocking literal is SAT? Don't explore clause!
			wl[j] = w
			j++
			continue
		}
		c := w.clause
		// make sure c.Second() is lit
		if c.First() == lit.Negation() {
			c.swap(0, 1)
		}
		w2 := watcher{clause: c, other: c.First()}
		firstStatus := s.litStatus(c.First())
		if firstStatus == Sat { // Clause is already sat
			wl[j] = w2
			j++
		} else {
			found := false
			for k := 2; k < c.Len(); k++ {
				if litK := c.Get(k); s.litStatus(litK) != Unsat {
					c.swap(1, k)
					neg := litK.Negation()
					s.wl.wlist[neg] = append(s.wl.wlist[neg], w2)
					found = true
					break
				}
			}
			if !found { // No free lit found: unit propagation or UNSAT
				wl[j] = w2
				j++
				if firstStatus == Unsat {
					copy(wl[j:], wl[i+1:]) // Copy remaining clauses
					s.wl.wlist[lit] = wl[:len(wl)-((i+1)-j)]
					return c
				}
				s.propagateUnit(c, lvl, c.First())
			}
		}
	}
	s.wl.wlist[lit] = wl[:j]
	return nil
}

// simplifyCardConstr simplifies a constraint of cardinality > 1, but with all weights = 1.
// returns false iff the clause cannot be satisfied.
func (s *Solver) simplifyCardConstr(clause *Clause, lvl decLevel) bool {
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
		case Sat:
			nbTrue++
			if nbTrue == card {
				return true
			}
		case Unsat:
			nbFalse++
			if length-nbFalse < card {
				return false
			}
		}
		if nbUnb+nbTrue > card {
			break
		}
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
		ni := &s.wl.wlistPb[clause.Get(i).Negation()]
		nj := &s.wl.wlistPb[clause.Get(j).Negation()]
		clause.swap(i, j)
		*ni = removeFrom(*ni, clause)
		*nj = append(*nj, clause)
		i++
		j++
	}
}

// slack Sumreturns slack value for c and whether the clause is already sat or not.
// The slack value is defined as sum of weights - cardinality - sum of weights of falsified lits.
// It can be negative, meaning the whole constraint is falsified.
// If it's 0 or above, it means all literals with a weight >= slack must be propagated.
// If the clause is already satisfied, the slack value shall not be used.
// This is mostly useful for PB constraints.
func (s *Solver) slackSum(c *Clause) (slack int, sat bool) {
	card := c.Cardinality()
	slack = -card
	sum := 0
	for i, w := range c.pbData.weights {
		status := s.litStatus(c.Get(i))
		switch status {
		case Indet:
			slack += w
		case Sat:
			slack += w
			sum += w
			if sum >= card {
				return slack, true
			}
		}
	}
	return slack, false
}

// propagateAll propagates all unbounded literals from c as unit literals
func (s *Solver) propagateAll(c *Clause, lvl decLevel) {
	for i := 0; i < c.Len(); i++ {
		if lit := c.Get(i); s.litStatus(lit) == Indet {
			s.propagateUnit(c, lvl, lit)
		}
	}
}

func (s *Solver) simplifyPseudoBool(clause *Clause, lvl decLevel) bool {
	foundUnit := true
	for foundUnit {
		slack, sat := s.slackSum(clause)
		if sat {
			return true
		}
		if slack < 0 {
			return false
		}
		if slack == 0 {
			s.propagateAll(clause, lvl)
			return true
		}
		foundUnit = false
		for i := 0; i < clause.Len(); i++ {
			lit := clause.Get(i)
			if s.litStatus(lit) == Indet && clause.Weight(i) > slack { // lit will be propagated
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
			if clause.pbData.watched[i] {
				ni := &s.wl.wlistPb[lit.Negation()]
				*ni = removeFrom(*ni, clause)
				clause.pbData.watched[i] = false
			}
		} else {
			weightWatched += clause.Weight(i)
			if !clause.pbData.watched[i] {
				ni := &s.wl.wlistPb[lit.Negation()]
				*ni = append(*ni, clause)
				clause.pbData.watched[i] = true
			}
		}
		i++
	}
	// If there are some more watched literals, they are now useless
	for i := i; i < clause.Len(); i++ {
		if clause.pbData.watched[i] {
			ni := &s.wl.wlistPb[clause.Get(i).Negation()]
			*ni = removeFrom(*ni, clause)
			clause.pbData.watched[i] = false
		}
	}
}
