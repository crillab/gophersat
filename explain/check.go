// Package explain provides facilities to check and understand UNSAT instances.
package explain

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/crillab/gophersat/solver"
)

// Options is a set of options that can be set to true during the checking process.
type Options struct {
	// If Verbose is true, information about resolution will be written on stdout.
	Verbose bool
}

// Checks whether the clause satisfies the problem or not.
// Will return true if the problem is UNSAT, false if it is SAT or indeterminate.
func unsat(pb *Problem, clause []int) bool {
	oldUnits := make([]int, len(pb.units))
	copy(oldUnits, pb.units)
	// lits is supposed to be implied by the problem.
	// We add the negation of each lit as a unit clause to see if this is true.
	for _, lit := range clause {
		if lit > 0 {
			pb.units[lit-1] = -1
		} else {
			pb.units[-lit-1] = 1
		}
	}
	res := pb.unsat()
	pb.units = oldUnits // We must restore the previous state
	return res
}

// UnsatChan will wait RUP clauses from ch and use them as a certificate.
// It will return true iff the certificate is valid, i.e iff it makes the problem UNSAT
// through unit propagation.
// If pb.Options.ExtractSubset is true, a subset will also be extracted for that problem.
func (pb *Problem) UnsatChan(ch chan string) (valid bool, err error) {
	defer pb.restore()
	pb.initTagged()
	for line := range ch {

		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		if _, err := strconv.Atoi(fields[0]); err != nil {
			// This is not a clause: ignore the line
			continue
		}
		clause, err := parseClause(fields)
		if err != nil {
			return false, err
		}
		if !unsat(pb, clause) {
			return false, nil
		}
		if len(clause) == 0 {
			// This is the empty and unit propagation made the problem UNSAT: we're done.
			return true, nil
		}
		// Since clause is a logical consequence, append it to the problem
		pb.Clauses = append(pb.Clauses, clause)
	}

	// If we did not send any information through the channel
	// It implies that the problem is trivially unsatisfiable
	// Since we had only unit clauses inside the channel.
	return true, nil
}

// Unsat will parse a certificate, and return true iff the certificate is valid, i.e iff it makes the problem UNSAT
// through unit propagation.
// If pb.Options.ExtractSubset is true, a subset will also be extracted for that problem.
func (pb *Problem) Unsat(cert io.Reader) (valid bool, err error) {
	defer pb.restore()
	pb.initTagged()
	sc := bufio.NewScanner(cert)
	for sc.Scan() {
		line := sc.Text()
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		if _, err := strconv.Atoi(fields[0]); err != nil {
			// This is not a clause: ignore the line
			continue
		}
		clause, err := parseClause(fields)
		if err != nil {
			return false, err
		}
		if !unsat(pb, clause) {
			return false, nil
		}
		// Since clause is a logical consequence, append it to the problem
		pb.Clauses = append(pb.Clauses, clause)
	}
	if err := sc.Err(); err != nil {
		return false, fmt.Errorf("could not parse certificate: %v", err)
	}
	return true, nil
}

// UnsatSubset returns an unsatisfiable subset of the problem.
// The subset is not guaranteed to be a MUS, meaning some clauses of the resulting
// problem might be removed while still keeping the unsatisfiability of the problem.
// However, this method is much more efficient than extracting a MUS, as it only calls
// the SAT solver once.
func (pb *Problem) UnsatSubset() (subset *Problem, err error) {
	s := solver.New(solver.ParseSlice(pb.Clauses))
	s.Certified = true
	s.CertChan = make(chan string)
	status := solver.Unsat
	go func() {
		status = s.Solve()
		close(s.CertChan)
	}()
	if valid, err := pb.UnsatChan(s.CertChan); !valid || status == solver.Sat {

		return nil, fmt.Errorf("problem is not UNSAT")

	} else if err != nil {

		return nil, fmt.Errorf("could not solve problem: %v", err)
	}
	subset = &Problem{
		NbVars: pb.NbVars,
	}
	for i, clause := range pb.Clauses {
		if pb.tagged[i] {
			// clause was used to prove pb is UNSAT: it's part of the subset
			subset.Clauses = append(subset.Clauses, clause)
			subset.NbClauses++
		}
	}
	return subset, nil
}
