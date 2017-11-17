package solver

// A CardConstr is a cardinality constraint, i.e a set of literals (represented with integer variables) associated with a minimal number of literals that must be true.
// A propositional clause (i.e a disjunction of literals) is a cardinality constraint with a minimal cardinality of 1.
type CardConstr struct {
	Lits    []int
	AtLeast int
}

// AtLeast1 returns a cardinality constraint stating that at least one of the given lits must be true.
// This is the equivalent of a propositional clause.
func AtLeast1(lits ...int) CardConstr {
	return CardConstr{Lits: lits, AtLeast: 1}
}

// AtMost1 returns a cardinality constraint stating that at most one of the given lits can be true.
func AtMost1(lits ...int) CardConstr {
	for i, lit := range lits {
		lits[i] = -lit
	}
	return CardConstr{Lits: lits, AtLeast: len(lits) - 1}
}

// Exactly1 returns two cardinality constraints stating that exactly one of the given lits must be true.
func Exactly1(lits ...int) []CardConstr {
	return []CardConstr{AtLeast1(lits...), AtMost1(lits...)}
}
