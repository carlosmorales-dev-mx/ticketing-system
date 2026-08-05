package shared

// DomainError representa una violación de una regla de negocio.
// Se define aquí (en el dominio) y NO en el adaptador HTTP, porque la
// regla "este asiento ya está reservado" es una verdad del negocio,
// no un detalle de transporte. El adaptador HTTP luego decide cómo
// traducir este Code a un status HTTP (ver adapters/in/http).
type DomainError struct {
	Code    string // ej: "SEAT_ALREADY_RESERVED"
	Message string
}

func (e *DomainError) Error() string {
	return e.Message
}

func NewDomainError(code, message string) *DomainError {
	return &DomainError{Code: code, Message: message}
}
