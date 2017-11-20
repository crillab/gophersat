/*
Package solver gives access to a simple SAT and pseudo-boolean solver.
Its input can be either a DIMACS CNF file, a PBS file or a solver.Problem object,
containing the set of clauses to be solved. In the last case,
the problem can be either a set of propositional clauses,
or a set of pseudo-boolean constraints.

No matter the input format,
the solver.Solver will then solve the problem and indicate whether the problem is
satisfiable or not. In the former case, it will be able to provide a model, i.e a set of bindings
for all variables that makes the problem true.

Describing a problem

A problem can be described in several ways:

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
    pb := solver.ParseSlice(clauses)

3. create a list of cardinality constraints (CardConstr), if the problem to be solved is better represented this way.
For instance, the problem stating that at least two literals must be true among the literals 1, 2, 3 and 4 could be described as a set of clauses:

    clauses := [][]int{
        []int{1, 2, 3},
        []int{2, 3, 4},
        []int{1, 2, 4},
        []int{1, 3, 4},
    }
    pb := solver.ParseSlice(clauses)

The number of clauses necessary to describe such a constrain can grow exponentially. Alternatively, it is possible to describe the same this way:

    constrs := []CardConstr{
        {Lits: []int{1, 2, 3, 4}, AtLeast: 2},
    }
    pb := solver.ParseCardConstrs(clauses)

Note that a propositional clause has an implicit cardinality constraint of 1, since at least one of its literals must be true.

4. parse a PBS stream (io.Reader). If the io.Reader contains the following problem:

    2 ~x1 +1 x2 +1 x3 >= 3 ;

the programmer can create the Problem by doing:

    pb, err := solver.ParsePBS(f)

5. create a list of PBConstr. For instance, the following set of one PBConstrs will generate the same problem as above:

    constrs := []PBConstr{GtEq([]int{1, 2, 3}, []int{2, 1, 1}, 3)}
    pb := solver.ParsePBConstrs(constrs)

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

For the above problem described in the DIMACS format, the output can be:

    SATISFIABLE
    -1 2 -3 4 -5 -6

*/
package solver
