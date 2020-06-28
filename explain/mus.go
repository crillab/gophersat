package explain

import (
	"fmt"

	"github.com/crillab/gophersat/solver"
)

// MUSMaxSat returns a Minimal Unsatisfiable Subset for the problem using the MaxSat strategy.
// A MUS is an unsatisfiable subset such that, if any of its clause is removed,
// the problem becomes satisfiable.
// A MUS can be useful to understand why a problem is UNSAT, but MUSes are expensive to compute since
// a SAT solver must be called several times on parts of the original problem to find them.
// With the MaxSat strategy, the function computes the MUS through several calls to MaxSat.
func (pb *Problem) MUSMaxSat() (mus *Problem, err error) {
	pb2 := pb.clone()
	nbVars := pb2.NbVars
	nbClauses := pb2.nbClauses
	weights := make([]int, nbClauses)          // Weights of each clause
	relaxLits := make([]solver.Lit, nbClauses) // Set of all relax lits
	relaxLit := nbVars + 1                     // Index of last used relax lit
	for i, clause := range pb2.Clauses {
		pb2.Clauses[i] = append(clause, relaxLit)
		relaxLits[i] = solver.IntToLit(int32(relaxLit))
		weights[i] = 1
		relaxLit++
	}
	prob := solver.ParseSlice(pb2.Clauses)
	prob.SetCostFunc(relaxLits, weights)
	s := solver.New(prob)
	s.Verbose = pb.Options.Verbose
	var musClauses [][]int
	done := make([]bool, nbClauses) // Indicates whether a clause is already part of MUS or not yet
	for {
		cost := s.Minimize()
		if cost == -1 {
			return makeMus(nbVars, musClauses), nil
		}
		if cost == 0 {
			return nil, fmt.Errorf("cannot extract MUS from satisfiable problem")
		}
		model := s.Model()
		for i, clause := range pb.Clauses {
			if !done[i] && !satClause(clause, model) {
				// The clause is part of the MUS
				pb2.Clauses = append(pb2.Clauses, []int{-(nbVars + i + 1)}) // Now, relax lit has to be false
				pb2.nbClauses++
				musClauses = append(musClauses, clause)
				done[i] = true
				// Make it a hard clause before restarting solver
				lits := make([]solver.Lit, len(clause))
				for j, lit := range clause {
					lits[j] = solver.IntToLit(int32(lit))
				}
				s.AppendClause(solver.NewClause(lits))
			}
		}
		if pb.Options.Verbose {
			fmt.Printf("c Currently %d/%d clauses in MUS\n", len(musClauses), nbClauses)
		}
		prob = solver.ParseSlice(pb2.Clauses)
		prob.SetCostFunc(relaxLits, weights)
		s = solver.New(prob)
		s.Verbose = pb.Options.Verbose
	}
}

// true iff the clause is satisfied by the model
func satClause(clause []int, model []bool) bool {
	for _, lit := range clause {
		if (lit > 0 && model[lit-1]) || (lit < 0 && !model[-lit-1]) {
			return true
		}
	}
	return false
}

func makeMus(nbVars int, clauses [][]int) *Problem {
	mus := &Problem{
		Clauses:   clauses,
		NbVars:    nbVars,
		nbClauses: len(clauses),
		units:     make([]int, nbVars),
	}
	for _, clause := range clauses {
		if len(clause) == 1 {
			lit := clause[0]
			if lit > 0 {
				mus.units[lit-1] = 1
			} else {
				mus.units[-lit-1] = -1
			}
		}
	}
	return mus
}

