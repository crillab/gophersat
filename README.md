# Gophersat, a SAT and pseudo-boolean solver written in Go

![Status](https://img.shields.io/badge/status-stable-green.svg?style=plastic)
[![GitHub tag](https://img.shields.io/github/tag/crillab/gophersat.svg)](https://github.com/crillab/gophersat)
![license](https://img.shields.io/github/license/crillab/gophersat.svg)
[![GoReport](https://goreportcard.com/badge/github.com/crillab/gophersat)](https://goreportcard.com/report/github.com/crillab/gophersat)
[![GoDoc](https://godoc.org/github.com/crillab/gophersat?status.svg)](https://godoc.org/github.com/crillab/gophersat)
[![Build Status](https://travis-ci.org/crillab/gophersat.svg?branch=master)](https://travis-ci.org/crillab/gophersat)

![gophersat logo](https://raw.githubusercontent.com/crillab/gophersat/master/gophersat.png)

This is Gophersat, a SAT and pseudo-boolean solver written purely in Go. 
Gophersat was developed by the [CRIL (Centre de Recherche en Informatique
de Lens)](http://www.cril.fr) at the Artois University & CNRS. It is
released under the MIT license. Gophersat is rather efficient, i.e on
typical SAT benchmarks it runs about 2 to 5 times slower than top-level
solvers (namely, [glucose](http://www.labri.fr/perso/lsimon/glucose/) or
[minisat](http://minisat.se/)) from which it is strongly inspired.
It can also solve MAXSAT problems, and pseudo-boolean decision and optimization problems.


## Version 1.3

Gophersat's last stable version is version 1.3. It is a minor update, adding the ability to access the underlying solver when dealing with MAXSAT problems.


## Version 1.2: Explainable AI: UNSAT certification, MUS extraction

Gophersat version 1.2 includes a way to understand UNSAT instances, both by providing RUP certificates when a problem is UNSAT and by providing the ability to extract unsatisfiable subsets of the formula. A vew bugs were also corected, and the support for incremental SAT solving was improved.

To learn more about these functionalities, you can check the [tutorial about UNSAT certificates and MUSes](mus.md).

To generate a certificate with the gophersat executable, simply call:

    gophersat -certified problem.cnf

The certificate will then be printed on the standard output, using the RUP notation. The certificate is generated on the fly, so be aware that a partial, useless certificate will be generated even if the problem is actually satisfiable. This is common practice in the community, and although the generated clauses are useless noise, in practice this is not a problem.

To extract a MUS from an UNSAT instance, just call:

    gophersat -mus problem.cnf

The MUS will the be printed on the standard output. If the problem is not UNSAT, an error message will be displayed.

For the moment, these facilities are only available for pure SAT problems (i.e not pseudo-boolean problems).


## Version 1.1

Since its version 1.1, Gophersat includes a new, more efficient core solver for pure SAT problems
and a package dealing with MAXSAT problems. It also includes a new API for optimization and model counting,
where new models are written to channels as soon as they are found.

### About version numbers in Gophersat

Since version 1.0, Gophersat's API is considered stable, meaning that the API is guaranteed to stay backwards-compatible
until a major version shift. In other words, if your program works with version 1.0 of gophersat, it
will still work with version 1.1 or above, but not necessarily with versions above 2.0.

Note that, by "still working", we only mean "will compile and produce the same output", not "will have the
same performance memory-wise or time-wise". This is an important distinction: during minor version upgrades,
new heuristics or data types can be introduced, meaning some given problems could be solved faster or slower
than previously.

## How do I install it?

`go get github.com/crillab/gophersat && go install
github.com/crillab/gophersat`

### Solving SAT problems

Gophersat can be used as a standalone solver (reading DIMACS CNF files) or as a library in any go program.

To solve a DIMACS file, you can call gophersat with the following syntax:

    gophersat --verbose file.cnf

where `--verbose` is an optional parameters that makes the solver display informations during the solving process.

Gophersat is also able to read and solve more general boolean formulas,
not only problems represented in the user-unfriendly DIMACS format.
It also deals natively with cardinality constraints, i.e clauses that must have at least
n literals true, with n > 1.

### Solving pseudo-boolean problems

Gophersat can be used as a standalone solver (reading OPB files) or as a library in any go program.

To solve a pseudo-boolean problem (whether a decision one or an optimisation one), you can call gophersat with the
following syntax:

    gophersat --verbose file.opb

where `--verbose` is an optional parameters that makes the solver display informations during the solving process.

For the moment, it can only solve the so-called DEC-SMALLINT-LIN problems and OPT-SMALLINT-LIN,
i.e decision problems (is there a solution or not), for linear constraints (a sum of weighted literals)
on small integers (n < 2^30), or optimization problems (what is the best solution, minimizing a given cost function),
for linear constraints on small integers.

### Solving MAXSAT problems

Thanks to the `maxsat`package, Gophersat can now solve MAXSAT problems.

To solve a weighted MAXSAT problem, you can call gophersat with the following syntax:

    gophersat --verbose file.wcnf

where `--verbose` is an optional parameters that makes the solver display informations during the solving process.
The file is supposed to be represented in (the WCNF format)[http://www.maxsat.udl.cat/08/index.php?disp=requirements].

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

## What is MAXSAT?

MAXSAT is the optimization equivalent of the SAT decision problem. While a pure SAT solver will either return a
model satisfying all clauses or UNSAT if no such model exists, a MAXSAT solver will return a model that satisfies
as many clauses as possible.

### Partial and weighted MAXSAT

There are extensions to MAXSAT.

**Partial MAXSAT** means that, although we want to satisfy as many clauses as possible,
some clauses (called *hard clauses*) *must* be satisfied, not matter what. A partial MAXSAT problem can thus be
declared unsatisfiable.

For instance, generating a timetable for a school is a partial MAXSAT problem: there are both soft (we want to have as little classes as possible that start after 4 PM, for instance)
and hard (two teachers cannot be in two different places at the same time) constraints.

**Weighted MAXSAT** means that clauses are associated with a cost: although optional, some clauses are deemed more important than others. For instance, if clause C1 has a cost of 3 and clauses C2 and C3 both have a cost of 1, a solution satisfying C1 but neither C2 nor C3 will be considered better than a solution satisfying both C2 and C3 but not C1, all other things being equal.

**Partial weighted MAXSAT** means that there are both soft and hard clauses in the problem, and soft clauses are weighted. In this regard, "pure" MAXSAT is a particular case of the more generic partial weighted MAXSAT: a "pure" MAXSAT problem is a partial weighted MAXSAT problem where all clauses are soft clauses associated with a weight of 1.

## What about pseudo-boolean problems?
Pseudo-boolean problems are, in a way, a generalization of SAT problems: any propositional clause
can be written as a single pseudo-boolean constraint, but representing a pseudo-boolean constraint
can require an exponential number of propositional clauses.

A (linear) pseudo-boolean expression is an expression like:

    w1 l1 + w2 x2 + ... + wn xn ≥ y

where `y` and all `wi` are integer constants and all `li` are boolean literals.

For instance, for the following expression to be true

    2 ¬x1 + x2 + x3 ≥ 3

the boolean variable `x1` must be false, and at least one of `x2` and `x3` must be true.
This is equivalent to the propositional formula

    ¬x1 ∧ (x2 ∨ x3)

Pseudo-boolean expressions are a generalization of propositional clauses: any clause

    x1 ∨ x2 ∨ ... ∨ xn

can be rewritten as

    x1 + x2 + ... + xn ≥ 1

The description of a pseudo-boolean problem can be exponentially smaller than its
propositional counterpart, but solving a psudo-boolean problem is still NP-complete.

Gophersat solves both decision problems (is there a solution at all to the problem),
and optimization problem.
An optimization problem is a pseudo-boolean problem associated with a cost function,
i.e a sum of terms that must be minimized.
Rather than just trying to find a model that satisfies the given constraints,
gophersat will then try to find a model that is guaranteed to both satisfy the constraints
and minimize the cost function.

During the search for that optimal solutions, gophersat will provide suboptimal solutions
(i.e solutions that solve all the constraints but are not yet guaranteed to be optimal)
as soon as it finds them, so the user can either get a suboptimal solution in a given time limit,
or wait longer for the guaranteed optimal solution.

## Is Gophersat fast? Why use it at all?
Yes and no. It is much faster than naïve implementations, fast enough to be used on real-world problems, but slower than
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

There are a few other SAT solvers in Go, mainly go-sat and gini.
Gini's performance is pretty much on par with Gophersat, although a little slower
on average, according to tests we ran.

Gophersat is also providing cool features not always available with other solvers
(a user-friendly input format, for instance), so it can be used as a tool for
describing and solving NP-hard problems that can easily be reduced to a SAT/PB problem.

## Do I have to represent my SAT problem as CNF? Am I forced to use the unfriendly DIMACS format?
No. The `bf` (for "boolean formula") package provides facilities to
translate any boolean formula to CNF.

## Can I know how many solutions there are for a given formula?
This is known as model counting, and yes, there is a function for that: `solver.Solver.CountModels`.

You can also count models from command line, with

    gophersat --count filename

where filename can be a .opb or a .cnf file.
