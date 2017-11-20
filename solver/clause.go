package solver

import (
	"fmt"
	"log"
	"strings"
)

// A Clause is a list of Lit, associated with possible data (for learned clauses).
type Clause struct {
	lits []Lit
	// lbdValue's bits are as follow:
	// leftmost bit: learned flag.
	// second bit: locked flag (if learned).
	// last 30 bits: LBD value (if learned) or minimal cardinality - 1 (if !learned).
	// NOTE: actual cardinality is value + 1, since this is the default value and go defaults to 0.
	lbdValue uint32
	activity float32
	weights  []int // For PB constraints, weight of each literal. If nil, weights are all 1.
}

const (
	learnedMask uint32 = 1 << 31
	lockedMask  uint32 = 1 << 30
	bothMasks   uint32 = learnedMask | lockedMask
)

// NewClause returns a clause whose lits are given as an argument.
func NewClause(lits []Lit) *Clause {
	return &Clause{lits: lits}
}

// NewCardClause returns a clause whose lits are given as an argument and
// for which at least 'card' literals must be true.
// Note tha NewClause(lits) is equivalent to NewCardClause(lits, 1).
func NewCardClause(lits []Lit, card int) *Clause {
	if card < 1 || card > len(lits) {
		panic("Invalid cardinality value")
	}
	return &Clause{lits: lits, lbdValue: uint32(card - 1)}
}

// NewPBClause returns a pseudo-boolean clause with the given lits, weights and minimal cardinality.
func NewPBClause(lits []Lit, weights []int, card int) *Clause {
	if card < 1 {
		panic("Invalid cardinality value")
	}
	return &Clause{lits: lits, lbdValue: uint32(card - 1), weights: weights}
}

// NewLearnedClause returns a new clause marked as learned.
func NewLearnedClause(lits []Lit) *Clause {
	return &Clause{lits: lits, lbdValue: learnedMask}
}

// Cardinality returns the minimum number of literals that must be true to satisfy the clause.
func (c *Clause) Cardinality() int {
	if c.Learned() {
		return 1
	}
	return int(c.lbdValue & ^bothMasks) + 1
}

// Learned returns true iff c was a learned clause.
func (c *Clause) Learned() bool {
	return c.lbdValue&learnedMask == learnedMask
}

// PseudoBoolean returns true iff c is a pseudo boolean constraint, and not
// just a propositional clause or cardinality constraint.
func (c *Clause) PseudoBoolean() bool {
	return c.weights != nil
}

func (c *Clause) lock() {
	c.lbdValue = c.lbdValue | lockedMask
}

func (c *Clause) unlock() {
	c.lbdValue = c.lbdValue & ^lockedMask
}

func (c *Clause) lbd() int {
	return int(c.lbdValue & ^bothMasks)
}

func (c *Clause) setLbd(lbd int) {
	c.lbdValue = (c.lbdValue & bothMasks) | uint32(lbd)
}

func (c *Clause) incLbd() {
	c.lbdValue++
}

func (c *Clause) isLocked() bool {
	return c.lbdValue&bothMasks == bothMasks
}

// Len returns the nb of lits in the clause.
func (c *Clause) Len() int {
	return len(c.lits)
}

// First returns the first lit from the clause.
func (c *Clause) First() Lit {
	return c.lits[0]
}

// Second returns the second lit from the clause.
func (c *Clause) Second() Lit {
	return c.lits[1]
}

// Get returns the ith literal from the clause.
func (c *Clause) Get(i int) Lit {
	return c.lits[i]
}

// Set sets the ith literal of the clause.
func (c *Clause) Set(i int, l Lit) {
	c.lits[i] = l
}

// Weight returns the weight of the ith literal.
// In a propositional clause or a cardinality constraint, that value will always be 1.
func (c *Clause) Weight(i int) int {
	if c.weights == nil {
		return 1
	}
	return c.weights[i]
}

// WeightSum returns the sum of the PB weights.
// If c is a propositional clause, the function will return the length of the clause.
func (c *Clause) WeightSum() int {
	if c.weights == nil {
		return len(c.lits)
	}
	res := 0
	for _, w := range c.weights {
		res += w
	}
	return res
}

// swap swaps the ith and jth lits from the clause.
func (c *Clause) swap(i, j int) {
	c.lits[i], c.lits[j] = c.lits[j], c.lits[i]
	if c.weights != nil {
		c.weights[i], c.weights[j] = c.weights[j], c.weights[i]
	}
}

// updateCardinality adds "add" to c's cardinality.
// Must not be called on learned clauses!
func (c *Clause) updateCardinality(add int) {
	log.Printf("card is %d, add is %d", c.lbdValue+1, add)
	if add < 0 && uint32(add) > c.lbdValue {
		c.lbdValue = 0
	}
	c.lbdValue += uint32(add)
}

// removeLit remove the idx'th lit from c.
// The order of literals might be updated, so this function
// should not be called once the whole solver was created,
// only during simplification of the problem.
func (c *Clause) removeLit(idx int) {
	c.lits[idx] = c.lits[len(c.lits)-1]
	c.lits = c.lits[:len(c.lits)-1]
	if c.weights != nil {
		c.weights[idx] = c.weights[len(c.weights)-1]
		c.weights = c.weights[:len(c.weights)-1]
	}
}

// Shrink reduces the length of the clauses, by removing all lits
// starting from position newLen.
func (c *Clause) Shrink(newLen int) {
	c.lits = c.lits[:newLen]
	if c.weights != nil {
		c.weights = c.weights[:newLen]
	}
}

// CNF returns a DIMACS CNF representation of the clause.
func (c *Clause) CNF() string {
	res := ""
	for _, lit := range c.lits {
		res += fmt.Sprintf("%d ", lit.Int())
	}
	return fmt.Sprintf("%s0", res)
}

// PBString returns a string representation of c as a pseudo-boolean expression.
func (c *Clause) PBString() string {
	terms := make([]string, c.Len())
	for i, lit := range c.lits {
		weight := ""
		if c.weights != nil && c.weights[i] != 1 {
			weight = fmt.Sprintf("%d * ", c.weights[i])
		}
		val := lit.Int()
		sign := ""
		if val < 0 {
			val = -val
			sign = "¬"
		}
		terms[i] = fmt.Sprintf("%s%sx%d", weight, sign, val)
	}
	return fmt.Sprintf("%s ≥ %d", strings.Join(terms, " + "), c.Cardinality())
}
