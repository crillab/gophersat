/*
Package solver gives access to a simple SAT solver.
Its input can be either a DIMACS CNF file or a solver.Problem object,
containing the set of clauses to be solved.

The solver.Solver will then solve the problem and indicate whether the problem is
satisfiable or not. In the former case, it will be able to provide a model, i.e a set of bindings
for all variables that makes the problem true.

Describing a problem

A problem can be described in two ways:

1. parse a DIMACS stream (io.Reader). If the io.Reader produces the following content:

    p cnf 6 7
    1 2 3 0
    4 5 6 0
    -1 -4 0
    -2 -5 0
    -3 -6 0
    -1 -3 0
    -4 -6 0

the programmer can create the Problem by doing:

    pb, err := solver.ParseCNF(f)

2. create the equivalent list of list of literals. The problem above can be created programatically this way:

    clauses := [][]int{
        []int{1, 2, 3},
        []int{4, 5, 6},
        []int{-1, -4},
        []int{-2, -5},
        []int{-3, -6},
        []int{-1, -3},
        []int{-4, -6},
    }
    pb, err := solver.ParseSlice(clauses)

Solving a problem

To solve a problem, one simply creates a solver with said problem.
The solve() method then solves the problem and returns the corresponding status: Sat or Unsat.

    s := solver.New(pb)
    status := s.Solve()

If the status was Sat, the programmer can ask for a model, i.e an assignment that makes all the clauses of the problem true:

    m, err := s.Model()

For the above problem, the status will be Sat and the model can be {false, true, false, true, false, false}.

Alternatively, one can display on stdout the result and model (if any):

    s.OutputModel()

For the above problem, the output can be:

    SATISFIABLE
    -1 2 -3 4 -5 -6

*/
package solver
