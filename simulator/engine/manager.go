package engine

type PowerManager interface {
	TurnOn(initialState bool)
	TurnOff()
	SwitchTo(upcoming bool)
	Info() (bool, bool)
}
