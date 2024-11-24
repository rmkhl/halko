package engine

type TemperatureSensor interface {
	Temperature() float32
}

type PowerSensor interface {
	IsOn() bool
	Name() string
	CurrentCycle() uint8
}
