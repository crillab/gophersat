package solver

// a pbSet is a set representation of a PB constraint.
// It indicates, for each variable in the problem, what its weight is in the constraint.
// For instance, in a problem with 5 vars, the lits in constraint 5 x1 +3 ~x2 +2 x4+ x5 >= 6 will be encoded as
// []int{5, -3, 0, 2, 5}.
type pbSet struct {
	weights []int // The weight of each variable in the constraint, or 0 if the var isn't in the constraint.
	card    int
}

// pbSet converts c to the psSet structure.
// buffer is a buffer to store the weights. This is a parameter so as to avoid too frequent allocations.
func (s *Solver) pbSet(c *Clause, buffer []int) *pbSet {
	res := &pbSet{weights: buffer, card: c.Cardinality()}
	for i := range buffer { // Buffer has to be cleaned first
		buffer[i] = 0
	}
	for i := 0; i < c.Len(); i++ {
		lit := c.Get(i)
		v := lit.Var()
		w := c.Weight(i)
		if !lit.IsPositive() {
			w = -w
		}
		res.weights[v] = w
	}
	return res
}

func (pb *pbSet) clause() *Clause {
	lits := make([]Lit, 0, len(pb.weights))
	weights := make([]int, 0, len(pb.weights))
	for i, w := range pb.weights {
		if w == 0 {
			continue
		}
		idx := int32(i + 1)
		absW := w
		if w < 0 {
			absW = -w
			idx = -idx
		}
		lit := IntToLit(idx)
		lits = append(lits, lit)
		weights = append(weights, absW)
	}
	return NewPBClause(lits, weights, pb.card)
}

// clash will make pb1 and pb2 clash, and update the values in pb1.
// pb2 will be unmodified.
// There should be at least one variable whose weight becomes 0 in the process.
func (pb1 *pbSet) clash(s *Solver, pb2 *pbSet) {
	pb1.card += pb2.card
	for i, w1 := range pb1.weights {
		w2 := pb2.weights[i]
		pb1.weights[i] += w2
		if w1*w2 < 0 { // vars don't have the same polarity in both constraints
			pb1.card -= min(abs(w1), abs(w2))
		}
	}
}

// slack returns the slack of pb1 for the given decision level.
// It is defined as card - sum(falsified lits at decLevel <= lvl).
// If slack = 0, all unassigned lits shuld be propagated.
// If slack < 0, the constraint is falsified under current assignments.
func (s *Solver) slack(pb *pbSet, lvl decLevel) int {
	res := -pb.card
	for i, w := range pb.weights {
		if w == 0 {
			continue
		}
		l := IntToLit(int32(i + 1))
		absW := w
		if w < 0 {
			absW = -w
			l = l.Negation()
		}
		v := Var(i)
		litLvl := abs(s.model[v])
		if s.litStatus(l) != Unsat || litLvl > lvl {
			res += absW
		}
	}
	return res
}

// falsifies returns true iff lit's negation appears in pb.
// Only then should clashes happen.
func (pb *pbSet) falsifies(lit Lit) bool {
	v := lit.Var()
	w := pb.weights[v]
	if w == 0 {
		return false
	}
	return (w < 0) == lit.IsPositive()
}

