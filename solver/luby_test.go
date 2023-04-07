package solver

import "testing"

func TestLuby(t *testing.T) {
	vals := []uint{1, 1, 2, 1, 1, 2, 4, 1, 1, 2, 1, 1, 2, 4, 8, 1, 1, 2, 1, 1, 2, 4}
	for i, val := range vals {
		if luby(uint(i)+1) != val {
			t.Errorf("invalid luby term luby(%d): expected %d, got %d", i+1, val, luby(uint(i)+1))
		}
	}
}
