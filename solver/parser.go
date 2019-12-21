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
			lit := IntToLit(line[0])
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
				lits[j] = IntToLit(val)
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
	pb.simplify2()
	return &pb
}

func isSpace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r'
}

// readInt reads an int from r.
// 'b' is the last read byte. It can be a space, a '-' or a digit.
// The int can be negated.
// All spaces before the int value are ignored.
// Can return EOF.
func readInt(b *byte, r *bufio.Reader) (res int, err error) {
	for err == nil && isSpace(*b) {
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
		if isSpace(*b) {
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
				if err == io.EOF {
					if len(lits) != 0 { // This is not a trailing space at the end...
						return nil, fmt.Errorf("unfinished clause while EOF found")
					}
					break // When there are only several useless spaces at the end of the file, that is ok
				}
				if err != nil {
					return nil, fmt.Errorf("cannot parse clause: %v", err)
				}
				if val == 0 {
					pb.Clauses = append(pb.Clauses, NewClause(lits))
					break
				} else {
					if val > pb.NbVars || -val > pb.NbVars {
						return nil, fmt.Errorf("invalid literal %d for problem with %d vars only", val, pb.NbVars)
					}
					lits = append(lits, IntToLit(val))
				}
			}
		}
		b, err = r.ReadByte()
	}
	if err != io.EOF {
		return nil, err
	}
	pb.simplify2()
	return &pb, nil
}
