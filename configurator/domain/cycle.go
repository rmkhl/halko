package domain

type Cycle struct {
	Name   string `json:"name"`
	States []bool `json:"states"`
}
