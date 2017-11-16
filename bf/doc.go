// Package bf offers facilities to test the satisfiability of generic boolean formula.
//
// SAT solvers usually expect as an input CNF formulas.
// A CNF, or Conjunctive Normal Form, is a set of clauses that must all be true, each clause
// being a set of potentially negated literals. For the clause to be true, at least one of
// these literals must be true.
//
// However, manually translating a given boolean formula to an equivalent CNF is tedious and error-prone.
// This package provides a set of logical connectors to define and solve generic logical formulas.
// Those formulas are then automatically translated to CNF and passed to the gophersat solver.
//
// For example, the following boolean formula:
//
//     ¬(a ∧ b) → ((c ∨ ¬d) ∧ ¬(c ∧ (e ↔ ¬c)) ∧ ¬(a ⊻ b))
//
// Will be defined with the following code:
//
//     f := Not(Implies(And(Var("a"), Var("b")), And(Or(Var("c"), Not(Var("d"))), Not(And(Var("c"), Eq(Var("e"), Not(Var("c"))))), Not(Xor(Var("a"), Var("b"))))))
//
// When calling `Solve(f)`, the following equivalent CNF will be generated:
//
//     a ∧ b ∧ (¬c ∨ ¬x1) ∧ (d ∨ ¬x1) ∧ (c ∨ ¬x2) ∧ (¬e ∨ ¬c ∨ ¬x2) ∧ (e ∨ c ∨ ¬x2) ∧ (¬a ∨ ¬b ∨ ¬x3) ∧ (a ∨ b ∨ ¬x3) ∧ (x1 ∨ x2 ∨ x3)
//
// Note that this formula is longer than the original one and that some variables were added to it.
// The translation is both polynomial in time and space.
// When fed this CNF as an input, gophersat then returns the following map:
//
//     map[a:true b:true c:false d:true e:false]
//
// It is also possible to create boolean formulas using a dedicated syntax. The BNF grammar is as follows:
//
//    formula ::= formula { ';' clause }*
//    clause  ::= implies { '=' implies }*
//    implies ::= or { '->' or}*
//    or      ::= and { '|' and}*
//    and     ::= not { '&' not}*
//    not     ::= '^'not | atom
//    atom    ::= ident | '(' formula ')'
//
// So the formula
//
//     ¬(a ∧ b) → ((c ∨ ¬d) ∧ ¬(c ∧ (e ↔ ¬c)) ∧ ¬(a ⊻ b))
//
// would be written as
//
//    ^(a & b) -> ((c | ^d) & ^(c & (e = ^c)) & ^(a = ^b))
//
// a call to the `Parse` function will then create the associated Formula.
package bf
