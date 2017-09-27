package solver

// This file deals with an attempt for an efficient binary/ternary clause allocator/deallocator.
// Since lots of binary/ternary clauses are created then (sometimes) destroyed, we allocate
// and reuse lots of them here, to relax the GC's work.

const (
	nbLitsAlloc = 5000000 // How many literals are initialized at first?
)

type allocator struct {
	lits    []Lit // A list of lits, that will be sliced to make []Lit
	ptrFree int   // Index of the first free item in lits
}

var alloc allocator

// newLits returns a slice of lits containing the given literals.
// It is taken from the preinitialized pool if possible,
// or is created from scratch.
func (a *allocator) newLits(lits ...Lit) []Lit {
	if a.ptrFree+len(lits) > len(a.lits) {
		a.lits = make([]Lit, nbLitsAlloc)
		copy(a.lits, lits)
		a.ptrFree = len(lits)
		return a.lits[:len(lits)]
	}
	copy(a.lits[a.ptrFree:], lits)
	a.ptrFree += len(lits)
	return a.lits[a.ptrFree-len(lits) : a.ptrFree]
}
