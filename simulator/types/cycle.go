package types

type (
	Cycle struct {
		percentage int
		ticks      [10]bool
	}
)

func NewCycle(percentage int) *Cycle {
	sc := Cycle{percentage: percentage}

	for i := 0; i < 10; i++ {
		sc.ticks[i] = i*10 < percentage
	}
	return &sc
}
