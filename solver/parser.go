package solver

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// ParseSlice parse a slice of slice of lits and returns the equivalent problem.
// The argument is supposed to be a well-formed CNF.
func ParseSlice(cnf [][]int) *Problem {
	var pb Problem
	for _, line := range cnf {
		switch len(line) {
		case 0:
			pb.Status = Unsat
			return &pb
		case 1:
			if line[0] == 0 {
				panic("null unit clause")
			}
			lit := IntToLit(int32(line[0]))
			v := lit.Var()
			if int(v) >= pb.NbVars {
				pb.NbVars = int(v) + 1
			}
			pb.Units = append(pb.Units, lit)
		default:
			lits := make([]Lit, len(line))
			for j, val := range line {
				if val == 0 {
					panic("null literal in clause %q")
				}
				lits[j] = IntToLit(int32(val))
				if v := int(lits[j].Var()); v >= pb.NbVars {
					pb.NbVars = v + 1
				}
			}
			pb.Clauses = append(pb.Clauses, NewClause(lits))
		}
	}
	pb.Model = make([]decLevel, pb.NbVars)
	for _, unit := range pb.Units {
		v := unit.Var()
		if pb.Model[v] == 0 {
			if unit.IsPositive() {
				pb.Model[v] = 1
			} else {
				pb.Model[v] = -1
			}
		} else if pb.Model[v] > 0 != unit.IsPositive() {
			pb.Status = Unsat
			return &pb
		}
	}
	pb.simplify()
	return &pb
}

// ParseCardConstrs parses the given cardinality constraints.
// Will panic if a zero value appears in the literals.
func ParseCardConstrs(constrs []CardConstr) *Problem {
	var pb Problem
	for _, constr := range constrs {
		card := constr.AtLeast
		if card <= 0 { // Clause is trivially SAT, ignore
			continue
		}
		if len(constr.Lits) < card { // Clause cannot be satsfied
			pb.Status = Unsat
			return &pb
		}
		if len(constr.Lits) == card { // All lits must be true
			for i := range constr.Lits {
				if constr.Lits[i] == 0 {
					panic("literal 0 found in clause")
				}
				lit := IntToLit(int32(constr.Lits[i]))
				v := lit.Var()
				if int(v) >= pb.NbVars {
					pb.NbVars = int(v) + 1
				}
				pb.Units = append(pb.Units, lit)
			}
		} else {
			lits := make([]Lit, len(constr.Lits))
			for j, val := range constr.Lits {
				if val == 0 {
					panic("literal 0 found in clause")
				}
				lits[j] = IntToLit(int32(val))
				if v := int(lits[j].Var()); v >= pb.NbVars {
					pb.NbVars = v + 1
				}
			}
			pb.Clauses = append(pb.Clauses, NewCardClause(lits, card))
		}
	}
	pb.Model = make([]decLevel, pb.NbVars)
	for _, unit := range pb.Units {
		v := unit.Var()
		if pb.Model[v] == 0 {
			if unit.IsPositive() {
				pb.Model[v] = 1
			} else {
				pb.Model[v] = -1
			}
		} else if pb.Model[v] > 0 != unit.IsPositive() {
			pb.Status = Unsat
			return &pb
		}
	}
	pb.simplify()
	return &pb
}

// Parses a CNF line containing a clause and adds it to the problem.
func (pb *Problem) parseClause(line string) error {
	fields := strings.Fields(line)
	lits := make([]Lit, len(fields)-1)
	for i, field := range fields {
		if i == len(fields)-1 { // Ignore last field: it is the 0 clause terminator
			break
		}
		if field == "" {
			continue
		}
		cnfLit, err := strconv.Atoi(field)
		if err != nil {
			return fmt.Errorf("Invalid literal %q in CNF clause %q", field, line)
		}
		lits[i] = IntToLit(int32(cnfLit))
	}
	switch len(lits) {
	case 0:
		pb.Status = Unsat
		pb.Clauses = nil
	case 1:
		lit := lits[0]
		pb.Units = append(pb.Units, lit)
		v := lit.Var()
		if pb.Model[v] == 0 {
			if lit.IsPositive() {
				pb.Model[lit.Var()] = 1
			} else {
				pb.Model[lit.Var()] = -1
			}
		} else if pb.Model[v] > 0 != lit.IsPositive() {
			pb.Status = Unsat
		}
	default:
		pb.Clauses = append(pb.Clauses, NewClause(lits))
	}
	return nil
}

// ParseCNF parses a CNF file and returns the corresponding Problem.
func ParseCNF(f io.Reader) (*Problem, error) {
	scanner := bufio.NewScanner(f)
	var nbClauses int
	var pb Problem
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		if line[0] == 'p' {
			fields := strings.Split(line, " ")
			if len(fields) < 4 {
				return nil, fmt.Errorf("invalid syntax %q in CNF file", line)
			}
			var err error
			pb.NbVars, err = strconv.Atoi(fields[2])
			if err != nil {
				return nil, fmt.Errorf("nbvars not an int : '%s'", fields[2])
			}
			pb.Model = make([]decLevel, pb.NbVars)
			nbClauses, err = strconv.Atoi(fields[3])
			if err != nil {
				return nil, fmt.Errorf("nbClauses not an int : '%s'", fields[3])
			}
			pb.Clauses = make([]*Clause, 0, nbClauses)
		} else if line[0] != 'c' { // Not a header, not a comment : a clause
			if err := pb.parseClause(line); err != nil {
				return nil, err
			}
		}
	}
	pb.simplify()
	return &pb, nil
}
