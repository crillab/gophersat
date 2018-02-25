package maxsat

// A Lit is a potentially-negated boolean variable.
type Lit struct {
	Var     string
	Negated bool
}

// Var returns a new positive Lit whose var is named "name".
func Var(name string) Lit {
	return Lit{Var: name}
}

// Not returns a new negated Lit whose var is named "name".
func Not(name string) Lit {
	return Lit{Var: name, Negated: true}
}

func (l Lit) String() string {
	if l.Negated {
		return "¬" + l.Var
	}
	return l.Var
}

// Negation returns the logical negation of l.
func (l Lit) Negation() Lit {
	return Lit{Var: l.Var, Negated: !l.Negated}
}

// A Constr is a weighted pseudo-boolean constraint.
type Constr struct {
	Lits    []Lit // The list of lits in the problem.
	Coeffs  []int // The coefficients associated with each literals. If nil, all coeffs are supposed to be 1.
	AtLeast int   // Minimal cardinality for the constr to be satisfied.
	Weight  int   // The weight of the clause, or 0 for a hard clause.
}

// HardClause returns a propositional clause that must be satisfied.
func HardClause(lits ...Lit) Constr {
	return Constr{Lits: lits, AtLeast: 1}
}

// SoftClause returns an optional propositional clause.
func SoftClause(lits ...Lit) Constr {
	return Constr{Lits: lits, AtLeast: 1, Weight: 1}
}

// WeightedClause returns a weighted optional propositional clause.
func WeightedClause(lits []Lit, weight int) Constr {
	return Constr{Lits: lits, AtLeast: 1, Weight: weight}
}

// HardPBConstr returns a pseudo-boolean constraint that must be satisfied.
func HardPBConstr(lits []Lit, coeffs []int, atLeast int) Constr {
	return Constr{Lits: lits, Coeffs: coeffs, AtLeast: atLeast}
}

// SoftPBConstr returns an optional pseudo-boolean constraint.
func SoftPBConstr(lits []Lit, coeffs []int, atLeast int) Constr {
	return Constr{Lits: lits, Coeffs: coeffs, AtLeast: atLeast, Weight: 1}
}

// WeightedPBConstr returns a weighted optional pseudo-boolean constraint.
func WeightedPBConstr(lits []Lit, coeffs []int, atLeast int, weight int) Constr {
	return Constr{Lits: lits, Coeffs: coeffs, AtLeast: atLeast, Weight: weight}
}
