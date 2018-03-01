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

func oldParseCNF(f io.Reader) (*Problem, error) {
	scanner := bufio.NewScanner(f)
	var nbClauses int
	var pb Problem
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		if line[0] == 'p' {
			fields := strings.Fields(line)
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

// readInt reads an int from r.
// 'b' is the last read byte. It can be a space, a '-' or a digit.
// The int can be negated.
// All spaces before the int value are ignored.
// Can return EOF.
func readInt(b *byte, r *bufio.Reader) (res int, err error) {
	for err == nil && (*b == ' ' || *b == '\t' || *b == '\n') {
		*b, err = r.ReadByte()
	}
	if err == io.EOF {
		return res, io.EOF
	}
	if err != nil {
		return res, fmt.Errorf("could not read digit: %v", err)
	}
	neg := 1
	if *b == '-' {
		neg = -1
		*b, err = r.ReadByte()
		if err != nil {
			return 0, fmt.Errorf("cannot read int: %v", err)
		}
	}
	for err == nil {
		if *b < '0' || *b > '9' {
			return 0, fmt.Errorf("cannot read int: %q is not a digit", *b)
		}
		res = 10*res + int(*b-'0')
		*b, err = r.ReadByte()
		if *b == ' ' || *b == '\t' || *b == '\n' {
			break
		}
	}
	res *= neg
	return res, err
}

func parseHeader(r *bufio.Reader) (nbVars, nbClauses int, err error) {
	line, err := r.ReadString('\n')
	if err != nil {
		return 0, 0, fmt.Errorf("cannot read header: %v", err)
	}
	fields := strings.Fields(line)
	if len(fields) < 3 {
		return 0, 0, fmt.Errorf("invalid syntax %q in header", line)
	}
	nbVars, err = strconv.Atoi(fields[1])
	if err != nil {
		return 0, 0, fmt.Errorf("nbvars not an int : %q", fields[1])
	}
	nbClauses, err = strconv.Atoi(fields[2])
	if err != nil {
		return 0, 0, fmt.Errorf("nbClauses not an int : '%s'", fields[2])
	}
	return nbVars, nbClauses, nil
}

// ParseCNF parses a CNF file and returns the corresponding Problem.
func ParseCNF(f io.Reader) (*Problem, error) {
	r := bufio.NewReader(f)
	var (
		nbClauses int
		pb        Problem
	)
	b, err := r.ReadByte()
	for err == nil {
		if b == 'c' { // Ignore comment
			b, err = r.ReadByte()
			for err == nil && b != '\n' {
				b, err = r.ReadByte()
			}
		} else if b == 'p' { // Parse header
			pb.NbVars, nbClauses, err = parseHeader(r)
			if err != nil {
				return nil, fmt.Errorf("cannot parse CNF header: %v", err)
			}
			pb.Model = make([]decLevel, pb.NbVars)
			pb.Clauses = make([]*Clause, 0, nbClauses)
		} else {
			lits := make([]Lit, 0, 3) // Make room for some lits to improve performance
			for {
				val, err := readInt(&b, r)
				if err != nil {
					return nil, fmt.Errorf("cannot parse clause: %v", err)
				}
				if val == 0 {
					pb.Clauses = append(pb.Clauses, NewClause(lits))
					break
				} else {
					lits = append(lits, IntToLit(int32(val)))
				}
			}
		}
		b, err = r.ReadByte()
	}
	if err != io.EOF {
		return nil, err
	}
	pb.simplify()
	return &pb, nil
}
