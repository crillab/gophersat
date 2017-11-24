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

// ParsePBConstrs parses and returns a PB problem from PBConstr values.
func ParsePBConstrs(constrs []PBConstr) *Problem {
	var pb Problem
	for _, constr := range constrs {
		for i := range constr.Lits {
			lit := IntToLit(int32(constr.Lits[i]))
			v := lit.Var()
			if int(v) >= pb.NbVars {
				pb.NbVars = int(v) + 1
			}
		}
		card := constr.AtLeast
		if card <= 0 { // Clause is trivially SAT, ignore
			continue
		}
		sumW := constr.WeightSum()
		if sumW < card { // Clause cannot be satsfied
			pb.Status = Unsat
			return &pb
		}
		if sumW == card { // All lits must be true
			for i := range constr.Lits {
				lit := IntToLit(int32(constr.Lits[i]))
				pb.Units = append(pb.Units, lit)
			}
		} else {
			lits := make([]Lit, len(constr.Lits))
			for j, val := range constr.Lits {
				lits[j] = IntToLit(int32(val))
			}
			pb.Clauses = append(pb.Clauses, NewPBClause(lits, constr.Weights, card))
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
	pb.simplifyPB()
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

// parsePBOptim parses the "min:" instruction.
func (pb *Problem) parsePBOptim(fields []string, line string) error {
	weights, lits, err := pb.parseTerms(fields[1:], line)
	if err != nil {
		return err
	}
	pb.minLits = make([]Lit, len(lits))
	for i, lit := range lits {
		pb.minLits[i] = IntToLit(int32(lit))
	}
	pb.minWeights = weights
	return nil
}

func (pb *Problem) parsePBLine(line string) error {
	if line[len(line)-1] != ';' {
		return fmt.Errorf("line %q does not end with semicolon", line)
	}
	fields := strings.Fields(line[:len(line)-1])
	if len(fields) == 0 {
		return fmt.Errorf("empty line in file")
	}
	if fields[0] == "min:" { // Optimization constraint
		return pb.parsePBOptim(fields, line)
	}
	return pb.parsePBConstrLine(fields, line)
}

func (pb *Problem) parsePBConstrLine(fields []string, line string) error {
	if len(fields) < 4 || len(fields)%2 != 0 {
		return fmt.Errorf("invalid syntax %q", line)
	}
	operator := fields[len(fields)-2]
	if operator != ">=" && operator != "=" {
		return fmt.Errorf("invalid operator %q in %q: expected \">=\" or \"=\"", operator, line)
	}
	rhs, err := strconv.Atoi(fields[len(fields)-1])
	if err != nil {
		return fmt.Errorf("invalid value %q in %q: %v", fields[len(fields)-1], line, err)
	}
	weights, lits, err := pb.parseTerms(fields[:len(fields)-2], line)
	if err != nil {
		return err
	}
	var constrs []PBConstr
	if operator == ">=" {
		constrs = []PBConstr{GtEq(lits, weights, rhs)}
	} else {
		constrs = Eq(lits, weights, rhs)
	}
	for _, constr := range constrs {
		card := constr.AtLeast
		sumW := constr.WeightSum()
		if sumW < card { // Clause cannot be satsfied
			pb.Status = Unsat
			return nil
		}
		if sumW == card { // All lits must be true
			for i := range constr.Lits {
				lit := IntToLit(int32(constr.Lits[i]))
				pb.Units = append(pb.Units, lit)
			}
		} else {
			lits := make([]Lit, len(constr.Lits))
			for j, val := range constr.Lits {
				lits[j] = IntToLit(int32(val))
			}
			pb.Clauses = append(pb.Clauses, NewPBClause(lits, constr.Weights, card))
		}
	}
	return nil
}

func (pb *Problem) parseTerms(terms []string, line string) (weights []int, lits []int, err error) {
	weights = make([]int, len(terms)/2)
	lits = make([]int, len(terms)/2)
	for i := range weights {
		w, err := strconv.Atoi(terms[i*2])
		if err != nil {
			return nil, nil, fmt.Errorf("invalid weight %q in %q: %v", terms[i*2], line, err)
		}
		weights[i] = w
		l := terms[i*2+1]
		if !strings.HasPrefix(l, "x") && !strings.HasPrefix(l, "~x") || len(l) < 2 {
			return nil, nil, fmt.Errorf("invalid variable name %q in %q", l, line)
		}
		if l[0] == '~' {
			lits[i], err = strconv.Atoi(l[2:])
		} else {
			lits[i], err = strconv.Atoi(l[1:])
		}
		if err != nil {
			return nil, nil, fmt.Errorf("invalid variable %q in %q: %v", l, line, err)
		}
		if lits[i] > pb.NbVars {
			pb.NbVars = lits[i]
		}
		if l[0] == '~' {
			lits[i] = -lits[i]
		}
	}
	return weights, lits, nil
}

// ParseOPB parses a file corresponding to the OPB syntax.
// See http://www.cril.univ-artois.fr/PB16/format.pdf for more details.
func ParseOPB(f io.Reader) (*Problem, error) {
	scanner := bufio.NewScanner(f)
	var pb Problem
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || line[0] == '*' {
			continue
		}
		if err := pb.parsePBLine(line); err != nil {
			return nil, err
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("could not parse OPB: %v", err)
	}
	pb.Model = make([]decLevel, pb.NbVars)
	pb.simplifyPB()
	return &pb, nil
}
