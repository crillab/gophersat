package solver

import (
	"fmt"
	"time"
)

const (
	initNbMaxClauses  = 2000  // Maximum # of learned clauses, at first.
	incrNbMaxClauses  = 300   // By how much # of learned clauses is incremented at each conflict.
	incrPostponeNbMax = 1000  // By how much # of learned is increased when lots of good clauses are currently learned.
	clauseDecay       = 0.999 // By how much clauses bumping decays over time.
	varDecay          = 0.95  // On each var decay, how much the varInc should be decayed
)

// Stats are statistics about the resolution of the problem.
// They are provided for information purpose only.
type Stats struct {
	NbRestarts      int
	NbConflicts     int
	NbDecisions     int
	NbUnitLearned   int // How many unit clauses were learned
	NbBinaryLearned int // How many binary clauses were learned
	NbLearned       int // How many clauses were learned
	NbDeleted       int // How many clauses were deleted
}

// The level a decision was made.
// A negative value means "negative assignement at that level".
// A positive value means "positive assignment at that level".
type decLevel int

// A Model is a binding for several variables.
// It can be totally bound (i.e all vars have a true or false binding)
// or only partially (i.e some vars have no binding yet or their binding has no impact).
// Each var, in order, is associated with a binding. Binding are implemented as
// decision levels:
// - a 0 value means the variable is free,
// - a positive value means the variable was set to true at the given decLevel,
// - a negative value means the variable was set to false at the given decLevel.
type Model []decLevel

func (m Model) String() string {
	bound := make(map[int]decLevel)
	for i := range m {
		if m[i] != 0 {
			bound[i+1] = m[i]
		}
	}
	return fmt.Sprintf("%v", bound)
}

// A Solver solves a given problem. It is the main data structure.
type Solver struct {
	Verbose  bool // Indicates whether the solver should display information during solving or not. False by default
	nbVars   int
	status   Status
	wl       watcherList
	trail    []Lit     // Current assignment stack
	model    Model     // 0 means unbound, other value is a binding
	activity []float64 // How often each var is involved in conflicts
	polarity []bool    // Preferred sign for each var
	// For each var, clause considered when it was unified
	// If the var is not bound yet, or if it was bound by a decision, value is nil.
	reason    []*Clause
	varQueue  queue
	varInc    float64 // On each var bump, how big the increment should be
	clauseInc float32 // On each var bump, how big the increment should be
	lbdStats  lbdStats
	Stats     Stats // Statistics about the solving process.
}

// New makes a solver, given a number of variables and a set of clauses.
// nbVars should be consistent with the content of clauses, i.e.
// the biggest variable in clauses should be >= nbVars.
func New(problem *Problem) *Solver {
	nbVars := problem.NbVars
	s := &Solver{
		nbVars:    nbVars,
		status:    problem.Status,
		trail:     make([]Lit, 0, nbVars),
		model:     problem.Model,
		activity:  make([]float64, nbVars),
		polarity:  make([]bool, nbVars),
		reason:    make([]*Clause, nbVars),
		varInc:    1.0,
		clauseInc: 1.0,
	}
	s.initWatcherList(problem.Clauses)
	s.varQueue = newQueue(s.activity)
	for _, lit := range problem.Units {
		if lit.IsPositive() {
			s.model[lit.Var()] = 1
		} else {
			s.model[lit.Var()] = -1
		}
		s.trail = append(s.trail, lit)
	}
	return s
}

// OutputModel outputs the model for the problem on stdout.
func (s *Solver) OutputModel() {
	switch s.status {
	case Unsat:
		fmt.Printf("UNSATISFIABLE\n")
	case Indet:
		fmt.Printf("INDETERMINATE\n")
	case Sat:
		fmt.Printf("SATISFIABLE\n")
		for i, val := range s.model {
			if val < 0 {
				fmt.Printf("%d ", -i-1)
			} else {
				fmt.Printf("%d ", i+1)
			}
		}
		fmt.Printf("\n")
	}
}

// litStatus returns whether the literal is made true (Sat) or false (Unsat) by the
// current bindings, or if it is unbounded (Indet).
func (s *Solver) litStatus(l Lit) Status {
	assign := s.model[l.Var()]
	if assign == 0 {
		return Indet
	}
	if assign > 0 == l.IsPositive() {
		return Sat
	}
	return Unsat
}

func (s *Solver) varDecayActivity() {
	s.varInc *= 1 / varDecay
}

func (s *Solver) varBumpActivity(v Var) {
	s.activity[v] += s.varInc
	if s.activity[v] > 1e100 { // Rescaling is needed to avoid overflowing
		for i := range s.activity {
			s.activity[i] *= 1e-100
		}
		s.varInc *= 1e-100
	}
	if s.varQueue.contains(int(v)) {
		s.varQueue.decrease(int(v))
	}
}

// Decays each clause's activity
func (s *Solver) clauseDecayActivity() {
	s.clauseInc *= 1 / clauseDecay
}

// Bumps the given clause's activity.
func (s *Solver) clauseBumpActivity(c *Clause) {
	if c.Learned() {
		c.activity += s.clauseInc
		if c.activity > 1e30 { // Rescale to avoid overflow
			for i := s.wl.nbOriginal; i < len(s.wl.clauses); i++ {
				s.wl.clauses[i].activity *= 1e-30
			}
			s.clauseInc *= 1e-30
		}
	}
}

// Chooses an unbound literal to be tested, or -1
// if all the variables are already bound.
func (s *Solver) chooseLit() Lit {
	v := Var(-1)
	for v == -1 && !s.varQueue.empty() {
		if v2 := Var(s.varQueue.removeMin()); s.model[v2] == 0 { // Ignore already bound vars
			v = v2
		}
	}
	if v == -1 {
		return Lit(-1)
	}
	s.Stats.NbDecisions++
	return v.SignedLit(!s.polarity[v])
}

