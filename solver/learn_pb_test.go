package solver

import (
	"testing"
)

func TestPbSet(t *testing.T) {
	prob := ParseSlice([][]int{{1, 2, 3, 4, 5}})
	s := New(prob)
	// constr = 5 x1 +3 ~x2 +2 x4 +1 x5 >= 6
	constr := PBConstr{Lits: []int{1, -2, 4, 5}, Weights: []int{5, 3, 2, 1}, AtLeast: 6}
	buffer := make([]int, s.nbVars)
	pb := s.pbSet(constr.Clause(), buffer)
	if pb.card != 6 {
		t.Errorf("invalid cardinality for pbSet, expected 6, got %d", pb.card)
	}
	if len(pb.weights) != 5 {
		t.Errorf("invalid size for weights in pbSet: expected 5, got %d", len(pb.weights))
	}
	if pb.weights[0] != 5 || pb.weights[1] != -3 || pb.weights[2] != 0 || pb.weights[3] != 2 || pb.weights[4] != 1 {
		t.Errorf("invalid vector in pbSet, expected %v, got %v", []int{5, -3, 0, 2, 1}, pb.weights)
	}
	constr2 := PBConstr{Lits: []int{2, -1, 4, 5, 3}, Weights: []int{6, 2, 2, 2, 1}, AtLeast: 7}
	buffer2 := make([]int, s.nbVars)
	pb2 := s.pbSet(constr2.Clause(), buffer2)
	if pb2.card != 7 {
		t.Errorf("invalid cardinality for pbSet, expected 7, got %d", pb2.card)
	}
	if len(pb.weights) != 5 {
		t.Errorf("invalid size for weights in pbSet: expected 5, got %d", len(pb2.weights))
	}
	if pb2.weights[0] != -2 || pb2.weights[1] != 6 || pb2.weights[2] != 1 || pb2.weights[3] != 2 || pb2.weights[4] != 2 {
		t.Errorf("invalid vector in pbSet, expected %v, got %v", []int{-2, 6, 1, 2, 2}, pb2.weights)
	}
}

func TestFalsifies(t *testing.T) {
	prob := ParseSlice([][]int{{1, 2, 3, 4, 5}, {-1, 2}})
	s := New(prob)
	// constr = 5 x1 +3 ~x2 +2 x4 +1 x5 >= 6
	constr := PBConstr{Lits: []int{1, -2, 4, 5}, Weights: []int{5, 3, 2, 1}, AtLeast: 6}
	buffer := make([]int, s.nbVars)
	pb := s.pbSet(constr.Clause(), buffer)
	if !pb.falsifies(IntToLit(-1)) {
		t.Errorf("pbSet should falsify -1 but does not")
	}
	if pb.falsifies(IntToLit(1)) {
		t.Errorf("pbSet shouldn't falsify 1 but it does")
	}
	if !pb.falsifies(IntToLit(2)) {
		t.Errorf("pbSet should falsify 2 but does not")
	}
	if pb.falsifies(IntToLit(-2)) {
		t.Errorf("pbSet shouldn't falsify -2 but it does")
	}
	if pb.falsifies(IntToLit(-3)) {
		t.Errorf("pbSet shouldn't falsify -3 but it does")
	}
	if pb.falsifies(IntToLit(3)) {
		t.Errorf("pbSet shouldn't falsify 3 but it does")
	}
}

func TestSlack(t *testing.T) {
	prob := ParseSlice([][]int{{1, 2, 3, 4, 5}})
	s := New(prob)
	// constr = 5 x1 +3 ~x2 +2 x4 +1 x5 >= 6
	constr := PBConstr{Lits: []int{1, -2, 4, 5}, Weights: []int{5, 3, 2, 1}, AtLeast: 6}
	buffer := make([]int, s.nbVars)
	pb1 := s.pbSet(constr.Clause(), buffer)
	if s.slack(pb1, 1) != 5 {
		t.Errorf("invalid slack, expected 5, got %d", s.slack(pb1, 1))
	}
}

func TestPbSet_clause(t *testing.T) {
	prob := ParseSlice([][]int{{1, 2, 3, 4, 5}})
	s := New(prob)
	// constr = 5 x1 +3 ~x2 +2 x4 +1 x5 >= 6
	constr := PBConstr{Lits: []int{1, -2, 4, 5}, Weights: []int{5, 3, 2, 1}, AtLeast: 6}
	buffer := make([]int, s.nbVars)
	pb1 := s.pbSet(constr.Clause(), buffer)
	c := pb1.clause()
	str := c.PBString()
	if str != "5 x1 +3 ~x2 +2 x4 +1 x5 >= 6 ;" {
		t.Errorf("error while generating PB constraint from pbSet: got %q", c.PBString())
	}
}
