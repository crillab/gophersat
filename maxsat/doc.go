// Package maxsat provides an optimization solver for SAT/PB.
// It allows the user to provide weighted partial MAXSAT problems or weighted pseudo-booleans problems.
//
// Definition
//
// A MAXSAT problem is a problem where, contrary to "plain-old" SAT decision problems,
// the user is not looking at whether the problem can be solved at all, but, if it cannot be solved,
// if at least a subset of it can be solved, with that subset being as big as important.
// In other words, the MAXSAT solver is trying to find a model that satisfies as many clauses as possible,
// ideally all of them.
//
// Pure MAXSAT is not very useful in practice. Generally, the user wants to add two more constraints :
// - a subset of the problem must be satisfied, no matter what; these are called *hard clauses*,
// - other clauses (called *soft clauses*) are optional, but some of them are deemed more important than
// others: they are associated with a cost.
//
// That problem is called weighted partial MAXSAT (WP-MAXSAT). Note that MAXSAT is a special case of WP-MAXSAT
// where all the clauses are soft clauses of weight 1. Also note that the traditional, SAT decision problem
// is a special case of WP-MAXSAT where all clauses are hard clauses.
//
// Gophersat is guaranteed to provide the best possible solution to any WP-MAXSAT problem, if given enough time.
// It will also give potentially suboptimal solutions as soon as it finds them.
// So, the user can either get a good-enough solution after a given amount of time, or wait as long as needed
// for the best possible solution.
package maxsat
