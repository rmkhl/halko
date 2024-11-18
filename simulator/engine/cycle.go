package engine

type (
	Cycle struct {
		percentage uint8
		ticks      [10]bool
	}
)

func NewCycle(percentage uint8) *Cycle {
	sc := Cycle{percentage: percentage}

	for i := 0; i < 10; i++ {
		sc.ticks[i] = uint8(i*10) < percentage
	}
	return &sc
}
