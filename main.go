package main

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/crillab/gophersat/solver"
)

func main() {
	debug.SetGCPercent(300)
	if len(os.Args) > 2 {
		fmt.Fprintf(os.Stderr, "Syntax : %s [file.cnf]\n", os.Args[0])
		os.Exit(1)
	}
	f := os.Stdin
	if len(os.Args) == 2 {
		var err error
		f, err = os.Open(os.Args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not open %s: %v\n", os.Args[1], err.Error())
			os.Exit(1)
		}
		defer f.Close()
	}
	fmt.Printf("c ======================================================================================\n")
	fmt.Printf("c | Parsing problem...                                                                 |\n")
	pb, err := solver.ParseCNF(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}
	fmt.Printf("c | Number of clauses   : %9d                                                    |\n", len(pb.Clauses))
	fmt.Printf("c | Number of variables : %9d                                                    |\n", pb.NbVars)
	s := solver.New(pb)
	s.Verbose = true
	s.Solve()
	fmt.Printf("c nb conflicts: %d\nc nb restarts: %d\nc nb decisions: %d\n", s.Stats.NbConflicts, s.Stats.NbRestarts, s.Stats.NbDecisions)
	fmt.Printf("c nb unit learned: %d\nc nb binary learned: %d\nc nb learned: %d\n", s.Stats.NbUnitLearned, s.Stats.NbBinaryLearned, s.Stats.NbLearned)
	fmt.Printf("c nb clauses deleted: %d\n", s.Stats.NbDeleted)
	s.OutputModel()
}
