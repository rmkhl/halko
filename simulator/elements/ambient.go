package elements

// Ambient is a fixed temperature source. It stands in for anything the
// simulation does not model but still has to report a reading for, such as
// the sensor board's cold junctions.
type Ambient struct {
	temperature float32
}

func NewAmbient(temp float32) *Ambient {
	return &Ambient{temperature: temp}
}

func (a *Ambient) Temperature() float32 {
	return a.temperature
}
