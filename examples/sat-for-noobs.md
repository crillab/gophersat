# The SAT problem for noobs

This document gives an overview of the SAT problem, and how to use a SAT solver to solve a given problem.
It is intended for users of such solvers. It does not describe strategies, algorithms or implementation
of such solvers.
The intended audience for this document are developpers, so it assumes the reader understands boolean logic.

This document is work in progress. Expect it to be improved over time.

## What is the SAT problem?

SAT is a shortcut for "boolean satisfiability".
Imagine you have a **boolean formula**, i.e a formula containing **boolean variables**. For instance:

    (a and (b or not(c))) or (not(b) and c)

In this formula, `a`, `b` and `c` can either be true or false.
The SAT problem is the problem of finding an assignment for these variables so that the
whole formula becomes true, if such an assignment exists. This assignment is called a **model**, and a formula
can have zero, one or several models.

In the above problem, `a=false, b=true, c=false` is not a model of the formula, because

    (false and (true or not(false))) or (not(true) and false)

reduces to `false`.

However, the assignment `a=true, b=true, c=false` is a model, because

    (true and (true or not(false))) or (not(true) and false)

reduces to `true`.

Now, the trivial problem

    a and not(a)

has no model, because there is no way to make this formula true, no matter the value associated with `a`.
This formula is said to be **unsatisfiable** (UNSAT for short).

## What is a SAT solver?

A **SAT solver** is a piece of software that, when given a formula (more on the input syntax below),
returns either one model, if at least one model exists, or the indication that the problem is
*unsatisfiable*.

Implementing such a piece of software is rather easy, but there's one little problem: the SAT problem is NP-complete,
meaning all known algorithms to solve it take an exponential time (worse-case complexity is `O(k^n)`, with `k > 1`), and we're not even sure there is a way to make
a non-exponential algorithm (this is the famous "P = NP" problem).
So, the most trivial algorithm to solve the SAT problem (test all possible assignments for all variables, stop when one makes the formula true,
return UNSAT once all assignments were unsuccessful), has a worse-case complexity of `O(2^n)`. That means there are, for instance, more than 1 billion assignments to test
for a formula with 30 variables, and about `1.27 * 10^30` for a formula with 100 variables: the time to solve a very simple problem is unacceptable.

But all hope is not lost.

The first good news is that exponential does not necessarily mean `2^n`. It means `k^n`, with `k > 1`. Solving the SAT problem is a whole research field, and algorithms are more and more efficient,
trying to bring `k` as close to 1 as possible. Because, for instance, although an algorithm with a worst-case complexity of `O(1.01^n)` is still exponential, there are only about 400 million operations to
solve a problem with 2000 variables.

The second good news is that practical, real-world, industrial problems are generally structured in such a way that worst-case complexity is just a theoretical limit. Those problems are usually easier to
solve than theoretical ones.

The third good news is that, according to the Cook-Levin theorem, any NP-complete problem can be translated in a SAT problem, meaning all the research effort in emproving the efficiency of SAT solvers
also profits all other NP-complete problems. In other words, when you are facing an NP-complete problem, you can either try to develop a specific algorithm to solve it,
use a dedicated piece of software (if such a software exists) or try to translate it to SAT and use a SAT solver. 

## What is the input format of a SAT solver? How do I represent my problem?

### Conjunctive Normal Form

To be solved by a SAT solver, a problem must be described in **CNF**, or Conjunctive Normal Form. The CNF is a conjunction of disjunctions.

In other words, a CNF formula is a **set of clauses**, all of which must be true:

    F = C1 and C2 and C3 and ... and Cn

A **clause** is a set of literals, of which one (at least) must be true:

    Ci = L1 or L2 or ... or Lm

A literal is a variable, potentially negated.

Any boolean formula can be translated to an equivalent CNF formula in polynomial time, although the process can sometimes be tricky.
For instance, the formula

    (a and (b or not(c))) or (not(b) and c)

is equivalent to the following CNF formula

    (x1 or x2) and (not(x1) or a) and (not(x1) or b or not(c)) and (not(x2) or not(b)) and (not(x2) or c)

