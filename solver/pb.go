package solver

// A PBConstr is a Pseudo-Boolean constraint.
type PBConstr struct {
	Lits    []int // List of literals, designed with integer values. A positive value means the literal is true, a negative one it is false.
	Weights []int // Weight of each lit from Lits. If nil, all lits == 1
	AtLeast int   // Sum of all lits must be at least this value
}

// WeightSum returns the sum of the weight of all terms.
func (c PBConstr) WeightSum() int {
	if c.Weights == nil { // All weights = 1
		return len(c.Lits)
	}
	res := 0
	for _, w := range c.Weights {
		res += w
	}
	return res
}

// Clause returns the clause associated with the given constraint.
func (c PBConstr) Clause() *Clause {
	lits := make([]Lit, len(c.Lits))
	for i, val := range c.Lits {
		lits[i] = IntToLit(val)
	}
	return NewPBClause(lits, c.Weights, c.AtLeast)
}

// PropClause returns a PB constraint equivalent to a propositional clause: at least one of the given
// literals must be true.
// It takes ownership of lits.
func PropClause(lits ...int) PBConstr {
	return PBConstr{Lits: lits, AtLeast: 1}
}

// AtLeast returns a PB constraint stating that at least n literals must be true.
// It takes ownership of lits.
func AtLeast(lits []int, n int) PBConstr {
	return PBConstr{Lits: lits, AtLeast: n}
}

// AtMost returns a PB constraint stating that at most n literals can be true.
// It takes ownership of lits.
func AtMost(lits []int, n int) PBConstr {
	for i := range lits {
		lits[i] = -lits[i]
	}
	return PBConstr{Lits: lits, AtLeast: len(lits) - n}
}

// GtEq returns a PB constraint stating that the sum of all literals multiplied by their weight
// must be at least n.
// Will panic if len(weights) != len(lits).
func GtEq(lits []int, weights []int, n int) PBConstr {
	if len(weights) != 0 && len(lits) != len(weights) {
		panic("not as many lits as weights")
	}
	for i := range weights {
		if weights[i] < 0 {
			weights[i] = -weights[i]
			n += weights[i]
			lits[i] = -lits[i]
		}
	}
	return PBConstr{Lits: lits, Weights: weights, AtLeast: n}
}

// LtEq returns a PB constraint stating that the sum of all literals multiplied by their weight
// must be at most n.
// Will panic if len(weights) != len(lits).
func LtEq(lits []int, weights []int, n int) PBConstr {
	sum := 0
	for i := range lits {
		lits[i] = -lits[i]
		sum += weights[i]
	}
	n = sum - n
	return GtEq(lits, weights, n)
}

// Eq returns a set of PB constraints stating that the sum of all literals multiplied by their weight
// must be exactly n.
// Will panic if len(weights) != len(lits).
func Eq(lits []int, weights []int, n int) []PBConstr {
	lits2 := make([]int, len(lits))
	weights2 := make([]int, len(weights))
	copy(lits2, lits)
	copy(weights2, weights)
	ge := GtEq(lits2, weights2, n)
	le := LtEq(lits, weights, n)
	var res []PBConstr
	if ge.AtLeast > 0 {
		res = append(res, ge)
	}
	if le.AtLeast > 0 {
		res = append(res, le)
	}
	return res
}
