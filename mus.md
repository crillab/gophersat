# Understanding UNSAT certificates and MUSes

Since version 1.2, gophersat provides facilities to understand UNSAT instances.

## What is an UNSAT instance?

When given a logical formula, one usually wants to find a way to make the logical formula true (a "model"). For instance, the formula

    (x1 ∨ x2 ∨ x3) ∧ (¬x1 ∨ ¬x2)

can be made true in several ways: for instance, if x1 is true and both x2 and x3 are false, the formula will be true.

But some formulas cannot be made true no matter what. For instance, the following formula

    (x1 ∨ x2 ∨ x3) ∧
    (¬x2 ∨ ¬x3 ∨ ¬x4) ∧
    (¬x1 ∨ ¬x3) ∧
    (¬x1 ∨ x3) ∧
    (x1 ∨ ¬x2) ∧
    (x1 ∨ x2)

cannot be true. This in an unsatisfiable formula, or UNSAT for short. For the remainder of this document, I'll adopt the DIMACS notation, so the variable x1
will be written as 1, each line is a clause, aka a disjunction of literal, and the whole problem is a conjunction of clauses.
Finally, each clause (each line) will end with the symbol 0 (for legacy reasons).
So, the above problem will be written as:

    1 2 3 0
    -2 -3 -4 0
    -1 -3 0
    -1 3 0
    1 -2 0
    1 2 0

## UNSAT certificates

When a SAT solver finds a model for a problem (i.e a way to make the formula true), it is easy to check whether the solution is correct or not:
just take the model, take the problem, and check whether the formula is indeed true or not. This is an easy problem that can be coded very easily and can be computed in polynomial time.

For the first formula above, if I tell you "the formula is true when x1 is true and both x2 and x3 are false", you can easily check whether it's true or not.

But when no model can be found, the problem is unsatisfiable (or UNSAT for short), and the user has to trust the solver. That situation has two issues:

- What if there's a bug in the solver? SAT solvers are complex pieces of software and can contain subtle bugs, and gophersat is no exception.
- What if I want an explanation? *Why* is the problem UNSAT?

The solution adopted by the community is to use certificates, using the RUP notation or one of its extensions.
Certificates consist in a sequence of clauses that can be deduced from the original formula. For instance, a certificate for the
UNSAT formula above could be:

    1 0
    0

What does that mean? That means that the literal "x1" can be deduced from the formula (in other words, it can be proven that "x1" has to be true), then that once "x" is set to true, it can be trivially proven the formula has to be false.

How do we check that certificate? The certificate says "x1" has to be true, so let's see what happens when "x1" is false.

    1 2 3 0
    -2 -3 -4 0
    -1 -3 0
    -1 3 0
    1 -2 0
    1 2 0
    -1 0

becomes (we remove clauses containing -1 and remove the literal 1 in the clauses containing it)

    2 3 0
    -2 -3 -4 0
    -2 0
    2 0

"x2" is both true and false: the formula is false. Therefore, we can indeed conclude that "x1" has to be true, and add the clause "1 0" that was given by the certificate. So the formula can be rewritten as:

    1 2 3 0
    -2 -3 -4 0
    -1 -3 0
    -1 3 0
    1 -2 0
    1 2 0
    1 0

After simplification (removing clauses containing 1 and removing the lit 1 where it appears), the formula can be rewritten as

    -2 -3 -4 0
    -3 0
    3 0
    -2 0
    2 0

Once again, this is false. Hence, the formula is UNSAT.

UNSAT certificates are a useful tool, and they are cheap to produce. Be careful, though: since SAT solving is an NP hard problem, and since certificates are a trace of the solving process, they can be huge, orders of magnitude bigger than the original formula.


## Minimal Unsatisfiable Subformula (MUS)

Knowing a formula is unsatisfiable, even with a certificate, is not always sufficient to understand, as a human being, *why* it is so. For instance, the above UNSAT formula is more complex than needeed. The formula


    1 2 3 0
    -2 -3 -4 0
    -1 -3 0
    -1 3 0
    1 -2 0
    1 2 0

is UNSAT because its subformula

    -2 -3 -4 0
    -1 -3 0
    -1 3 0
    1 -2 0
    1 2 0

is also UNSAT. Notice how we removed the first clause. And this subformula, being smaller, is easier to understand. Now, the following subformula is *even* shorter, and still UNSAT:

    -1 -3 0
    -1 3 0
    1 -2 0
    1 2 0

Can we make that subformula shorter? No. This subformula is UNSAT, but removing any of these clauses would make the formula satisfiable. Therefore, **that subformula is a MUS**. MUSes are great, because they let us focus on what is "wrong" with the original formula and, if that makes sense, try to fix it. And, the good news is, the certificate used on the original formula also certifies the MUS.

However, there are a few things to know.

First, **computing a MUS is expensive**. To compute a MUS, one must call the SAT solver several times. There are several algorithms to do so, but in Gophersat the MUS is found by checking whether each clause is required to make the problem UNSAT or not. That means the SAT solver is called *n* times, where *n* is the number of clauses in the original formula. The example formula contains 6 clauses, so the SAT solver will be called 6 times. Since each call runs an NP-complete algorithm, finding a MUS can be extremely long in practice, and untractable in many extreme cases.

Another thing is, there can be several MUSes in a problem, and finding *a* MUS doesn't mean you found *the shortest of all MUSes*. Let's take this problem as an example:

    1 2 0
    1 -2 0
    -1 0
    3 0
    -3 0

This problem is UNSAT. The following subformula is UNSAT, and removing any of its clause would make it SAT. It is, by definition, a MUS:

    1 2 0
    1 -2 0
    -1 0

But the following subformula is also a MUS, and it is even shorter:

    3 0
    -3 0

Finding the first MUS does not give you any guarantee there isn't a shorter MUS. In fact, you can't even know how many different MUSes a problem contains once you found one. Finding all the MUSes of a problem, and finding the smallest of them, is extremely complex (one way would be to call the SAT solver an exponential number of times), so this is currently not possible with Gophersat.
