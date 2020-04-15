package solver

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

const (
	initNbMaxClauses  = 2000  // Maximum # of learned clauses, at first.
	incrNbMaxClauses  = 300   // By how much # of learned clauses is incremented at each conflict.
	incrPostponeNbMax = 1000  // By how much # of learned is increased when lots of good clauses are currently learned.
	clauseDecay       = 0.999 // By how much clauses bumping decays over time.
	defaultVarDecay   = 0.8   // On each var decay, how much the varInc should be decayed at startup
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
	Verbose     bool        // Indicates whether the solver should display information during solving or not. False by default
	Certified   bool        // Indicates whether a certificate should be generated during solving or not, using the RUP notation. This is useful to prove UNSAT instances. False by default.
	CertChan    chan string // Indicates where to write the certificate. If Certified is true but CertChan is nil, the certificate will be written on stdout.
	nbVars      int
	status      Status
	wl          watcherList
	trail       []Lit     // Current assignment stack
	model       Model     // 0 means unbound, other value is a binding
	lastModel   Model     // Placeholder for last model found, useful when looking for several models
	activity    []float64 // How often each var is involved in conflicts
	polarity    []bool    // Preferred sign for each var
	assumptions []bool    // True iff the var's binding is assumed
	// For each var, clause considered when it was unified
	// If the var is not bound yet, or if it was bound by a decision, value is nil.
	reason          []*Clause
	varQueue        queue
	varInc          float64 // On each var bump, how big the increment should be
	clauseInc       float32 // On each var bump, how big the increment should be
	lbdStats        lbdStats
	Stats           Stats   // Statistics about the solving process.
	minLits         []Lit   // Lits to minimize if the problem was an optimization problem.
	minWeights      []int   // Weight of each lit to minimize if the problem was an optimization problem.
	hypothesis      []Lit   // Literals that are, ideally, true. Useful when trying to minimize a function.
	localNbRestarts int     // How many restarts since Solve() was called?
	varDecay        float64 // On each var decay, how much the varInc should be decayed
	trailBuf        []int   // A buffer while cleaning bindings
}

// New makes a solver, given a number of variables and a set of clauses.
// nbVars should be consistent with the content of clauses, i.e.
// the biggest variable in clauses should be >= nbVars.
func New(problem *Problem) *Solver {
	if problem.Status == Unsat {
		return &Solver{status: Unsat}
	}
	nbVars := problem.NbVars

	trailCap := nbVars
	if len(problem.Units) > trailCap {
		trailCap = len(problem.Units)
	}

	s := &Solver{
		nbVars:      nbVars,
		status:      problem.Status,
		trail:       make([]Lit, len(problem.Units), trailCap),
		model:       problem.Model,
		activity:    make([]float64, nbVars),
		polarity:    make([]bool, nbVars),
		assumptions: make([]bool, nbVars),
		reason:      make([]*Clause, nbVars),
		varInc:      1.0,
		clauseInc:   1.0,
		minLits:     problem.minLits,
		minWeights:  problem.minWeights,
		varDecay:    defaultVarDecay,
		trailBuf:    make([]int, nbVars),
	}
	s.resetOptimPolarity()
	s.initOptimActivity()
	s.initWatcherList(problem.Clauses)
	s.varQueue = newQueue(s.activity)
	for i, lit := range problem.Units {
		if lit.IsPositive() {
			s.model[lit.Var()] = 1
		} else {
			s.model[lit.Var()] = -1
		}
		s.trail[i] = lit
	}
	return s
}

// sets initial activity for optimization variables, if any.
func (s *Solver) initOptimActivity() {
	for i, lit := range s.minLits {
		w := 1
		if s.minWeights != nil {
			w = s.minWeights[i]
		}
		s.activity[lit.Var()] += float64(w)
	}
}

// resets polarity of optimization lits so that they are negated by default.
func (s *Solver) resetOptimPolarity() {
	if s.minLits != nil {
		for _, lit := range s.minLits {
			s.polarity[lit.Var()] = !lit.IsPositive() // Try to make lits from the optimization clause false
		}
	}
}

// Optim returns true iff the underlying problem is an optimization problem (rather than a satisfaction one).
func (s *Solver) Optim() bool {
	return s.minLits != nil
}

