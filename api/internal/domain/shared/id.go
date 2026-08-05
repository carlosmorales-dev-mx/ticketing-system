package shared

import "github.com/google/uuid"

// ID es un Value Object que envuelve un UUID.
// Envolverlo (en vez de usar uuid.UUID o string a pelo por todo el
// código) nos da un punto único de cambio y hace explícito, en la
// firma de cada función, que ese parámetro es un identificador de
// dominio y no un string cualquiera.
type ID uuid.UUID

func NewID() ID {
	return ID(uuid.New())
}

func ParseID(s string) (ID, error) {
	u, err := uuid.Parse(s)
	if err != nil {
		return ID{}, NewDomainError("INVALID_ID", "el identificador proporcionado no es un UUID válido")
	}
	return ID(u), nil
}

func (id ID) String() string {
	return uuid.UUID(id).String()
}

func (id ID) IsZero() bool {
	return id == ID{}
}
