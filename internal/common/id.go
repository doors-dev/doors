package common

type ID *id

type id struct {
	_ byte
}

func NewID() ID {
	return new(id)
}