// OutputModel outputs the model for the problem on stdout.
func (s *Solver) OutputModel() {
	if s.status == Sat || s.lastModel != nil {
		fmt.Printf("s SATISFIABLE\nv ")
		model := s.model
		if s.lastModel != nil {
			model = s.lastModel
		}
		for i, val := range model {
			if val < 0 {
				fmt.Printf("%d ", -i-1)
			} else {
				fmt.Printf("%d ", i+1)
			}
		}
		fmt.Printf("\n")
	} else if s.status == Unsat {
		fmt.Printf("s UNSATISFIABLE\n")
	} else {
		fmt.Printf("s INDETERMINATE\n")
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
	s.varInc *= 1 / s.varDecay
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
			for _, c2 := range s.wl.learned {
				c2.activity *= 1e-30
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

// Reinitializes bindings (both model & reason) for all variables bound at a decLevel >= lvl.
// TODO: check this method as it has a weird behavior regarding performance.
// TODO: clean-up commented-out code and understand underlying performance pattern.
func (s *Solver) cleanupBindings(lvl decLevel) {
	i := 0
	for i < len(s.trail) && abs(s.model[s.trail[i].Var()]) <= lvl {
		i++
	}
	/*
		for j := len(s.trail) - 1; j >= i; j-- {
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
	*/
	toInsert := s.trailBuf[:0] // make([]int, 0, len(s.trail)-i)
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
			toInsert = append(toInsert, int(v))
			s.varQueue.insert(int(v))
		}
	}
	s.trail = s.trail[:i]
	for i := len(toInsert) - 1; i >= 0; i-- {
		s.varQueue.insert(toInsert[i])
	}
	/*for i := len(s.trail) - 1; i >= 0; i-- {
		lit := s.trail[i]
		v := lit.Var()
		if abs(s.model[v]) <= lvl { // All lits in trail before here must keep their status.
			s.trail = s.trail[:i+1]
			break
		}
		s.model[v] = 0
		if s.reason[v] != nil {
			s.reason[v].unlock()
			s.reason[v] = nil
		}
		s.polarity[v] = lit.IsPositive()
		if !s.varQueue.contains(int(v)) {
			s.varQueue.insert(int(v))
		}
	}*/
	s.resetOptimPolarity()
}

// Given the last learnt clause and the levels at which vars were bound,
// Returns the level to bt to and the literal to bind
func backtrackData(c *Clause, model []decLevel) (btLevel decLevel, lit Lit) {
	btLevel = abs(model[c.Get(1).Var()])
	return btLevel, c.Get(0)
}

func (s *Solver) rebuildOrderHeap() {
	ints := make([]int, s.nbVars)
	for v := 0; v < s.nbVars; v++ {
		if s.model[v] == 0 {
			ints = append(ints, int(v))
		}
	}
	s.varQueue.build(ints)
}

// propagate binds the given lit, propagates it and searches for a solution,
// until it is found or a restart is needed.
func (s *Solver) propagateAndSearch(lit Lit, lvl decLevel) Status {
	for lit != -1 {
		// log.Printf("picked %d at lvl %d", lit.Int(), lvl)
		if conflict := s.unifyLiteral(lit, lvl); conflict == nil { // Pick new branch or restart
			if s.lbdStats.mustRestart() {
				s.lbdStats.clear()
				s.cleanupBindings(1)
				return Indet
			}
			if s.Stats.NbConflicts >= s.wl.idxReduce*s.wl.nbMax {
				s.wl.idxReduce = s.Stats.NbConflicts/s.wl.nbMax + 1
				s.reduceLearned()
				s.bumpNbMax()
			}
			lvl++
			lit = s.chooseLit()
		} else { // Deal with conflict
			s.Stats.NbConflicts++
			if s.Stats.NbConflicts%5000 == 0 && s.varDecay < 0.95 {
				s.varDecay += 0.01
			}
			s.lbdStats.addConflict(len(s.trail))
			learnt, unit := s.learnClause(conflict, lvl)
			if learnt == nil { // Unit clause was learned: this lit is known for sure
				if unit == -1 || (abs(s.model[unit.Var()]) == 1 && s.litStatus(unit) == Unsat) { // Top-level conflict
					return s.setUnsat()
				}
				s.Stats.NbUnitLearned++
				s.lbdStats.addLbd(1)
				s.cleanupBindings(1)
				s.addLearnedUnit(unit)
				s.model[unit.Var()] = lvlToSignedLvl(unit, 1)
				if conflict = s.unifyLiteral(unit, 1); conflict != nil { // top-level conflict
					return s.setUnsat()
				}
				s.rebuildOrderHeap()
				lit = s.chooseLit()
				lvl = 2
			} else {
				if learnt.Len() == 2 {
					s.Stats.NbBinaryLearned++
				}
				s.Stats.NbLearned++
				s.lbdStats.addLbd(learnt.lbd())
				s.addLearned(learnt)
				lvl, lit = backtrackData(learnt, s.model)
				s.cleanupBindings(lvl)
				s.reason[lit.Var()] = learnt
				learnt.lock()
			}
		}
	}
	return Sat
}

// Sets the status to unsat and do cleanup tasks.
func (s *Solver) setUnsat() Status {
	if s.Certified {
		if s.CertChan == nil {
			fmt.Printf("0\n")
		} else {
			s.CertChan <- "0"
		}
	}
	s.status = Unsat
	return Unsat
}

// Searches until a restart is needed.
func (s *Solver) search() Status {
	s.localNbRestarts++
	lvl := decLevel(2) // Level starts at 2, for implementation reasons : 1 is for top-level bindings; 0 means "no level assigned yet"
	s.status = s.propagateAndSearch(s.chooseLit(), lvl)
	return s.status
}

// Solve solves the problem associated with the solver and returns the appropriate status.
func (s *Solver) Solve() Status {
	if s.status == Unsat {
		return s.status
	}
	s.status = Indet
	//s.lbdStats.clear()
	s.localNbRestarts = 0
	var end chan struct{}
	if s.Verbose {
		end = make(chan struct{})
		defer close(end)
		go func() { // Function displaying stats during resolution
			fmt.Printf("c ======================================================================================\n")
			fmt.Printf("c | Restarts |  Conflicts  |  Learned  |  Deleted  | Del%% | Reduce |   Units learned   |\n")
			fmt.Printf("c ======================================================================================\n")
			ticker := time.NewTicker(3 * time.Second)
			defer ticker.Stop()
			for { // There might be concurrent access in a few places but this is okay since we are very conservative and don't modify state.
				select {
				case <-ticker.C:
				case <-end:
					return
				}
				if s.status == Indet {
					iter := s.Stats.NbRestarts + 1
					nbConfl := s.Stats.NbConflicts
					nbReduce := s.wl.idxReduce - 1
					nbLearned := len(s.wl.learned)
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
			s.rebuildOrderHeap()
		}
	}
	if s.status == Sat {
		s.lastModel = make(Model, len(s.model))
		copy(s.lastModel, s.model)
	}
	if s.Verbose {
		end <- struct{}{}
		fmt.Printf("c ======================================================================================\n")
	}
	return s.status
}

// Assume adds unit literals to the solver.
// This is useful when calling the solver several times, e.g to keep it "hot" while removing clauses.
func (s *Solver) Assume(lits []Lit) Status {
	s.cleanupBindings(0)
	s.trail = s.trail[:0]
	s.assumptions = make([]bool, s.nbVars)
	for _, lit := range lits {
		s.addLearnedUnit(lit)
		s.assumptions[lit.Var()] = true
		s.trail = append(s.trail, lit)
	}
	s.status = Indet
	if confl := s.propagate(0, 1); confl != nil {
		// Conflict after unit propagation
		s.status = Unsat
		return s.status
	}
	return s.status
}

// Enumerate returns the total number of models for the given problems.
// if "models" is non-nil, it will write models on it as soon as it discovers them.
// models will be closed at the end of the method.
func (s *Solver) Enumerate(models chan []bool, stop chan struct{}) int {
	if models != nil {
		defer close(models)
	}
	s.lastModel = make(Model, len(s.model))
	nb := 0
	lit := s.chooseLit()
	var lvl decLevel
	for s.status != Unsat {
		for s.status == Indet {
			s.search()
			if s.status == Indet {
				s.Stats.NbRestarts++
			}
		}
		if s.status == Sat {
			copy(s.lastModel, s.model)
			if models != nil {
				nb += s.addCurrentModels(models)
			} else {
				nb += s.countCurrentModels()
			}
			s.status = Indet
			lits := s.decisionLits()
			switch len(lits) {
			case 0:
				s.status = Unsat
			case 1:
				s.propagateUnits(lits)
			default:
				c := NewClause(lits)
				s.appendClause(c)
				lit = lits[len(lits)-1]
				v := lit.Var()
				lvl = abs(s.model[v]) - 1
				s.cleanupBindings(lvl)
				s.reason[v] = c // Must do it here because it won't be made by propagateAndSearch
				s.propagateAndSearch(lit, lvl)
			}
		}
	}
	return nb
}

// CountModels returns the total number of models for the given problem.
func (s *Solver) CountModels() int {
	var end chan struct{}
	if s.Verbose {
		end = make(chan struct{})
		defer close(end)
		go func() { // Function displaying stats during resolution
			fmt.Printf("c ======================================================================================\n")
			fmt.Printf("c | Restarts |  Conflicts  |  Learned  |  Deleted  | Del%% | Reduce |   Units learned   |\n")
			fmt.Printf("c ======================================================================================\n")
			ticker := time.NewTicker(3 * time.Second)
			defer ticker.Stop()
			for { // There might be concurrent access in a few places but this is okay since we are very conservative and don't modify state.
				select {
				case <-ticker.C:
				case <-end:
					return
				}
				if s.status == Indet {
					iter := s.Stats.NbRestarts + 1
					nbConfl := s.Stats.NbConflicts
					nbReduce := s.wl.idxReduce - 1
					nbLearned := len(s.wl.learned)
					nbDel := s.Stats.NbDeleted
					pctDel := int(100 * float64(nbDel) / float64(s.Stats.NbLearned))
					nbUnit := s.Stats.NbUnitLearned
					fmt.Printf("c | %8d | %11d | %9d | %9d | %3d%% | %6d | %8d/%8d |\n", iter, nbConfl, nbLearned, nbDel, pctDel, nbReduce, nbUnit, s.nbVars)
				}
			}
		}()
	}
	nb := 0
	lit := s.chooseLit()
	var lvl decLevel
	for s.status != Unsat {
		for s.status == Indet {
			s.search()
			if s.status == Indet {
				s.Stats.NbRestarts++
			}
		}
		if s.status == Sat {
			s.lastModel = s.model
			nb += s.countCurrentModels()
			if s.Verbose {
				fmt.Printf("c found %d model(s)\n", nb)
			}
			s.status = Indet
			lits := s.decisionLits()
			switch len(lits) {
			case 0:
				s.status = Unsat
			case 1:
				s.propagateUnits(lits)
			default:
				c := NewClause(lits)
				s.appendClause(c)
				lit = lits[len(lits)-1]
				v := lit.Var()
				lvl = abs(s.model[v]) - 1
				s.cleanupBindings(lvl)
				s.reason[v] = c // Must do it here because it won't be made by propagateAndSearch
				s.propagateAndSearch(lit, lvl)
			}
		}
	}
	if s.Verbose {
		end <- struct{}{}
		fmt.Printf("c ======================================================================================\n")
	}
	return nb
}

// decisionLits returns the negation of all decision values once a model was found, ordered by decision levels.
// This will allow for searching other models.
func (s *Solver) decisionLits() []Lit {
	lastLit := s.trail[len(s.trail)-1]
	lvls := abs(s.model[lastLit.Var()])
	lits := make([]Lit, lvls-1)
	for i, r := range s.reason {
		if lvl := abs(s.model[i]); r == nil && lvl > 1 {
			if s.model[i] < 0 {
				// lvl-2 : levels beside unit clauses start at 2, not 0 or 1!
				lits[lvl-2] = IntToLit(int32(i + 1))
			} else {
				lits[lvl-2] = IntToLit(int32(-i - 1))
			}
		}
	}
	return lits
}

func (s *Solver) propagateUnits(units []Lit) {
	for _, unit := range units {
		s.lbdStats.addLbd(1)
		s.Stats.NbUnitLearned++
		s.cleanupBindings(1)
		s.model[unit.Var()] = lvlToSignedLvl(unit, 1)
		if s.unifyLiteral(unit, 1) != nil {
			s.status = Unsat
			return
		}
		s.rebuildOrderHeap()
	}
}

// PBString returns a representation of the solver's state as a pseudo-boolean problem.
func (s *Solver) PBString() string {
	meta := fmt.Sprintf("* #variable= %d #constraint= %d #learned= %d\n", s.nbVars, len(s.wl.pbClauses), len(s.wl.learned))
	minLine := ""
	if s.minLits != nil {
		terms := make([]string, len(s.minLits))
		for i, lit := range s.minLits {
			weight := 1
			if s.minWeights != nil {
				weight = s.minWeights[i]
			}
			val := lit.Int()
			sign := ""
			if val < 0 {
				val = -val
				sign = "~"
			}
			terms[i] = fmt.Sprintf("%d %sx%d", weight, sign, val)
		}
		minLine = fmt.Sprintf("min: %s ;\n", strings.Join(terms, " +"))
	}
	clauses := make([]string, len(s.wl.pbClauses)+len(s.wl.learned))
	for i, c := range s.wl.pbClauses {
		clauses[i] = c.PBString()
	}
	for i, c := range s.wl.learned {
		clauses[i+len(s.wl.pbClauses)] = c.PBString()
	}
	for i := 0; i < len(s.model); i++ {
		if s.model[i] == 1 {
			clauses = append(clauses, fmt.Sprintf("1 x%d = 1 ;", i+1))
		} else if s.model[i] == -1 {
			clauses = append(clauses, fmt.Sprintf("1 x%d = 0 ;", i+1))
		}
	}
	return meta + minLine + strings.Join(clauses, "\n")
}

// AppendClause appends a new clause to the set of clauses.
// This is not a learned clause, but a clause that is part of the problem added afterwards (during model counting, for instance).
func (s *Solver) AppendClause(clause *Clause) {
	s.cleanupBindings(1)
	card := clause.Cardinality()
	minW := 0
	maxW := 0
	i := 0
	for i < clause.Len() {
		lit := clause.Get(i)
		switch s.litStatus(lit) {
		case Sat:
			w := clause.Weight(i)
			minW += w
			maxW += w
			clause.removeLit(i)
			clause.updateCardinality(-w)
		case Unsat:
			clause.removeLit(i)
		default:
			maxW += clause.Weight(i)
			i++
		}
	}
	if minW >= card { // clause is already sat
		return
	}
	if maxW < card { // clause cannot be satisfied
		s.status = Unsat
		return
	}
	if maxW == card { // Unit
		s.propagateUnits(clause.lits)
	} else {
		s.appendClause(clause)
	}
}

// Model returns a slice that associates, to each variable, its binding.
// If s's status is not Sat, the method will panic.
func (s *Solver) Model() []bool {
	if s.lastModel == nil {
		panic("cannot call Model() from a non-Sat solver")
	}
	res := make([]bool, s.nbVars)
	for i, lvl := range s.lastModel {
		res[i] = lvl > 0
	}
	return res
}

// addCurrentModels is called when a model was found.
// It returns the total number of models from this point, and sends all models on ch.
// The number can be different of 1 if there are unbound variables.
// For instance, if there are 4 variables in the problem and only 1, 3 and 4 are bound,
// there are actually 2 models currently: one with 2 set to true, the other with 2 set to false.
func (s *Solver) addCurrentModels(ch chan []bool) int {
	unbound := make([]int, 0, s.nbVars) // indices of unbound variables
	var nb uint64 = 1                   // total number of models found
	model := make([]bool, s.nbVars)     // partial model
	for i, lvl := range s.lastModel {
		if lvl == 0 {
			unbound = append(unbound, i)
			nb *= 2
		} else {
			model[i] = lvl > 0
		}
	}
	for i := uint64(0); i < nb; i++ {
		for j := range unbound {
			mask := uint64(1 << j)
			cur := i & mask
			idx := unbound[j]
			model[idx] = cur != 0
		}
		model2 := make([]bool, len(model))
		copy(model2, model)
		ch <- model2
	}
	return int(nb)
}

// countCurrentModels is called when a model was found.
// It returns the total number of models from this point.
// The number can be different of 1 if there are unbound variables.
// For instance, if there are 4 variables in the problem and only 1, 3 and 4 are bound,
// there are actually 2 models currently: one with 2 set to true, the other with 2 set to false.
func (s *Solver) countCurrentModels() int {
	var nb uint64 = 1 // total number of models found
	for _, lvl := range s.lastModel {
		if lvl == 0 {
			nb *= 2
		}
	}
	return int(nb)
}

// Optimal returns the optimal solution, if any.
// If results is non-nil, all solutions will be written to it.
// In any case, results will be closed at the end of the call.
func (s *Solver) Optimal(results chan Result, stop chan struct{}) (res Result) {
	if results != nil {
		defer close(results)
	}
	status := s.Solve()
	if status == Unsat { // Problem cannot be satisfied at all
		res.Status = Unsat
		if results != nil {
			results <- res
		}
		return res
	}
	if s.minLits == nil { // No optimization clause: this is a decision problem, solution is optimal
		s.lastModel = make(Model, len(s.model))
		copy(s.lastModel, s.model)
		res := Result{
			Status: Sat,
			Model:  s.Model(),
			Weight: 0,
		}
		if results != nil {
			results <- res
		}
		return res
	}
	maxCost := 0
	if s.minWeights == nil {
		maxCost = len(s.minLits)
	} else {
		for _, w := range s.minWeights {
			maxCost += w
		}
	}
	s.hypothesis = make([]Lit, len(s.minLits))
	for i, lit := range s.minLits {
		s.hypothesis[i] = lit.Negation()
	}
	weights := make([]int, len(s.minWeights))
	copy(weights, s.minWeights)
	sort.Sort(wLits{lits: s.hypothesis, weights: weights})
	s.lastModel = make(Model, len(s.model))
	var cost int
	for status == Sat {
		copy(s.lastModel, s.model) // Save this model: it might be the last one
		cost = 0
		for i, lit := range s.minLits {
			if s.model[lit.Var()] > 0 == lit.IsPositive() {
				if s.minWeights == nil {
					cost++
				} else {
					cost += s.minWeights[i]
				}
			}
		}
		res = Result{
			Status: Sat,
			Model:  s.Model(),
			Weight: cost,
		}
		if results != nil {
			results <- res
		}
		if cost == 0 {
			break
		}
		// Add a constraint incrementing current best cost
		lits2 := make([]Lit, len(s.minLits))
		weights2 := make([]int, len(s.minWeights))
		copy(lits2, s.hypothesis)
		copy(weights2, weights)
		s.AppendClause(NewPBClause(lits2, weights2, maxCost-cost+1))
		s.rebuildOrderHeap()
		status = s.Solve()
	}
	return res
}

// Minimize tries to find a model that minimizes the weight of the clause defined as the optimisation clause in the problem.
// If no model can be found, it will return a cost of -1.
// Otherwise, calling s.Model() afterwards will return the model that satisfy the formula, such that no other model with a smaller cost exists.
// If this function is called on a non-optimization problem, it will either return -1, or a cost of 0 associated with a
// satisfying model (ie any model is an optimal model).
func (s *Solver) Minimize() int {
	status := s.Solve()
	if status == Unsat { // Problem cannot be satisfied at all
		return -1
	}
	if s.minLits == nil { // No optimization clause: this is a decision problem, solution is optimal
		return 0
	}
	maxCost := 0
	if s.minWeights == nil {
		maxCost = len(s.minLits)
	} else {
		for _, w := range s.minWeights {
			maxCost += w
		}
	}
	s.hypothesis = make([]Lit, len(s.minLits))
	for i, lit := range s.minLits {
		s.hypothesis[i] = lit.Negation()
	}
	weights := make([]int, len(s.minWeights))
	copy(weights, s.minWeights)
	sort.Sort(wLits{lits: s.hypothesis, weights: weights})
	s.lastModel = make(Model, len(s.model))
	var cost int
	for status == Sat {
		copy(s.lastModel, s.model) // Save this model: it might be the last one
		cost = 0
		for i, lit := range s.minLits {
			if s.model[lit.Var()] > 0 == lit.IsPositive() {
				if s.minWeights == nil {
					cost++
				} else {
					cost += s.minWeights[i]
				}
			}
		}
		if cost == 0 {
			return 0
		}
		if s.Verbose {
			fmt.Printf("o %d\n", cost)
		}
		// Add a constraint incrementing current best cost
		lits2 := make([]Lit, len(s.minLits))
		weights2 := make([]int, len(s.minWeights))
		copy(lits2, s.hypothesis)
		copy(weights2, weights)
		s.AppendClause(NewPBClause(lits2, weights2, maxCost-cost+1))
		s.rebuildOrderHeap()
		status = s.Solve()
	}
	return cost
}

// functions to sort hypothesis for pseudo-boolean minimization clause.
type wLits struct {
	lits    []Lit
	weights []int
}

func (wl wLits) Len() int           { return len(wl.lits) }
func (wl wLits) Less(i, j int) bool { return wl.weights[i] > wl.weights[j] }

func (wl wLits) Swap(i, j int) {
	wl.lits[i], wl.lits[j] = wl.lits[j], wl.lits[i]
	wl.weights[i], wl.weights[j] = wl.weights[j], wl.weights[i]
}
