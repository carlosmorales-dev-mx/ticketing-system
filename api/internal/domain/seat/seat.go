package seat

import (
	"github.com/carlosmorales-dev-mx/ticketing-system/api/internal/domain/shared"
)

// Status representa el ciclo de vida de un asiento.
// AVAILABLE -> RESERVED -> SOLD
// AVAILABLE -> RESERVED -> AVAILABLE (si expira la reserva sin pago)
type Status string

const (
	StatusAvailable Status = "AVAILABLE"
	StatusReserved  Status = "RESERVED"
	StatusSold      Status = "SOLD"
)

// Seat es la entidad raíz del agregado Asiento. Toda mutación de estado
// pasa por sus métodos: así garantizamos que es IMPOSIBLE construir un
// asiento en un estado inconsistente desde fuera del dominio.
type Seat struct {
	id      shared.ID
	eventID shared.ID
	row     int    // fila, ej: 12
	label   string // letra del asiento dentro de la fila, ej: "C"
	status  Status
}

func NewSeat(eventID shared.ID, row int, label string) *Seat {
	return &Seat{
		id:      shared.NewID(),
		eventID: eventID,
		row:     row,
		label:   label,
		status:  StatusAvailable,
	}
}

// Reconstruct reconstruye un Seat desde persistencia (usado por el
// adaptador de Postgres al mapear una fila a entidad). No re-valida
// invariantes de creación porque el dato ya existía y fue válido.
func Reconstruct(id, eventID shared.ID, row int, label string, status Status) *Seat {
	return &Seat{id: id, eventID: eventID, row: row, label: label, status: status}
}

func (s *Seat) ID() shared.ID      { return s.id }
func (s *Seat) EventID() shared.ID { return s.eventID }
func (s *Seat) Row() int           { return s.row }
func (s *Seat) Label() string      { return s.label }
func (s *Seat) Status() Status     { return s.status }

// Reserve intenta pasar el asiento a estado RESERVED.
// Esta es LA regla de negocio central del sistema: solo un asiento
// AVAILABLE puede reservarse. Cualquier otro estado es un 409 Conflict
// en la capa HTTP, pero eso el dominio ni lo sabe ni le importa.
func (s *Seat) Reserve() error {
	if s.status != StatusAvailable {
		return shared.NewDomainError(
			"SEAT_ALREADY_RESERVED",
			"el asiento no está disponible para reservar",
		)
	}
	s.status = StatusReserved
	return nil
}

// Release libera un asiento reservado (por expiración de TTL o
// cancelación explícita), devolviéndolo a AVAILABLE.
func (s *Seat) Release() error {
	if s.status != StatusReserved {
		return shared.NewDomainError(
			"SEAT_NOT_RESERVED",
			"el asiento no está en estado reservado, no se puede liberar",
		)
	}
	s.status = StatusAvailable
	return nil
}

// Sell confirma la venta de un asiento previamente reservado.
func (s *Seat) Sell() error {
	if s.status != StatusReserved {
		return shared.NewDomainError(
			"SEAT_NOT_RESERVED",
			"el asiento debe estar reservado antes de poder venderse",
		)
	}
	s.status = StatusSold
	return nil
}
