package solver

import (
	"testing"
)

func TestExactly1(t *testing.T) {
	for i := 1; i < 4; i++ {
		vars := make([]int, i)
		for j := 0; j < i; j++ {
			vars[j] = j + 1
		}
		cardInstance := Exactly1(vars...)
		pb := ParseCardConstrs(cardInstance)
		s := New(pb)
		modelChan := make(chan []bool)
		stopChan := make(chan struct{})
		go s.Enumerate(modelChan, stopChan)
		models := [][]bool{}
		for model := range modelChan {
			models = append(models, model)
		}
		if len(models) != i {
			t.Errorf("Expected %d models for Exactly1(1,...,%d). Found %d. %+v",
				i, i, len(models), models)
		}
		for _, model := range models {
			set := false
			for _, bit := range model {
				if bit {
					if set {
						t.Errorf("Found unexpected model %v for Exactly1(1,...,%d).", model, i)
					}
					set = true
				}
			}
			if !set {
				t.Errorf("Found unexpected model %v for Exactly1(1,...,%d).", model, i)
			}
		}
	}
}
