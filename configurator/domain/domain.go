package domain

type Name string

const EmptyName Name = ""

func (i Name) IsValid() bool {
	return i != EmptyName
}

type HasName struct {
	Name Name `json:"name"`
}
