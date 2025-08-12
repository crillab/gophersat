package maxsat

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/crillab/gophersat/solver"
)

// A Model associates variable names with a binding.
type Model map[string]bool

// A Problem is a set of constraints.
type Problem struct {
	solver       *solver.Solver
	intVars      map[string]int // for each var, its integer counterpart
	varInts      []string       // for each int value, the associated variable
	blockWeights map[int]int    // for each blocking literal, the weight of the associated constraint
	maxWeight    int            // sum of all blockWeights
}

// New returns a new problem associated with the given constraints.
func New(constrs ...Constr) *Problem {
	pb := &Problem{intVars: make(map[string]int), blockWeights: make(map[int]int)}
	clauses := make([]solver.PBConstr, len(constrs))
	for i, constr := range constrs {
		lits := make([]int, len(constr.Lits))
		for j, lit := range constr.Lits {
			v := fmt.Sprintf("VAR_%s", lit.Var)
			if _, ok := pb.intVars[v]; !ok {
				pb.varInts = append(pb.varInts, v)
				pb.intVars[v] = len(pb.varInts)
			}
			lits[j] = pb.intVars[v]
			if lit.Negated {
				lits[j] = -lits[j]
			}
		}
		var coeffs []int
		if len(constr.Coeffs) != 0 {
			coeffs = make([]int, len(constr.Coeffs))
			copy(coeffs, constr.Coeffs)
		}
		if constr.Weight != 0 { // Soft constraint: add blocking literal
			pb.varInts = append(pb.varInts, fmt.Sprintf("BLOCK_%d", i)) // Create new blocking lit
			bl := len(pb.varInts)
			pb.blockWeights[bl] = constr.Weight
			pb.maxWeight += constr.Weight
			lits = append(lits, bl)
			if coeffs != nil { // If this is a clause, there is no explicit coeff
				// TODO: deal with card constraints: AtLeast !=1 but coeffs == nil!
				coeffs = append(coeffs, constr.AtLeast)
			}
		}
		clauses[i] = solver.GtEq(lits, coeffs, constr.AtLeast)
	}
	optLits := make([]solver.Lit, 0, len(pb.blockWeights))
	optWeights := make([]int, 0, len(pb.blockWeights))
	for v, w := range pb.blockWeights {
		optLits = append(optLits, solver.IntToLit(int32(v)))
		optWeights = append(optWeights, w)
	}
	prob := solver.ParsePBConstrs(clauses)
	prob.SetCostFunc(optLits, optWeights)
	pb.solver = solver.New(prob)
	return pb
}

// SetVerbose makes the underlying solver verbose, or not.
func (pb *Problem) SetVerbose(verbose bool) {
	pb.solver.Verbose = verbose
}

// Output the problem to stdout in the OPB format.
func (pb *Problem) Output() {
	fmt.Println(pb.solver.PBString())
}

// Solver gives access to the solver.Solver used to solve the MAXSAT problem.
// Unless you have specific needs, youè will usually not need to call this method,
// and rather want to call pb.Solve() instead.
func (pb *Problem) Solver() *solver.Solver {
	return pb.solver
}

// Solve returns an optimal Model for the problem, the associated cost, and the indices of any broken soft constraints.
// If the model is nil, the problem was not satisfiable (i.e hard clauses could not be satisfied).
func (pb *Problem) Solve() (Model, int, []int) {
	cost := pb.solver.Minimize()
	if cost == -1 {
		return nil, -1, nil
	}
	var broken []int
	res := make(Model)
	for i, binding := range pb.solver.Model() {
		name := pb.varInts[i]
		if name, ok := strings.CutPrefix(name, "BLOCK_"); ok { // Ignore blocking lits
			if binding { // if the blocking lit was disabled, add it to the broken list
				id, err := strconv.Atoi(name)
				if err == nil {
					broken = append(broken, id)
				}
			}
		} else if name, ok := strings.CutPrefix(name, "VAR_"); ok { // "Fix" normal lits
			res[name] = binding
		} else {
			panic("An unknown variable entered the model: " + name)
		}
	}
	return res, cost, broken
}
