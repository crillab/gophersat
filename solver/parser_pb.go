package solver

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

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
	pb.simplifyCard()
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
				found := false
				for _, u := range pb.Units {
					if u == lit {
						found = true
						break
					}
				}
				if !found {
					pb.Units = append(pb.Units, lit)
				}
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
	if len(fields) < 3 {
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
	weights = make([]int, 0, len(terms)/2)
	lits = make([]int, 0, len(terms)/2)
	i := 0
	for i < len(terms) {
		var l string
		w, err := strconv.Atoi(terms[i])
		if err != nil {
			l = terms[i]
			if !strings.HasPrefix(l, "x") && !strings.HasPrefix(l, "~x") {
				return nil, nil, fmt.Errorf("invalid weight %q in %q: %v", terms[i*2], line, err)
			}
			// This is a weightless lit, i.e a lit with weight 1.
			weights = append(weights, 1)
		} else {
			weights = append(weights, w)
			i++
			l = terms[i]
			if !strings.HasPrefix(l, "x") && !strings.HasPrefix(l, "~x") || len(l) < 2 {
				return nil, nil, fmt.Errorf("invalid variable name %q in %q", l, line)
			}
		}
		var lit int
		if l[0] == '~' {
			lit, err = strconv.Atoi(l[2:])
			lits = append(lits, -lit)
		} else {
			lit, err = strconv.Atoi(l[1:])
			lits = append(lits, lit)
		}
		if err != nil {
			return nil, nil, fmt.Errorf("invalid variable %q in %q: %v", l, line, err)
		}
		if lit > pb.NbVars {
			pb.NbVars = lit
		}
		i++
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
