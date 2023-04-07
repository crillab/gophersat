package solver

const lubyConstant = 512

func luby(i uint) uint {
	for k := 1; k < 32; k++ {
		if i == (1<<k)-1 {
			return 1 << (k - 1)
		}
	}
	k := 1
	for {
		if (1<<(k-1)) <= i && i < (1<<k)-1 {
			return luby(i - (1 << (k - 1)) + 1)
		}
		k++
	}
}
