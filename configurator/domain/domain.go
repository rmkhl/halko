package domain

type ID string

const EmptyID ID = ""

func (i ID) IsValid() bool {
	return i != EmptyID
}

type HasID struct {
	ID ID `json:"id"`
}
