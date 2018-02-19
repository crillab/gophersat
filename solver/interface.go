package solver

// A ModelMap associates variable identifiers with a binding.
// Typically, identifiers will be integer values, but other identifiers can be used in higher-level solvers, like human-readable strings, for instance.
type ModelMap map[interface{}]bool

// A Result is a status, either Sat, Unsat or Indet.
// If the status is Sat, the Result also associates a ModelMap with an integer value.
// This value is typically used in optimization processes.
// If the weight is 0, that means all constraints could be solved.
// By definition, in decision problems, the cost will always be 0.
type Result struct {
	Status Status
	Model  ModelMap
	Weight int
}

// Interface is any type implementing a solver.
// The basic Solver defined in this package implements it.
// Any solver that uses the basic solver to solve more complex problems
// (MAXSAT, MUS extraction, etc.) can implement it, too.
type Interface interface {
	// Optimal solves or optimizes the problem and returns the best result.
	// If the results chan is non nil, it will write the associated model each time one is found.
	// It will stop as soon as a model of cost 0 is found, or the problem is not satisfiable anymore.
	// The last satisfying model, if any, will be returned with the Sat status.
	// If no model at all could be found, the Unsat status will be returned.
	// If the solver prematurely stopped, the Indet status will be returned.
	// If data is sent to stop, the method may stop prematurely.
	// In any case, results will be closed before the function returns.
	// NOTE: data sent on stop may be ignored by an implementation.
	Optimal(results chan Result, stop chan struct{}) Result
	// Enumerate returns the number of models for the problem.
	// If the models chan is non nil, it will write the associated model each time one is found.
	// If data is sent to stop, the method may stop prematurely.
	// In any case, models will be closed before the function returns.
	// NOTE: data sent on stop may be ignored by an implementation.
	Enumerate(models chan ModelMap, stop chan struct{}) int
}
