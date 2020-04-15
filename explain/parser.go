package explain

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// parseClause parses a line representing a clause in the DIMACS CNF syntax.
func parseClause(fields []string) ([]int, error) {
	clause := make([]int, 0, len(fields)-1)
	for _, rawLit := range fields {
		lit, err := strconv.Atoi(rawLit)
		if err != nil {
			return nil, fmt.Errorf("could not parse clause %v: %v", fields, err)
		}
		if lit != 0 {
			clause = append(clause, lit)
		}
	}
	return clause, nil
}

// ParseCNF parses a CNF and returns the associated problem.
func ParseCNF(r io.Reader) (*Problem, error) {
	sc := bufio.NewScanner(r)
	var pb Problem
	for sc.Scan() {
		line := sc.Text()
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		switch fields[0] {
		case "c":
			continue
		case "p":
			if err := pb.parseHeader(fields); err != nil {
				return nil, fmt.Errorf("could not parse header %q: %v", line, err)
			}
		default:
			if err := pb.parseClause(fields); err != nil {
				return nil, fmt.Errorf("could not parse clause %q: %v", line, err)
			}
		}
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("could not parse problem: %v", err)
	}
	return &pb, nil
}

func (pb *Problem) parseHeader(fields []string) error {
	if len(fields) != 4 {
		return fmt.Errorf("expected 4 fields, got %d", len(fields))
	}
	strVars := fields[2]
	strClauses := fields[3]
	var err error
	pb.NbVars, err = strconv.Atoi(fields[2])
	if err != nil {
		return fmt.Errorf("invalid number of vars %q: %v", strVars, err)
	}
	if pb.NbVars < 0 {
		return fmt.Errorf("negative number of vars %d", pb.NbVars)
	}
	pb.units = make([]int, pb.NbVars)
	pb.nbClauses, err = strconv.Atoi(fields[3])
	if err != nil {
		return fmt.Errorf("invalid number of clauses %s: %v", strClauses, err)
	}
	if pb.nbClauses < 0 {
		return fmt.Errorf("negative number of clauses %d", pb.nbClauses)
	}
	pb.Clauses = make([][]int, 0, pb.nbClauses)
	return nil
}

func (pb *Problem) parseClause(fields []string) error {
	clause, err := parseClause(fields)
	if err != nil {
		return err
	}
	pb.Clauses = append(pb.Clauses, clause)
	if len(clause) == 1 {
		lit := clause[0]
		v := lit
		if lit < 0 {
			v = -v
		}
		if v > pb.NbVars {
			// There was an error in the header
			return fmt.Errorf("found lit %d but problem was supposed to hold only %d vars", lit, pb.NbVars)
		}
		if lit > 0 {
			pb.units[v-1] = 1
		} else {
			pb.units[v-1] = -1
		}
	}
	return nil
}
