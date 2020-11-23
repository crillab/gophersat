package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/debug"
	"sort"
	"strings"

	"github.com/crillab/gophersat/bf"
	"github.com/crillab/gophersat/explain"
	"github.com/crillab/gophersat/maxsat"
	"github.com/crillab/gophersat/solver"
)

const helpString = "This is gophersat version 1.3, a SAT and Pseudo-Boolean solver by Fabien Delorme.\n"

func main() {
	// defer profile.Start().Stop()
	debug.SetGCPercent(300)
	var (
		verbose bool
		cert    bool
		mus     bool
		count   bool
		help    bool
	)
	flag.BoolVar(&verbose, "verbose", false, "sets verbose mode on")
	flag.BoolVar(&cert, "certified", false, "displays RUP certificate on stdout")
	flag.BoolVar(&mus, "mus", false, "extracts a MUS from an unsat problem")
	flag.BoolVar(&count, "count", false, "rather than solving the problem, counts the number of models it accepts")
	flag.BoolVar(&help, "help", false, "displays help")
	flag.Parse()
	if !help && len(flag.Args()) != 1 {
		fmt.Printf(helpString)
		fmt.Fprintf(os.Stderr, "Syntax : %s [options] (file.cnf|file.wcnf|file.bf|file.opb)\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}
	if help {
		fmt.Printf(helpString)
		fmt.Printf("Syntax : %s [options] (file.cnf|file.wcnf|file.bf|file.opb)\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(0)
	}
	path := flag.Args()[0]
	if mus {
		extractMUS(path)
	} else {
		fmt.Printf("c solving %s\n", path)
		if strings.HasSuffix(path, ".bf") {
			if err := parseAndSolveBF(path); err != nil {
				fmt.Fprintf(os.Stderr, "could not parse formula: %v\n", err)
				os.Exit(1)
			}
		} else if strings.HasSuffix(path, ".wcnf") {
			if err := parseAndSolveWCNF(path, verbose); err != nil {
				fmt.Fprintf(os.Stderr, "could not parse MAXSAT file %q: %v", path, err)
				os.Exit(1)
			}
		} else {
			if pb, printFn, err := parse(flag.Args()[0]); err != nil {
				fmt.Fprintf(os.Stderr, "could not parse problem: %v\n", err)
				os.Exit(1)
			} else if count {
				countModels(pb, verbose)
			} else {
				solve(pb, verbose, cert, printFn)
			}
		}
	}
}

func extractMUS(path string) {
	f, err := os.Open(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not parse problem: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()
	pb, err := explain.ParseCNF(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not parse problem: %v\n", err)
		os.Exit(1)
	}
	pb2, err := pb.MUS()
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not extract subset: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(pb2.CNF())
}

func countModels(pb *solver.Problem, verbose bool) {
	s := solver.New(pb)
	if verbose {
		fmt.Printf("c ======================================================================================\n")
		fmt.Printf("c | Number of non-unit clauses : %9d                                             |\n", len(pb.Clauses))
		fmt.Printf("c | Number of variables        : %9d                                             |\n", pb.NbVars)
		s.Verbose = true
	}
	models := make(chan []bool)
	go s.Enumerate(models, nil)
	nb := 0
	for range models {
		nb++
		if verbose {
			fmt.Printf("c %d models found\n", nb)
		}
	}
	fmt.Println(nb)
}

func solve(pb *solver.Problem, verbose, cert bool, printFn func(chan solver.Result)) {
	s := solver.New(pb)
	if verbose {
		fmt.Printf("c ======================================================================================\n")
		fmt.Printf("c | Number of non-unit clauses : %9d                                             |\n", len(pb.Clauses))
		fmt.Printf("c | Number of variables        : %9d                                             |\n", pb.NbVars)
		s.Verbose = true
	}
	s.Certified = cert
	results := make(chan solver.Result)
	go s.Optimal(results, nil)
	printFn(results)
	if verbose {
		fmt.Printf("c nb conflicts: %d\nc nb restarts: %d\nc nb decisions: %d\n", s.Stats.NbConflicts, s.Stats.NbRestarts, s.Stats.NbDecisions)
		fmt.Printf("c nb unit learned: %d\nc nb binary learned: %d\nc nb learned: %d\n", s.Stats.NbUnitLearned, s.Stats.NbBinaryLearned, s.Stats.NbLearned)
		fmt.Printf("c nb learned clauses deleted: %d\n", s.Stats.NbDeleted)
	}
}

func parseAndSolveWCNF(path string, verbose bool) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("could not open %q: %v", path, err)
	}
	defer f.Close()
	s, err := maxsat.ParseWCNF(f)
	if err != nil {
		return fmt.Errorf("could not parse wcnf content: %v", err)
	}
	results := make(chan solver.Result)
	go s.Optimal(results, nil)
	printOptimizationResults(results)
	return nil
}

func parseAndSolveBF(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("could not open %q: %v", path, err)
	}
	defer f.Close()
	form, err := bf.Parse(f)
	if err != nil {
		return fmt.Errorf("could not parse formula in %q: %v", path, err)
	}
	solveBF(form)
	return nil
}

func parse(path string) (pb *solver.Problem, printFn func(chan solver.Result), err error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, fmt.Errorf("could not open %q: %v", path, err)
	}
	defer f.Close()
	if strings.HasSuffix(path, ".bf") {
		_, err := bf.Parse(f)
		if err != nil {
			return nil, nil, fmt.Errorf("could not parse %q: %v", path, err)
		}
		panic("not yet implemented")
	}
	if strings.HasSuffix(path, ".cnf") {
		pb, err := solver.ParseCNF(f)
		if err != nil {
			return nil, nil, fmt.Errorf("could not parse DIMACS file %q: %v", path, err)
		}
		return pb, printDecisionResults, nil
	}
	if strings.HasSuffix(path, ".opb") {
		pb, err := solver.ParseOPB(f)
		if err != nil {
			return nil, nil, fmt.Errorf("could not parse OPB file %q: %v", path, err)
		}
		return pb, printOptimizationResults, nil
	}
	return nil, nil, fmt.Errorf("invalid file format for %q", path)
}