// cuttingPlanes learns a new PB constraint using the cutting plane resolution system.
// This is usually less efficient than calling learnClause, but will dramatically improve
// efficiency in corner cases, such as the pigeonhole problem.
func (s *Solver) cuttingPlanes(confl *Clause, lvl decLevel) (learned *Clause, propagated []Lit, newLvl decLevel) {
	// fmt.Printf("conflict at level %d! Constraint is: %s, trail is %s\n", lvl, confl.PBString(), s.trailString())
	seen := make([]bool, s.nbVars) // Was the var seen in the resolution process, making it a candidate for bumping?
	s.clauseBumpActivity(confl)
	for _, lit := range confl.lits {
		seen[lit.Var()] = true
	}
	pb := s.pbSet(confl, s.pbSetBuf)
	ptr := len(s.trail) - 1
	for pb.onlyFalsified(s, ptr, lvl) < 0 {
		if lvl == 1 { // Top-level conflict: UNSAT
			return nil, nil, -1
		}
		lit := s.trail[ptr]
		for !pb.falsifies(lit) {
			if s.reason[lit.Var()] == nil {
				lvl--
			}
			s.model[lit.Var()] = 0
			// s.trail = s.trail[:ptr]
			ptr--
			lit = s.trail[ptr]
		}
		v := lit.Var()
		s.varBumpActivity(v) // RoundingSAT's strategy: eliminated variables are bumped twice
		pb.roundToOne(s, v, lvl)
		reason := s.reason[v]
		if reason == nil {
			lvl--
			continue
		}
		for _, lit := range reason.lits {
			seen[lit.Var()] = true
		}
		s.clauseBumpActivity(reason)
		pb2 := s.pbSet(reason, s.pbSetBuf2)
		pb2.roundToOne(s, v, lvl)
		pb.clash(s, pb2)
	}
	unit := pb.onlyFalsified(s, ptr, lvl).Negation()
	btLvl := pb.backtrackLevel(s, unit)
	pb.roundToOne(s, unit.Var(), lvl)
	for i := range seen {
		if seen[i] {
			s.varBumpActivity(Var(i))
		}
	}
	// s.varDecayActivity()
	// s.clauseDecayActivity()
	if propagated, learned, ok := pb.clause().SimplifyPB(); !ok {
		return nil, nil, -1
	} else if len(propagated) > 0 {
		return nil, propagated, 1
	} else {
		return learned, []Lit{unit}, btLvl
	}
}

func (pb *pbSet) backtrackLevel(s *Solver, falsified Lit) decLevel {
	v := falsified.Var()
	lvl := abs(s.model[v])
	maxLvl := decLevel(1)
	for i, w := range pb.weights {
		if w == 0 || Var(i) == v {
			continue
		}
		if lvlI := abs(s.model[i]); lvlI > maxLvl && lvlI != lvl {
			maxLvl = lvlI
		}
	}
	return maxLvl
}

// returns the only false literal at level lvl in pb, or -1 if no or several lits are false.
func (pb *pbSet) onlyFalsified(s *Solver, ptr int, lvl decLevel) Lit {
	var res Lit = -1
	for ptr >= 0 {
		lit := s.trail[ptr]
		if abs(s.model[lit.Var()]) != lvl { // We're out of lvl now: we're done
			return res
		}
		if pb.falsifies(lit) {
			if res != -1 {
				// We already had one: there are several non-false literals!
				return -1
			}
			res = lit
		}
		ptr--
	}
	return res
}

// roundToOne weakens pb by rounding the falsified literal ('locked'), as described in the RoundingSAT paper.
func (pb *pbSet) roundToOne(s *Solver, locked Var, lvl decLevel) {
	wi := abs(pb.weights[locked])
	if wi == 1 {
		return
	}
	for j, wj := range pb.weights {
		if wj == 0 {
			continue
		}
		assign := s.model[j]
		if wj%wi != 0 && (assign == 0 || ((assign > 0) == (wj > 0))) {
			// lit j isn't falsified: weaken constraint by removing it
			pb.weights[j] = 0
			pb.card -= abs(wj)
		}
	}
	pb.divideBy(wi)
}

// divideBy performs a division on the conflict var by applying rounded division on its weight.
// locked should be present in pb or a panic will ensue.
// Ideally, this function shouldn't be called when coeff == 1.
func (pb *pbSet) divideBy(coeff int) {
	for j, wj := range pb.weights {
		if wj == 0 {
			continue
		}
		if wj%coeff == 0 {
			pb.weights[j] = wj / coeff
		} else if wj > 0 {
			pb.weights[j] = (wj / coeff) + 1
		} else {
			pb.weights[j] = (wj / coeff) - 1
		}
	}
	if pb.card%coeff == 0 {
		pb.card = pb.card / coeff
	} else {
		pb.card = (pb.card / coeff) + 1
	}
}