Note that a few more variables had to be introduced.


### DIMACS format

the standard way of describing a CNF formula is through the DIMACS syntax. A DIMACS file consists in a one-line prolog, and the description of each clause on each following lines.
The prolog starts with `p cnf ` and is followed by the number of variables and the number of clauses:

    p cnf 5 5

Then, on each line, a clause is described. A clause is described as a list of integer values separated by spaces, and ending with th value 0. That means each variable must be translated as an integer value.
Negation is represented by the minus sign. For instance, let's say `a` is 1, `b` is 2, `c` is 3, `x1` is 4, `x2` is 5. The clause `not(x2) or c` will be translated as `-5 3 0`. The above formula will then
be written as:

    p cnf 5 5
    4 5 0
    -4 1 0
    -4 2 -3 0
    -5 -2 0
    -5 3 0

The SAT solver, when given an input, will then return either the string **UNSATISFIABLE** or the string **SATISFIABLE** with an assignment that makes the formula true.
For instance, when fed the formula above, Gophersat will answer:

    SATISFIABLE
    -1 -2 3 -4 5

It means that, when both `c` and `x2` are true, and all other variables are false, the whole formula becomes true. Since `x1` and `x2` were dummy variables we introduced to make our original formula a CNF,
we can say that our original formula is true when `a=false, b=false, c=true`. Remember this is not necessarily the only model for this formula, but finding all models
(or getting the total number of models) would take a longer time and is not always useful.

## Translating a general boolean formula to a CNF formula

Formulas need to be in CNF to be used by a SAT solver. However, problems are usually not formulated this way. Translating general boolean formulas to CNF is thus required. This process can be automated,
but for  the moment Gophersat doesn't include such a translator. Here are a few guidelines to translate general formulas to CNF.

### Negation

In a CNF formula, negation only applies to literals, not to subformulas. Using DeMorgan's laws, such a transformation is trivial: `not(F1 or F2)` (where `F1` and `F2` are subformulas)
becomes `not(F1) and not(F2)` and `not(F1 and F2)` becomes `not(F1) or not(F2)`. Such transformations are then applied recursively, until only literals are negated.

### Implication (if... then)

Implication is easily translated in CNF: `a -> b` is equivalent to `not(a) or b`.

### Equivalence (equality)

The formula `a <-> b` is equivalent to `(a -> b) and (b -> a)`, and can thus be translated as the set of clauses `(not(a) or b) and (not(b) or a)`.

### Exclusive or

The formula `a xor b` is equivalent to `a <-> not(b)`, so it can be written `(a or b) and (not(a) or not(b))`.

### From disjunction of conjunctions to conjunctions of disjunctions

In a CNF formula, literals are linked by `or` operators to form clauses, and clauses are linked by `and` operators.
How do you deal with the opposite, i.e a subformula like `(F1 and F2) or (F3 and F4)` where `F1`, `F2`, `F3` and `F4` are subformulas?
There are several ways, but the most efficient one requires the introduction of new variables: let's rewrite this formula as

    (x1 <-> (F1 and F2)) and (x2 <-> (F3 and F4)) and (x1 or x2)

According to the rules above, `x1 <-> (F1 and F2)` can be rewritten as

    (not(x1) or F1) and (not(x1) or F2) and (not(F1) or not(F2) or x1)

So, the original formula can be written as the following set of clauses:

    not(x1) or F1
    not(x1) or F2
    x1 or not(F1) or not(F2)
    not(x2) or F3
    not(x2) or F4
    x2 or not(F3) or not(F4)
    x1 or x2

By applying this rule (and all rules above) recursively, you can translate any boolean formula in a CNF equivalent.

### Uniqueness

Very often, one wants to specify that exactly one boolean variable among several must be true, the others being false.
For instance, you may want to state that a device is in one of four states : s1, s2, s3 or s4.
It can be written as this set of clauses :

    s1 or s2 or s3 or s4
    not(s1) or not(s2)
    not(s1) or not(s3)
    not(s1) or not(s4)
    not(s2) or not(s3)
    not(s2) or not(s4)
    not(s3) or not(s4)

