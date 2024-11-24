package engine

type PowerManager interface {
	TurnOn(cycle *Cycle)
	TurnOff()
	SwitchTo(cycle *Cycle)
}
