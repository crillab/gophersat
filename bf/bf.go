package bf

import (
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"github.com/crillab/gophersat/solver"
)

// A Formula is any kind of boolean formula, not necessarily in CNF form.
type Formula interface {
	nnf() Formula
	String() string
}

// Solve solves the given formula.
// f is first converted as a CNF formula. It is then given to gophersat.
// The function returns a boolean indicating if the formula was satisfiable.
// If it was, a model is then provided, associating each variable name with its binding.
func Solve(f Formula) (sat bool, model map[string]bool, err error) {
	return asCnf(f).solve()
}

// Dimacs writes the DIMACS CNF version of the formula on w.
// It is useful so as to feed it to any SAT solver.
// The original names of variables is associated with their DIMACS integer counterparts
// in comments, between the prolog and the set of clauses.
// For instance, if the variable "a" is associated with the index 1, there will be a comment line
// "c a=1".
func Dimacs(f Formula, w io.Writer) error {
	cnf := asCnf(f)
	nbVars := len(cnf.vars.all)
	nbClauses := len(cnf.clauses)
	prefix := fmt.Sprintf("p cnf %d %d\n", nbVars, nbClauses)
	if _, err := io.WriteString(w, prefix); err != nil {
		return fmt.Errorf("could not write DIMACS output: %v", err)
	}
	var pbVars []string
	for v := range cnf.vars.pb {
		pbVars = append(pbVars, string(v))
	}
	sort.Sort(sort.StringSlice(pbVars))
	for _, v := range pbVars {
		idx := cnf.vars.pb[variable(v)]
		line := fmt.Sprintf("c %s=%d\n", v, idx)
		if _, err := io.WriteString(w, line); err != nil {
			return fmt.Errorf("could not write DIMACS output: %v", err)
		}
	}
	for _, clause := range cnf.clauses {
		strClause := make([]string, len(clause))
		for i, lit := range clause {
			strClause[i] = strconv.Itoa(lit)
		}
		line := fmt.Sprintf("%s 0\n", strings.Join(strClause, " "))
		if _, err := io.WriteString(w, line); err != nil {
			return fmt.Errorf("could not write DIMACS output: %v", err)
		}
	}
	return nil
}

// Var generates a named boolean variable in a formula.
func Var(name string) Formula {
	return variable(name)
}

type variable string

func (v variable) nnf() Formula {
	return lit{signed: false, v: v}
}

func (v variable) String() string {
	return string(v)
}

type lit struct {
	signed bool
	v      variable
}

func (l lit) nnf() Formula {
	return l
}

func (l lit) String() string {
	if l.signed {
		return "not(" + string(l.v) + ")"
	}
	return string(l.v)
}

// Not represents a negation. It negates the given subformula.
func Not(f Formula) Formula {
	return not{f}
}

type not [1]Formula

func (n not) nnf() Formula {
	switch f := n[0].(type) {
	case variable:
		l := f.nnf().(lit)
		l.signed = true
		return l
	case lit:
		f.signed = !f.signed
		return f
	case not:
		return f[0].nnf()
	case and:
		subs := make([]Formula, len(f))
		for i, sub := range f {
			subs[i] = not{sub}.nnf()
		}
		return or(subs).nnf()
	case or:
		subs := make([]Formula, len(f))
		for i, sub := range f {
			subs[i] = not{sub}.nnf()
		}
		return and(subs).nnf()
	default:
		panic("invalid formula type")
	}
}

func (n not) String() string {
	return "not(" + n[0].String() + ")"
}

// And generates a conjunction of subformulas.
func And(subs ...Formula) Formula {
	return and(subs)
}

type and []Formula

func (a and) nnf() Formula {
	var res and
	for _, s := range a {
		nnf := s.nnf()
		if a2, ok := nnf.(and); ok { // // Simplify: "and"s in the "and" get to the higher level
			res = append(res, a2...)
		} else {
			res = append(res, nnf)
		}
	}
	if len(res) == 1 {
		return res[0]
	}
	return res
}

func (a and) String() string {
	strs := make([]string, len(a))
	for i, f := range a {
		strs[i] = f.String()
	}
	return "and(" + strings.Join(strs, ", ") + ")"
}

// Or generates a disjunction of subformulas.
func Or(subs ...Formula) Formula {
	return or(subs)
}

type or []Formula

func (o or) nnf() Formula {
	var res or
	for _, s := range o {
		nnf := s.nnf()
		if o2, ok := nnf.(or); ok { // Simplify: "or"s in the "or" get to the higher level
			res = append(res, o2...)
		} else {
			res = append(res, nnf)
		}
	}
	if len(res) == 1 {
		return res[0]
	}
	return res
}

