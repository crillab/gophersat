package solver

// Describes basic types and constants that are used in the solver

// Status is the status of a given problem or clause at a given moment.
type Status byte

const (
	// Indet means the problem is not proven sat or unsat yet.
	Indet = Status(iota)
	// Sat means the problem or clause is satisfied.
	Sat
	// Unsat means the problem or clause is unsatisfied.
	Unsat
	// Unit is a constant meaning the clause contains only one unassigned literal.
	Unit
	// Many is a constant meaning the clause contains at least 2 unassigned literals.
	Many
)

func (s Status) String() string {
	switch s {
	case Indet:
		return "INDETERMINATE"
	case Sat:
		return "SAT"
	case Unsat:
		return "UNSAT"
	case Unit:
		return "UNIT"
	case Many:
		return "MANY"
	default:
		panic("invalid status")
	}
}

// Var start at 0 ; thus the CNF variable 1 is encoded as the Var 0.
type Var int32

// Lit start at 0 and are positive ; the sign is the last bit.
// Thus the CNF literal -3 is encoded as 2 * (3-1) + 1 = 5.
type Lit int32

// IntToLit converts a CNF literal to a Lit.
func IntToLit(i int) Lit {
	if i < 0 {
		return Lit(2*(-i-1) + 1)
	}
	return Lit(2 * (i - 1))
}

// IntToVar converts a CNF variable to a Var.
func IntToVar(i int32) Var {
	return Var(i - 1)
}

// Lit returns the positive Lit associated to v.
func (v Var) Lit() Lit {
	return Lit(v * 2)
}

// SignedLit returns the Lit associated to v, negated if 'signed', positive else.
func (v Var) SignedLit(signed bool) Lit {
	if signed {
		return Lit(v*2) + 1
	}
	return Lit(v * 2)
}

// Var returns the variable of l.
func (l Lit) Var() Var {
	return Var(l / 2)
}

// Int returns the equivalent CNF literal.
func (l Lit) Int() int32 {
	sign := l&1 == 1
	res := int32(l/2 + 1)
	if sign {
		return -res
	}
	return res
}

// IsPositive is true iff l is > 0
func (l Lit) IsPositive() bool {
	return l%2 == 0
}

// Negation returns -l, i.e the positive version of l if it is negative,
// or the negative version otherwise.
func (l Lit) Negation() Lit {
	return l ^ 1
}