The first clause states "at least one state is true". Then, for each pair of states, one clause states that at least one of these two states is false.
Note that it can lead to the generation of a lot of clauses: there are `n*(n-1)/2` pairs of variables. So, when `n` is big, a huge amount of clauses can be generated.
There are other, more efficient encodings, but they are a little more complicated and won't be explained in this document.

## Examples

In this section, you will find the translation of a few problems translated as a boolean formula, then as a CNF boolean formula, then as a DIMACS file format.

## Sudoku

Let's say we want to solve a sudoku problem. Such a problem has a set of constraints:

- each spot has exactly one number from 1 to 9,
- on each line, each number appears exctaly once,
- on each column, each number appears exactly once,
- in each 3x3 box, each number appears exactly once,
- a few spots are already assigned numbers.

We must translate this problem in a formalism using only boolean variables, i.e assertions that are either true or false. In this problem, the variables we will use will be:

    line-1-col-1-is-a-1
    line-1-col-1-is-a-2
    ...
    line-1-col-1-is-a-9
    line-1-col-2-is-a-1
    ...

and so on. So, we will have `9*9*9=729` variables to represent this problem, each one indicating `line-x-col-y-is-a-z`, or, as a shortcut, `x;y=z`, so we will have `1;1=1`, `1;1=2` and so on.

Now, we must represent all our contraints.

### Each spot has exactly one number

For each spot, we must indicate that the associated value is unique:

    unique(1;1=1, 1;1=2, 1;1=3, 1;1=4, 1;1=5, 1;1=6, 1;1=7, 1;1=8, 1;1=9) and
    unique(1;2=1, 1;2=2, 1;2=3, 1;2=4, 1;2=5, 1;2=6, 1;2=7, 1;2=8, 1;2=9) and
    ... and
    unique(9;9=1, 9;9=2, 9;9=3, 9;9=4, 9;9=5, 9;9=6, 9;9=7, 9;9=8, 9;9=9)

This can be translated as a CNF set of clauses (see transformation rules above):

    1;1=1 or 1;1=2 or 1;1=3 or 1;1=4 or 1;1=5 or 1;1=6 or 1;1=7 or 1;1=8 or 1;1=9
    not(1;1=1) or not(1;1=2)
    not(1;1=1) or not(1;1=3)
    not(1;1=1) or not(1;1=4)
    ...
    not(1;1=8) or not(1;1=9)

This will lead to the generation of `1 + (9*8)/2 = 37` clauses for each spot, i.e `9*9*37 = 2997` clauses for all spots.

### Each line, column, box contains each number

Since there are 9 spots in a line and each number from 1 to 9 must appear once, stating that each number must appear at least once is enough. So, we will write:

    (1;1=1 or 1;2=1 or 1;3=1 or 1;4=1 or 1;5=1 or 1;6=1 or 1;7=1 or 1;8=1 or 1;9=1) and
    (2;1=1 or 2;2=1 or 2;3=1 or 2;4=1 or 2;5=1 or 2;6=1 or 2;7=1 or 2;8=1 or 2;9=1) and
    ... and
    (9;1=1 or 9;2=1 or 9;3=1 or 9;4=1 or 9;5=1 or 9;6=1 or 9;7=1 or 9;8=1 or 9;9=1) and
    (1;1=2 or 1;2=2 or 1;3=2 or 1;4=2 or 1;5=2 or 1;6=2 or 1;7=2 or 1;8=2 or 1;9=2) and
    ...

This will lead to the generation of `9*9=81` clauses.

The same reasoning can be applied for columns and boxes, leading to the generation of 81 clauses each time.

### Already assigned numbers

Now this is easy, we just have to state that a few variables are known to be true. So, for the following problem:

