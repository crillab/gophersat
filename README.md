# Gophersat, a SAT solver written in pure Go

![gophersat logo](https://github.com/crillab/gophersat/gophersat.png)

This is Gophersat, a SAT solver written purely in Go. 
Gophersat was developed by the [CRIL (Centre de Recherche en Informatique
de Lens)](http://www.cril.fr) at the Artois University & CNRS. It is
released under the GNU LGPL license. Gophersat is rather efficient, i.e on
typical benchmarks it runs about 1 to 5 times slower than top-level
solvers (namely, [glucose](http://www.labri.fr/perso/lsimon/glucose/) or
[minisat](http://minisat.se/)) from which it is strongly inspired.


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

## Is Gophersat fast? Why use it at all?
Yes and no. It is much faster than naïve implementations, but slower than
top-level, highly optimised, state-of-the-art solvers.

Gophersat is not aiming at being the fastest SAT solver available. The
goal is to give access to SAT technologies to gophers, without resorting
to interfaces to solvers written in other languages (typically C/C++).

However, in some cases, interfacing with a solver written in another
language is the best choice for your application. If you have lots of
small problems to solve, Gophersat will be the fastest alternative. For
bigger problems, if you want to have as little dependencies as possible
at the expense of solving time, Gophersat is good too. If you need to
solve difficult problems and don't mind using cgo or use an external
program, Gophersat is probably not the best option.

## What does it feature?
For the moment, almost nothing. It does not
propose incremental SAT solving, MUS extraction, UNSAT certification,
model counting, user-friendly input format and the like. It does not even
feature a preprocessor.

But all those features might come later. Feel free to ask for features. 