func solveBF(f bf.Formula) {
	if model := bf.Solve(f); model == nil {
		fmt.Println("UNSATISFIABLE")
	} else {
		fmt.Println("SATISFIABLE")
		keys := make(sort.StringSlice, 0, len(model))
		for k := range model {
			keys = append(keys, k)
		}
		sort.Sort(keys)
		for _, k := range keys {
			fmt.Printf("%s: %t\n", k, model[k])
		}
	}
}

// prints the result to a SAT decision problem in the competition format.
func printDecisionResults(results chan solver.Result) {
	var res solver.Result
	for res = range results {
	}
	switch res.Status {
	case solver.Unsat:
		fmt.Println("s UNSATISFIABLE")
	case solver.Sat:
		fmt.Println("s SATISFIABLE")
		fmt.Printf("v ")
		for i := range res.Model {
			val := i + 1
			if !res.Model[i] {
				val = -i - 1
			}
			fmt.Printf("%d ", val)
		}
		fmt.Println("0")
	default:
		fmt.Println("s UNKNOWN")
	}
}

// prints the result to a PB optimization problem in the competition format.
func printOptimizationResults(results chan solver.Result) {
	var res solver.Result
	for res = range results {
		if res.Status == solver.Sat {
			fmt.Printf("o %d\n", res.Weight)
		}
	}
	switch res.Status {
	case solver.Unsat:
		fmt.Println("s UNSATISFIABLE")
	case solver.Sat:
		fmt.Println("s OPTIMUM FOUND")
		fmt.Printf("v ")
		for i := 0; i < len(res.Model); i++ {
			var val string
			if !res.Model[i] {
				val = fmt.Sprintf("-x%d", i+1)
			} else {
				val = fmt.Sprintf("x%d", i+1)
			}
			fmt.Printf("%s ", val)
		}
		fmt.Println()
	default:
		fmt.Println("s UNKNOWN")
	}
}