![Sudoky problem example](https://upload.wikimedia.org/wikipedia/commons/e/e0/Sudoku_Puzzle_by_L2G-20050714_standardized_layout.svg)

the following clauses will be added:

    1;1=5
    1;2=3
    1;5=7
    2;1=6
    2;4=1
    2;5=9
    2;6=5
    ....

This will lead to the generation of one clause per already assigned number, i.e 30 clauses for the example problem above.

### Putting it all together in a DIMACS file

To generate the corresponding DIMACS file, we must indicate the total number of variables and clauses in the prolog line.
There are `729` variables and `2997+81+81+81+30=3270` clauses, so the prolog will be

    p cnf 729 3270

Now, each variable must be associated with a numeric value, from 1 to 729. Let's say that `x;y=z` is translated as `(x-1) * 81 + (y-1) * 9 + z`. So, `1;1=1` will be `1` and `9;9=9` will be 729.
The [complete DIMACS file](sudoku.cnf) is available.

When fed that problem, gophersat will answer

    SATISFIABLE
    -1 -2 -3 -4 5 -6 -7 -8 -9 -10 -11 12 -13 -14 -15 -16 -17 -18 -19 -20 -21 22 -23 -24 -25 -26 -27 -28 -29 -30 -31 -32 33 -34 -35 -36 -37 -38 -39 -40 -41 -42 43 -44 -45 -46 -47 -48 -49 -50 -51 -52 53 -54 -55 -56 -57 -58 -59 -60 -61 -62 63 64 -65 -66 -67 -68 -69 -70 -71 -72 -73 74 -75 -76 -77 -78 -79 -80 -81 -82 -83 -84 -85 -86 87 -88 -89 -90 -91 -92 -93 -94 -95 -96 97 -98 -99 -100 101 -102 -103 -104 -105 -106 -107 -108 109 -110 -111 -112 -113 -114 -115 -116 -117 -118 -119 -120 -121 -122 -123 -124 -125 126 -127 -128 -129 -130 131 -132 -133 -134 -135 -136 -137 138 -139 -140 -141 -142 -143 -144 -145 -146 -147 148 -149 -150 -151 -152 -153 -154 -155 -156 -157 -158 -159 -160 161 -162 163 -164 -165 -166 -167 -168 -169 -170 -171 -172 -173 -174 -175 -176 -177 -178 -179 180 -181 -182 -183 -184 -185 -186 -187 188 -189 -190 -191 192 -193 -194 -195 -196 -197 -198 -199 -200 -201 202 -203 -204 -205 -206 -207 -208 209 -210 -211 -212 -213 -214 -215 -216 -217 -218 -219 -220 221 -222 -223 -224 -225 -226 -227 -228 -229 -230 231 -232 -233 -234 -235 -236 -237 -238 -239 -240 241 -242 -243 -244 -245 -246 -247 -248 -249 -250 251 -252 -253 -254 -255 -256 257 -258 -259 -260 -261 -262 -263 -264 -265 -266 -267 -268 -269 270 -271 -272 -273 -274 -275 -276 277 -278 -279 -280 -281 -282 -283 -284 285 -286 -287 -288 289 -290 -291 -292 -293 -294 -295 -296 -297 -298 -299 -300 301 -302 -303 -304 -305 -306 -307 308 -309 -310 -311 -312 -313 -314 -315 -316 -317 318 -319 -320 -321 -322 -323 -324 -325 -326 -327 328 -329 -330 -331 -332 -333 -334 335 -336 -337 -338 -339 -340 -341 -342 -343 -344 -345 -346 -347 348 -349 -350 -351 -352 -353 -354 -355 -356 -357 -358 359 -360 -361 -362 -363 -364 365 -366 -367 -368 -369 -370 -371 372 -373 -374 -375 -376 -377 -378 -379 -380 -381 -382 -383 -384 385 -386 -387 -388 -389 -390 -391 -392 -393 -394 -395 396 397 -398 -399 -400 -401 -402 -403 -404 -405 -406 -407 -408 -409 -410 -411 412 -413 -414 415 -416 -417 -418 -419 -420 -421 -422 -423 -424 -425 426 -427 -428 -429 -430 -431 -432 -433 -434 -435 -436 -437 -438 -439 -440 441 -442 443 -444 -445 -446 -447 -448 -449 -450 -451 -452 -453 454 -455 -456 -457 -458 -459 -460 -461 -462 -463 -464 -465 -466 467 -468 -469 -470 -471 -472 473 -474 -475 -476 -477 -478 -479 -480 -481 -482 483 -484 -485 -486 -487 -488 -489 -490 -491 -492 -493 -494 495 -496 -497 -498 -499 -500 501 -502 -503 -504 505 -506 -507 -508 -509 -510 -511 -512 -513 -514 -515 -516 -517 518 -519 -520 -521 -522 -523 -524 525 -526 -527 -528 -529 -530 -531 -532 -533 -534 -535 -536 -537 538 -539 -540 -541 542 -543 -544 -545 -546 -547 -548 -549 -550 -551 -552 -553 -554 -555 -556 557 -558 -559 -560 -561 562 -563 -564 -565 -566 -567 -568 569 -570 -571 -572 -573 -574 -575 -576 -577 -578 -579 -580 -581 -582 -583 584 -585 -586 -587 -588 -589 -590 -591 592 -593 -594 -595 -596 -597 598 -599 -600 -601 -602 -603 604 -605 -606 -607 -608 -609 -610 -611 -612 -613 -614 -615 -616 -617 -618 -619 -620 621 -622 -623 -624 -625 -626 627 -628 -629 -630 -631 -632 633 -634 -635 -636 -637 -638 -639 -640 -641 -642 -643 644 -645 -646 -647 -648 -649 -650 651 -652 -653 -654 -655 -656 -657 -658 -659 -660 661 -662 -663 -664 -665 -666 -667 -668 -669 -670 671 -672 -673 -674 -675 -676 677 -678 -679 -680 -681 -682 -683 -684 -685 -686 -687 -688 -689 -690 -691 692 -693 -694 -695 -696 -697 -698 699 -700 -701 -702 703 -704 -705 -706 -707 -708 -709 -710 -711 -712 -713 -714 -715 -716 -717 718 -719 -720 -721 -722 -723 -724 -725 -726 -727 -728 729

This answer must now be parsed: each positive value indicates the corresponding spot contains the corresponding value. For example, the value 661 is positive. And the variable 661 means
`9;2=4`; so, the value at line 9, column 2 is `4`. 

Representing a sudoku puzzle can seem tedious, but once it is done, the process can be easily automated and used for any other puzzle.

## Automated planning

Let's say we have a small factory that should be running 24/7. Now, we have a few constraints:

- an employee should only work 8 hours at most in a given day,
- an employee should at most work 5 days per week,
- there must be at least one employee at any given moment,
- we only want to hire 4 employees.

Now, to simplify things, we consider there are 3 working time slots per day: night(12AM - 7AM), morning(7AM - 2PM) and afternoon(2PM - 12AM). How could we generate a timetable for our employees?

### Variables

Let's deal with variables first. There are `7*3=21` time slots and 4 employees, so we will have at least `21*4=84` variables, each variable meaning `employee-x-works-on-day-y-timeslot-z`,
or, as a shortcut `x=y;z`. Now, that could be enough, but to simplify the writing of our formula, let's add a few variables, to indicate `employee-x-works-on-day-y`, with no time slot indicated,
or, as a shortcut `x=dy`. There will be `4*7=28` such variables, i.e `112` total variables.

### Employees work only 8 hours a day

This is akin to a unique constraint: given a set of variable, exactly one must be true. But, here, this is not an "exactly one" constraint, this is an "at most one" constraint: an employee does not
necessarily work on a given day, but if he does, he only works during one time slot. So, for employee 1 on day 1, we will have the following set of clauses:

    not(1=1;1) or not(1=1;2)
    not(1=1;1) or not(1=1;3)
    not(1=1;2) or not(1=1;3)

And we will do the same for the 7 days, for all 4 employees, leading to the generation of `3*7*4=84` clauses.

### At least one employee in each timeslot

That's an easy one: for each time slot, we indicate that one of the 4 variables representing each employee for that time slot is true. So, for monday night (day 1, time slot 1), this will be:

    1=1;1 or 2=1;1 or 3=1;1 or 4=1;1

Whether we have one, two, three or all four employees working, this clause will be true. We do the same for all 21 time slots, leading to the generation of 21 clauses.


### Employees work at most 5 days per week

This one is a little trickier. No, we want our workers to have at most five working days, meaning that, for each set of 6 days, at least one must be a day off. For employee 1, this will lead to the generation
of these clauses:
    
    not(1=d1) or not(1=d2) or not(1=d3) or not(1=d4) or not(1=d5) or not(1=d6)
    not(1=d1) or not(1=d2) or not(1=d3) or not(1=d4) or not(1=d5) or not(1=d7)
    not(1=d1) or not(1=d2) or not(1=d3) or not(1=d4) or not(1=d6) or not(1=d7)
    not(1=d1) or not(1=d2) or not(1=d3) or not(1=d5) or not(1=d6) or not(1=d7)
    not(1=d1) or not(1=d2) or not(1=d4) or not(1=d5) or not(1=d6) or not(1=d7)
    not(1=d1) or not(1=d3) or not(1=d4) or not(1=d5) or not(1=d6) or not(1=d7)
    not(1=d2) or not(1=d3) or not(1=d4) or not(1=d5) or not(1=d6) or not(1=d7)

So, for all 4 employees, we will have `4*7=28` more clauses.


### Linking `x=y;z` variables with `x=dy` ones

We must link both types of variables: if an employee works on day 1, that means he either works on day1-slot1, on day1-slot2 or on day1-slot3. For employee 1, that means:

    1=d1 <-> (1=1;1 or 1=1;2 or 1=1;3)

In CNF:

    not(1=1;1) or 1=d1
    not(1=1;2) or 1=d1
    not(1=1;3) or 1=d1
    not(1=d1) or 1=1;1 or 1=1;2 or 1=1;3

For each day and each employee, we have to generate `4*7*4=112` clauses.

### Translation to DIMACS

We need to associate each variable with an integer identifiant. Let's say `x=y;z` be associated with `(x-1)*21 + (y-1)*3 + z`. That leads to the generation of 84 variables. We also need a different identifier
for each variable of the kind `x=dy`. We will associate them with the value `84 + (x-1)*7 + y`.

The [DIMACS file](timetable4.cnf) is available for download. When we feed it to gophersat, it answers

    UNSATISFIABLE

Oops! Looks like there is no way to solve our problem. Actually, we could have found that without a solver: there are 4 employees, and each employee works at most during 5 timeslots, meaning only `5*4=20`
time slots with an employee at most. But there are `7*3=21` time slots to fill. We need to relax our constraints. Let's say we decide to hire a fifth employee. We rewrite our problem, as described in [DIMACS file timetable5.cnf](timetable5.cnf) and ask gophersat his advice. The solver answers:

    SATISFIABLE
    -1 2 -3 -4 -5 -6 -7 -8 -9 -10 -11 12 13 -14 -15 16 -17 -18 -19 -20 21 -22 -23 -24 25 -26 -27 28 -29 -30 -31 32 -33 -34 -35 -36 -37 38 -39 40 -41 -42 43 -44 -45 -46 -47 48 -49 50 -51 -52 -53 -54 -55 56 -57 -58 -59 -60 -61 62 -63 -64 -65 66 -67 68 -69 -70 -71 72 73 -74 -75 -76 -77 78 -79 -80 -81 -82 -83 -84 -85 -86 -87 -88 -89 -90 -91 -92 -93 -94 -95 -96 -97 -98 -99 -100 -101 102 -103 -104 -105 106 -107 -108 109 110 111 112 -113 114 115 116 -117 118 119 120 121 122 -123 124 -125 126 127 128 129 130 131 -132 -133 -134 -135 -136 -137 -138 139 -140 

Great! Now, we have to translate this answer, by checking what each positive value means. Here, the solver gave us the following timetable:

    1=1;2
    1=4;3
    1=5;1
    1=6;1
    1=7;3
    2=2;1
    2=3;1
    2=4;2
    2=6;2
    2=7;1
    3=1;1
    3=2;3
    3=3;2
    3=5;2
    3=7;2
    4=1;3
    4=2;2
    4=3;3
    4=4;1
    4=5;3
    5=6;3

Everything is okay now. All employees from 1 to 4 work only five days a week, and part-time employee 5 fills the last, unassigned slot.