func abs(val decLevel) decLevel {
	if val < 0 {
		return -val
	}
	return val
}

// Reinitializes bindings (both model & reason) for all variables
// That have been bound at a decLevel >= lvl
func (s *Solver) cleanupBindings(lvl decLevel) {
	for i, lit := range s.trail {
		if abs(s.model[lit.Var()]) > lvl { // All lits in trail from here must be made unbound.
			for j := i; j < len(s.trail); j++ {
				lit2 := s.trail[j]
				v := lit2.Var()
				s.model[v] = 0
				if s.reason[v] != nil {
					s.reason[v].unlock()
					s.reason[v] = nil
				}
				s.polarity[v] = lit2.IsPositive()
				if !s.varQueue.contains(int(v)) {
					s.varQueue.insert(int(v))
				}
			}
			s.trail = s.trail[:i]
			break
		}
	}
}

// Given the last learnt clause and the levels at which vars were bound,
// Returns the level to bt to and the literal to bind
func backtrackData(c *Clause, model []decLevel) (btLevel decLevel, lit Lit) {
	btLevel = abs(model[c.Get(1).Var()])
	return btLevel, c.Get(0)
}

// Searches until a restart is needed.
func (s *Solver) search() Status {
	lvl := decLevel(2) // Level starts at 2, for implementation reasons : 1 is for top-level bindings; 0 means "no level assigned yet"
	lit := s.chooseLit()
	// log.Printf("LVL %d: chose %d", lvl, lit.Int())
	for lit != -1 {
		if conflict := s.unifyLiteral(lit, lvl); conflict == nil { // Pick new branch or restart
			if s.lbdStats.mustRestart() {
				// log.Printf("must restart")
				s.lbdStats.clear()
				s.cleanupBindings(1)
				return Indet
			}
			if s.Stats.NbConflicts >= s.wl.idxReduce*s.wl.nbMax {
				// log.Printf("musr reduce")
				s.wl.idxReduce = (s.Stats.NbConflicts / s.wl.nbMax) + 1
				s.reduceLearned()
				s.bumpNbMax()
			}
			lvl++
			lit = s.chooseLit()
			// log.Printf("LVL %d: chose %d", lvl, lit.Int())
		} else { // Deal with conflict
			s.Stats.NbConflicts++
			// log.Printf("conflict clause %v, %d conflicts now", conflict.CNF(), s.Stats.NbConflicts)
			learnt, unit := s.learnClause(conflict, lvl)
			if learnt == nil { // Unit clause was learned: this lit is known for sure
				// log.Printf("learned unit lit %d", unit.Int())
				s.lbdStats.add(1)
				s.Stats.NbUnitLearned++
				s.cleanupBindings(1)
				s.model[unit.Var()] = lvlToSignedLvl(unit, 1)
				s.trail = append(s.trail, unit)
				if conflict = s.unifyLiteral(unit, 1); conflict != nil { // top-level conflict
					s.status = Unsat
					return Unsat
				}
				lit = s.chooseLit()
			} else {
				// log.Printf("learned clause %v", learnt.CNF())
				if learnt.Len() == 2 {
					s.Stats.NbBinaryLearned++
				}
				s.Stats.NbLearned++
				s.lbdStats.add(learnt.lbd())
				s.addClause(learnt)
				lvl, lit = backtrackData(learnt, s.model)
				s.cleanupBindings(lvl)
				s.reason[lit.Var()] = learnt
				learnt.lock()
				// log.Printf("LVL %d: chose %d", lvl, lit.Int())
			}
		}
	}
	s.status = Sat
	return s.status
}

// Solve solves the problem associated with the solver and returns the appropriate status.
func (s *Solver) Solve() Status {
	if s.Verbose {
		go func() { // Function displaying stats during resolution
			fmt.Printf("c ======================================================================================\n")
			fmt.Printf("c | Restarts |  Conflicts  |  Learned  |  Deleted  | Del%% | Reduce |   Units learned   |\n")
			fmt.Printf("c ======================================================================================\n")
			for s.status == Indet { // There might be concurrent access in a few places but this is okay since we are very conservative and don't modify state.
				time.Sleep(3 * time.Second)
				if s.status == Indet {
					iter := s.Stats.NbRestarts + 1
					nbConfl := s.Stats.NbConflicts
					nbReduce := s.wl.idxReduce - 1
					nbLearned := s.wl.nbLearned
					nbDel := s.Stats.NbDeleted
					pctDel := int(100 * float64(nbDel) / float64(s.Stats.NbLearned))
					nbUnit := s.Stats.NbUnitLearned
					fmt.Printf("c | %8d | %11d | %9d | %9d | %3d%% | %6d | %8d/%8d |\n", iter, nbConfl, nbLearned, nbDel, pctDel, nbReduce, nbUnit, s.nbVars)
				}
			}
		}()
	}
	for s.status == Indet {
		s.search()
		if s.status == Indet {
			s.Stats.NbRestarts++
		}
	}
	if s.Verbose {
		fmt.Printf("c ======================================================================================\n")
	}
	return s.status
}

// Model returns a slice that associates, to each variable, its binding.
// If s's status is not Sat, will return an error.
func (s *Solver) Model() ([]bool, error) {
	if s.status != Sat {
		return nil, fmt.Errorf("cannot call Model() from a non-Sat solver")
	}
	res := make([]bool, s.nbVars)
	for i, lvl := range s.model {
		res[i] = lvl > 0
	}
	return res, nil
}
