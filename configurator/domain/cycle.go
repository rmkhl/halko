package domain

type Cycle struct {
	HasID
	Name   string `json:"name"`
	States []bool `json:"states"`
}