// MUSInsertion returns a Minimal Unsatisfiable Subset for the problem using the insertion method.
// A MUS is an unsatisfiable subset such that, if any of its clause is removed,
// the problem becomes satisfiable.
// A MUS can be useful to understand why a problem is UNSAT, but MUSes are expensive to compute since
// a SAT solver must be called several times on parts of the original problem to find them.
// The insertion algorithm is efficient is many cases, as it calls the same solver several times in a row.
// However, in some cases, the number of calls will be higher than using other methods.
// For instance, if  called on a formula that is already a MUS, it will perform n*(n-1) calls to SAT, where
// n is the number of clauses of the problem.
func (pb *Problem) MUSInsertion() (mus *Problem, err error) {
	pb2, err := pb.UnsatSubset()
	if err != nil {
		return nil, fmt.Errorf("could not extract MUS: %v", err)
	}
	mus = &Problem{NbVars: pb2.NbVars}
	clauses := pb2.Clauses
	for {
		if pb.Options.Verbose {
			fmt.Printf("c mus currently contains %d clauses\n", mus.nbClauses)
		}
		s := solver.New(solver.ParseSliceNb(mus.Clauses, mus.NbVars))
		s.Verbose = pb.Options.Verbose
		st := s.Solve()
		if st == solver.Unsat { // Found the MUS
			return mus, nil
		}
		// Add clauses until the problem becomes UNSAT
		idx := 0
		for st == solver.Sat {
			clause := clauses[idx]
			lits := make([]solver.Lit, len(clause))
			for i, lit := range clause {
				lits[i] = solver.IntToLit(int32(lit))
			}
			cl := solver.NewClause(lits)
			s.AppendClause(cl)
			idx++
			st = s.Solve()
		}
		idx--                                           // We went one step too far, go back
		mus.Clauses = append(mus.Clauses, clauses[idx]) // Last clause is part of the MUS
		mus.nbClauses++
		if pb.Options.Verbose {
			fmt.Printf("c removing %d/%d clause(s)\n", len(clauses)-idx, len(clauses))
		}
		clauses = clauses[:idx] // Remaining clauses are not part of the MUS
	}
}

// MUSDeletion returns a Minimal Unsatisfiable Subset for the problem using the insertion method.
// A MUS is an unsatisfiable subset such that, if any of its clause is removed,
// the problem becomes satisfiable.
// A MUS can be useful to understand why a problem is UNSAT, but MUSes are expensive to compute since
// a SAT solver must be called several times on parts of the original problem to find them.
// The deletion algorithm is guaranteed to call exactly n SAT solvers, where n is the number of clauses in the problem.
// It can be quite efficient, but each time the solver is called, it is starting from scratch.
// Other methods keep the solver "hot", so despite requiring more calls, these methods can be more efficient in practice.
func (pb *Problem) MUSDeletion() (mus *Problem, err error) {

	pb2, err := pb.UnsatSubset()
	if err != nil {
		return nil, fmt.Errorf("could not extract MUS: %v", err)
	}
	pb2.NbVars += pb2.nbClauses          // Add one relax var for each clause
	for i, clause := range pb2.Clauses { // Add relax lit to each clause
		newClause := make([]int, len(clause)+1)
		copy(newClause, clause)
		newClause[len(clause)] = pb.NbVars + i + 1 // Add relax lit to the clause
		pb2.Clauses[i] = newClause
	}
	s := solver.New(solver.ParseSlice(pb2.Clauses))
	asumptions := make([]solver.Lit, pb2.nbClauses)
	for i := 0; i < pb2.nbClauses; i++ {
		asumptions[i] = solver.IntToLit(int32(-(pb.NbVars + i + 1))) // At first, all asumptions are false
	}
	for i := range pb2.Clauses {
		// Relax current clause
		asumptions[i] = asumptions[i].Negation()
		s.Assume(asumptions)
		if s.Solve() == solver.Sat {
			// It is now sat; reinsert the clause, i.e re-falsify the relax lit
			asumptions[i] = asumptions[i].Negation()
			if pb.Options.Verbose {
				fmt.Printf("c clause %d/%d: kept\n", i+1, pb2.nbClauses)
			}
		} else if pb.Options.Verbose {
			fmt.Printf("c clause %d/%d: removed\n", i+1, pb2.nbClauses)
		}
	}
	mus = &Problem{
		NbVars: pb.NbVars,
	}
	for i, val := range asumptions {
		if !val.IsPositive() {
			// Lit is not relaxed, meaning the clause is part of the MUS
			clause := pb2.Clauses[i]
			clause = clause[:len(clause)-1] // Remove relax lit
			mus.Clauses = append(mus.Clauses, clause)
		}
		mus.nbClauses = len(mus.Clauses)
	}
	return mus, nil
}

// MUS returns a Minimal Unsatisfiable Subset for the problem.
// A MUS is an unsatisfiable subset such that, if any of its clause is removed,
// the problem becomes satisfiable.
// A MUS can be useful to understand why a problem is UNSAT, but MUSes are expensive to compute since
// a SAT solver must be called several times on parts of the original problem to find them.
// The exact algorithm used to compute the MUS is not guaranteed. If you want to use a given algorithm,
// use the relevant functions.
func (pb *Problem) MUS() (mus *Problem, err error) {

	return pb.MUSDeletion()
}
