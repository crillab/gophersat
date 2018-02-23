package optim

import (
	"fmt"
	"math/rand"
	"testing"
)

func TestUnsat(t *testing.T) {
	pb := New(
		HardClause(Var("a"), Var("b"), Var("c")),
		HardClause(Var("a"), Var("b"), Not("c")),
		HardClause(Var("a"), Not("b"), Var("c")),
		HardClause(Var("a"), Not("b"), Not("c")),
		HardClause(Not("a"), Var("b"), Var("c")),
		HardClause(Not("a"), Var("b"), Not("c")),
		HardClause(Not("a"), Not("b"), Var("c")),
		HardClause(Not("a"), Not("b"), Not("c")),
	)
	if model, cost := pb.Solve(); model != nil {
		t.Errorf("expected unsat, got model %v, cost %d", model, cost)
	}
}

func TestSat(t *testing.T) {
	pb := New(
		HardClause(Var("a"), Var("b"), Var("c")),
		HardClause(Var("a"), Var("b"), Not("c")),
		HardClause(Var("a"), Not("b"), Var("c")),
		HardClause(Var("a"), Not("b"), Not("c")),
		HardClause(Not("a"), Var("b"), Var("c")),
		HardClause(Not("a"), Var("b"), Not("c")),
		HardClause(Not("a"), Not("b"), Not("c")),
	)
	if model, cost := pb.Solve(); model == nil {
		t.Errorf("expected sat, got unsat")
	} else if !model["a"] || !model["b"] || model["c"] {
		t.Errorf("invalid model, got %v", model)
	} else if cost != 0 {
		t.Errorf("invalid cost, expected 0, got %d", cost)
	}
}

func TestOptim(t *testing.T) {
	pb := New(
		HardClause(Var("a"), Var("b"), Var("c")),
		HardPBConstr([]Lit{Not("a"), Not("b"), Not("c")}, []int{1, 1, 1}, 2),
		SoftPBConstr([]Lit{Var("a"), Var("b"), Var("c")}, []int{1, 1, 1}, 2),
		WeightedClause([]Lit{Not("a"), Var("d")}, 2),
		WeightedPBConstr([]Lit{Var("b"), Var("c"), Var("d")}, []int{1, 1, 1}, 2, 3),
		SoftClause(Not("c"), Not("d")),
	)
	if model, cost := pb.Solve(); model == nil {
		t.Errorf("expected sat, got unsat")
	} else if model["a"] || !model["b"] || model["c"] || !model["d"] {
		t.Errorf("invalid model, got %v", model)
	} else if cost != 1 {
		t.Errorf("invalid cost, expected 1, got %d", cost)
	}
}

// A coord is the coordinates for a city in a TSP problem.
type coord struct {
	line int
	col  int
}

// generateTSP generates a representation for the TSP problem.
func generateTSP(nbCities int) []Constr {
	var constrs []Constr
	coords := make([]coord, nbCities)
	for i := range coords {
		coords[i] = coord{line: rand.Intn(2 * nbCities), col: rand.Intn(2 * nbCities)}
	}
	distTo := make([][]int, nbCities)
	for i := range distTo {
		distTo[i] = make([]int, nbCities)
		for j := 0; j < nbCities; j++ {
			if j != i {
				distLine := coords[i].line - coords[j].line
				if distLine < 0 {
					distLine = -distLine
				}
				distCol := coords[i].col - coords[j].col
				if distCol < 0 {
					distCol = -distCol
				}
				distTo[i][j] = distLine + distCol // Manhattan distance
			}
		}
	}
	format := "city-%d-step-%d"
	for i := range coords {
		lits := make([]Lit, nbCities)
		negs := make([]Lit, nbCities)
		for j := range coords {
			lits[j] = Var(fmt.Sprintf(format, i, j))
			negs[j] = Not(fmt.Sprintf(format, j, i))
		}
		constrs = append(constrs, HardClause(lits...))                 // Each city is visited at least once
		constrs = append(constrs, HardPBConstr(negs, nil, nbCities-1)) // At each step, at most one city is visited
	}
	for i := 0; i < nbCities-1; i++ {
		for j := i + 1; j < nbCities; j++ {
			for step := 0; step < nbCities-1; step++ {
				// at any step, going from i to j or from j to i has a cost equal to the distance
				constrs = append(constrs, WeightedClause([]Lit{Not(fmt.Sprintf(format, i, step)), Not(fmt.Sprintf(format, j, step+1))}, distTo[i][j]))
				constrs = append(constrs, WeightedClause([]Lit{Not(fmt.Sprintf(format, i, step+1)), Not(fmt.Sprintf(format, j, step))}, distTo[i][j]))
			}
		}
	}
	// start with city0
	constrs = append(constrs, HardClause(Var(fmt.Sprintf(format, 0, 0))))
	return constrs
}

func TestTSP(t *testing.T) {
	constrs := generateTSP(9)
	pb := New(constrs...)
	if model, cost := pb.Solve(); model == nil {
		t.Errorf("expected sat, got unsat")
	} else if cost == 0 {
		t.Errorf("invalid cost, expected non null, got %d", cost)
	}
}

func BenchmarkTSP(b *testing.B) {
	for i := 0; i < b.N; i++ {
		New(generateTSP(9)...).Solve()
	}
}
