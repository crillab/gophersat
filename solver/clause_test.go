package solver

import "testing"

func TestSimplifyPB(t *testing.T) {
	l1 := IntToLit(1)
	l2 := IntToLit(2)
	l3 := IntToLit(3)
	l4 := IntToLit(4)
	lits := []Lit{l1, l2, l3, l4}
	weights := []int{5, 3, 1, 1}
	c := NewPBClause(lits, weights, 7) // 5a + 3b + c + d >= 7
	units, c2, ok := c.SimplifyPB()
	if !ok {
		t.Errorf("unexpected unsat while simplifying %s", c.PBString())
	}
	if len(units) != 1 || units[0] != l1 {
		t.Errorf("unexpected unit lits when simplifying %s: got %v", c.PBString(), units)
	}
	if c2 == nil {
		t.Errorf("unexpected satisfied clause when simplifying %s", c.PBString())
	}
	t.Logf("new clause is %s", c2.PBString())
	if c2.Len() != 3 || c2.Cardinality() != 2 {
		t.Errorf("unexpected simplified constraint when simplifying %s: got %s", c.PBString(), c2.PBString())
	}
	c = NewPBClause(lits, weights, 8) // 5a + 3b + c + d >= 8
	units, c2, ok = c.SimplifyPB()
	if !ok {
		t.Errorf("unexpected unsat while simplifying %s", c.PBString())
	}
	if len(units) != 2 || units[0] != l1 || units[1] != l2 {
		t.Errorf("unexpected unit lits when simplifying %s: got %v", c.PBString(), units)
	}
	if c2 != nil {
		t.Errorf("expected satisfied clause when simplifying %s, got %v", c.PBString(), c2.PBString())
	}
	c = NewPBClause(lits, weights, 11) // 5a + 3b + c + d >= 22
	units, c2, ok = c.SimplifyPB()
	if ok {
		t.Errorf("unexpected sat while simplifying %s: got units %v and new clause %v", c.PBString(), units, c2)
	}
}
