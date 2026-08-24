package types

// RunNote is an operator's observation, recorded while a run is in progress.
// Everything but Text is stamped by the control unit at the moment the note is
// taken, so a note read months later carries the state it was written about
// instead of having to be matched against the chart by hand.
//
// Notes are only ever created during a live run, so every field is always
// populated; there is no "context unavailable" case to represent.
type RunNote struct {
	Time         int64             `json:"time"`         // unix seconds
	Step         string            `json:"step"`         // step running at the time
	Temperatures TemperatureStatus `json:"temperatures"` // kiln and material
	Text         string            `json:"text"`
}