func (o or) String() string {
	strs := make([]string, len(o))
	for i, f := range o {
		strs[i] = f.String()
	}
	return "or(" + strings.Join(strs, ", ") + ")"
}

// Implies indicates a subformula implies another one.
func Implies(f1, f2 Formula) Formula {
	return or{not{f1}, f2}
}

// Eq indicates a subformula is equivalent to another one.
func Eq(f1, f2 Formula) Formula {
	return and{or{not{f1}, f2}, or{f1, not{f2}}}
}

// Xor indicates exactly one of the two given subformulas is true.
func Xor(f1, f2 Formula) Formula {
	return and{or{not{f1}, not{f2}}, or{f1, f2}}
}

// Unique indicates exactly one of the given variables must be true.
func Unique(vars ...variable) Formula {
	res := make([]Formula, 1, 1+(len(vars)*len(vars)-1)/2)
	varsAsForms := make([]Formula, len(vars))
	for i, v := range vars {
		varsAsForms[i] = v
	}
	res[0] = Or(varsAsForms...)
	for i := 0; i < len(vars)-1; i++ {
		for j := i + 1; j < len(vars); j++ {
			res = append(res, Or(Not(vars[i]), Not(vars[j])))
		}
	}
	return And(res...)
}

// vars associate variable names with numeric indices.
type vars struct {
	all map[variable]int // all vars, including those created when converting the formula
	pb  map[variable]int // Only the vars that appeared orinigally in the problem
}

// litValue returns the int value associated with the given problem var.
// If the var was not referenced yet, it is created first.
func (vars *vars) litValue(l lit) int {
	val, ok := vars.all[l.v]
	if !ok {
		val = len(vars.all) + 1
		vars.all[l.v] = val
		vars.pb[l.v] = val
	}
	if l.signed {
		return -val
	}
	return val
}

// Dummy creates a dummy variable and returns its associated index.
func (vars *vars) dummy() int {
	val := len(vars.all) + 1
	vars.all[variable(fmt.Sprintf("dummy-%d", val))] = val
	return val
}

// A CNF is the representation of a boolean formula as a conjunction of disjunction.
// It can be solved by a SAT solver.
type cnf struct {
	vars    vars
	clauses [][]int
}

// solve solves the given formula.
// cnf is given to gophersat.
// The function returns a boolean indicating if the formula was satisfiable.
// If it was, a model is then provided, associating each variable name with its binding.
func (cnf *cnf) solve() (sat bool, vars map[string]bool, err error) {
	pb, err := solver.ParseSlice(cnf.clauses)
	if err != nil {
		return false, nil, fmt.Errorf("could not create problem from formula: %v", err)
	}
	s := solver.New(pb)
	if s.Solve() != solver.Sat {
		return false, nil, nil
	}
	m, err := s.Model()
	if err != nil {
		return false, nil, fmt.Errorf("could not retrieve model: %v", err)
	}
	vars = make(map[string]bool)
	for name, idx := range cnf.vars.pb {
		vars[string(name)] = m[idx-1]
	}
	return true, vars, nil
}

// asCnf returns a CNF representation of the given formula.
func asCnf(f Formula) *cnf {
	vars := vars{all: make(map[variable]int), pb: make(map[variable]int)}
	clauses := cnfRec(f.nnf(), &vars)
	return &cnf{vars: vars, clauses: clauses}
}

// transforms the f NNF formula into a CNF formula.
// nbDummies is the current number of dummy variables created.
// Note: code should be improved, there are a few useless allocs/deallocs
// here and there.
func cnfRec(f Formula, vars *vars) [][]int {
	switch f := f.(type) {
	case lit:
		return [][]int{{vars.litValue(f)}}
	case and:
		var res [][]int
		for _, sub := range f {
			res = append(res, cnfRec(sub, vars)...)
		}
		return res
	case or:
		var res [][]int
		var lits []int
		for _, sub := range f {
			switch sub := sub.(type) {
			case lit:
				lits = append(lits, vars.litValue(sub))
			case and:
				d := vars.dummy()
				lits = append(lits, d)
				for _, sub2 := range sub {
					nnf := cnfRec(sub2, vars)[0]
					nnf = append(nnf, -d)
					res = append(res, nnf)
				}
			default:
				panic("unexpected or in or")
			}
		}
		res = append(res, lits)
		return res
	default:
		panic("invalid NNF formula")
	}
}
