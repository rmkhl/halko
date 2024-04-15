package types

type (
	Cycle struct {
		name       string
		percentage int
		ticks      [10]bool
	}
)

func NewCycle(name string, ticks [10]bool) *Cycle {
	sc := Cycle{name: name}

	sc.percentage = 0
	for i := 0; i < 10; i++ {
		sc.ticks[i] = ticks[i]
		if ticks[i] {
			sc.percentage += 10
		}
	}
	return &sc
}
