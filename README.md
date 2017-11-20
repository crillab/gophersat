# Gophersat, a SAT and pseudo-boolean solver written in Go

[![GoReport](https://goreportcard.com/badge/github.com/crillab/gophersat)](https://goreportcard.com/report/github.com/crillab/gophersat)
[![GoDoc](https://godoc.org/github.com/crillab/gophersat?status.svg)](https://godoc.org/github.com/crillab/gophersat)

![gophersat logo](https://raw.githubusercontent.com/crillab/gophersat/master/gophersat.png)

This is Gophersat, a SAT and pseudo-boolean solver written purely in Go. 
Gophersat was developed by the [CRIL (Centre de Recherche en Informatique
de Lens)](http://www.cril.fr) at the Artois University & CNRS. It is
released under the MIT license. Gophersat is rather efficient, i.e on
typical SAT benchmarks it runs about 1 to 5 times slower than top-level
solvers (namely, [glucose](http://www.labri.fr/perso/lsimon/glucose/) or
[minisat](http://minisat.se/)) from which it is strongly inspired.
It can also solve pseudo-boolean decision problems, although this is still
at an early development stage.

## How do I install it?

`go get github.com/crillab/gophersat && go install
github.com/crillab/gophersat`

### Solving SAT problems

Gophersat can be used as a standalone solver (reading DIMACS CNF files
from a file or standard input) or as a library in any go program.

Gophersat is also able to read and solve more general boolean formulas,
not only problems represented in the user-unfriendly DIMACS format.
It also deals natively with cardinality constraints, i.e clauses that must have at least
n literals true, with n > 1.

### Solving pseudo-boolean problems

Gophersat can be used as a standalone solver (reading PBS files from a file or
standatd input) or as a library in any go program.

For the moment, it can only solve the so-called DEC-SMALLINT-LIN problems,
i.e decision problems (is there a solution or not), for linear constraints (a sum of weighted literals)
on small integers (n < 2^30).
This will be improved in the future, though.

## What is a SAT solver? What is the SAT problem?
SAT, which stands for *Boolean Satisfiability Problem*, is the canonical
NP-complete problem, i.e a problem for which there is no known solution that does
not take exponential time in the worse case. In a few words, a SAT solver tries to find,
for a given propositional formula, an assignment
for all its variables that makes it true, if such an assignment exists.

While it's trivial to implement a SAT solver using a naïve algorithm, such
an implementation would be very inefficient in practice. Gophersat
implements state-of-the-art features and is thus quite efficient, making
it usable in practice in Go programs that need an efficient inference
engine.

Although this is not always the best decision for practical reasons, any
NP-complete problem can be translated as a SAT problem and solved by
Gophersat. Such problems include the Traveling Salesman Problem,
constraint programming, Knapsack problem, planning, model checking,
software correctness checking, etc.

More about the SAT problem can be found on [wikipedia's article about
SAT](https://en.wikipedia.org/wiki/Boolean_satisfiability_problem).

You can also find information about how to represent your own boolean
formulas so they can be used by gophersat in the [tutorial "SAT for
noobs"](examples/sat-for-noobs.md).

## What about pseudo-boolean problems?
Pseudo-boolean problems are a generalization of SAT problems: any propositional clause
can be written as a single pseudo-boolean constraint, but representing a pseudo-boolean constraint
can require an exponential number of propositional clauses.

A pseudo-boolean expression is an expression like:

    w1 l1 + w2 x2 + ... + wn xn ≥ y

where y and all wi are integer constants and all li are boolean literals.

For instance, for the following expression to be true

    2 ¬x1 + x2 + x3 ≥ 3

the boolean variable x1 must be false, and at least one of x2 and x3 must be true.
This is equivalent to the propositional formula

    ¬x1 ∧ (x2 ∨ x3)

Pseudo-boolean expressions are a generalization of propositional clauses: any clause

    x1 ∨ x2 ∨ ... ∨ xn

can be rewritten as

    x1 + x2 + ... + xn ≥ 1

The description of a pseudo-boolean problem can be exponentially smaller than its
propositional counterpart, but solving a psudo-boolean problem is still NP-complete.

## Is Gophersat fast? Why use it at all?
Yes and no. It is much faster than naïve implementations, but slower than
top-level, highly optimised, state-of-the-art solvers.

Gophersat is not aiming at being the fastest SAT/PB solver available. The
goal is to give access to SAT and PB technologies to gophers, without resorting
to interfaces to solvers written in other languages (typically C/C++).

However, in some cases, interfacing with a solver written in another
language is the best choice for your application. If you have lots of
small problems to solve, Gophersat will be the fastest alternative. For
bigger problems, if you want to have as little dependencies as possible
at the expense of solving time, Gophersat is good too. If you need to
solve difficult problems and don't mind using cgo or use an external
program, Gophersat is probably not the best option.

Gophersat is also providing cool features not always available with other solvers
(a user-friendly input format, for instance), so it can be used as a tool for
describing and solving NP-hard problems that can easily be reduced to a SAT instance.

## Do I have to represent my SAT problem as CNF? Am I forced to use the unfriendly DIMACS format?
Not anymore. The `bf` (for "boolean formula") package provides facilities to
translate any boolean formula to CNF.

## What else does it feature?
For the moment, not very much. It does not
propose incremental SAT solving, MUS extraction, UNSAT certification,
model counting  and the like. It does not even feature a preprocessor.

But all those features might come later. Feel free to ask for features. 
