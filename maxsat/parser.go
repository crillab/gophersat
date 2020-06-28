package maxsat

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	"../solver"
)

// A Solver is a [partial][weighted] MAXSAT solver.
// It implements solver.Interface.
type Solver struct {
	solver     *solver.Solver
	firstRelax int // Identifier of first relax variable: those must not be provided as part of the actual result
}

// Optimal looks for the optimal solution to the underlying problem.
// If results is not nil, it writes a suboptimal solution every time it finds a new, better one.
// In any case, it returns the optimal solution to the problem, or UNSAT if the problem cannot be found.
func (s *Solver) Optimal(results chan solver.Result, stop chan struct{}) solver.Result {
	if results == nil {
		res := s.solver.Optimal(nil, stop)
		return res
	}
	localRes := make(chan solver.Result)
	defer close(results)
	go s.solver.Optimal(localRes, stop)
	var res solver.Result
	for res = range localRes {
		if res.Status == solver.Sat {
			res.Model = res.Model[:s.firstRelax] // Remove relax vars from the model
		}
		results <- res
	}
	return res // Last result is returned
}

// Enumerate does not make sense for a MAXSAT problem, so it will panic when called.
// This might change in later versions.
func (s *Solver) Enumerate(models chan []bool, stop chan struct{}) int {
	panic("trying to call Enumerate on a MAXSAT problem")
}

// ParseWCNF parses a CNF file and returns the corresponding solver.Interface.
func ParseWCNF(f io.Reader) (solver.Interface, error) {
	scanner := bufio.NewScanner(f)
	var (
		nbVars    int
		nbClauses int
		topWeight int // weight of hard clauses
		clauses   [][]int
		weights   []int
		maxWeight int
		relaxLit  int // index of current relax lit
	)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		if line[0] == 'p' {
			fields := strings.Fields(line)
			if len(fields) < 4 || fields[1] != "wcnf" {
				return nil, fmt.Errorf("invalid syntax %q in WCNF file", line)
			}
			var err error
			nbVars, err = strconv.Atoi(fields[2])
			if err != nil {
				return nil, fmt.Errorf("nbvars not an int: %q", fields[2])
			}
			nbClauses, err = strconv.Atoi(fields[3])
			if err != nil {
				return nil, fmt.Errorf("nbClauses not an int: %q", fields[3])
			}
			relaxLit = nbVars + 1
			clauses = make([][]int, 0, nbClauses)
			weights = make([]int, 0, nbClauses)
			if len(fields) == 5 {
				topWeight, err = strconv.Atoi(fields[4])
				if err != nil {
					return nil, fmt.Errorf("top weight not an int: %q", fields[4])
				}
			}
		} else if line[0] != 'c' { // Not a header, not a comment : a clause
			clause, weight, err := parseWCNFClause(line, topWeight, relaxLit)
			if err != nil {
				return nil, err
			}
			clauses = append(clauses, clause)
			if topWeight == 0 || weight < topWeight {
				weights = append(weights, weight)
				maxWeight += weight
				relaxLit++
			}
		}
	}
	relaxLits := make([]solver.Lit, relaxLit-nbVars-1)
	for i := range relaxLits {
		relaxLits[i] = solver.IntToLit(int32(nbVars + i + 1))
	}
	prob := solver.ParseSlice(clauses)
	prob.SetCostFunc(relaxLits, weights)
	s := solver.New(prob)
	return &Solver{solver: s, firstRelax: nbVars}, nil
}

// Parses a WCNF line containing a clause and returns the clause with a relaxing literal, and its weight.
func parseWCNFClause(line string, topWeight, relaxLit int) (lits []int, weight int, err error) {
	fields := strings.Fields(line)
	lits = make([]int, len(fields)-1)
	for i := 0; i < len(fields); i++ { // Last field is clause terminator 0
		field := fields[i]
		if field == "" {
			continue
		}
		val, err := strconv.Atoi(field)
		if err != nil {
			return nil, 0, fmt.Errorf("Invalid integer %q in WCNF clause %q", field, line)
		}
		if i == 0 {
			weight = val
		} else {
			lits[i-1] = val
		}
	}
	if topWeight == 0 || weight < topWeight {
		lits[len(lits)-1] = relaxLit
	} else {
		lits = lits[:len(lits)-1]
	}
	return lits, weight, nil
}
