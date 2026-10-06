package out

import (
	"context"
	"time"

	"github.com/carlosmorales-dev-mx/ticketing-system/api/internal/domain/reservation"
	"github.com/carlosmorales-dev-mx/ticketing-system/api/internal/domain/seat"
	"github.com/carlosmorales-dev-mx/ticketing-system/api/internal/domain/shared"
	"github.com/carlosmorales-dev-mx/ticketing-system/api/internal/domain/ticket"
)

// SeatRepository es el puerto de salida hacia la persistencia de
// asientos. La implementación real (Postgres) vive en
// adapters/out/postgres. El dominio y los casos de uso solo conocen
// esta interfaz.
//
// TryReserve es el método CRÍTICO de concurrencia: debe implementarse
// con un UPDATE ... WHERE status = 'AVAILABLE' atómico (o
// SELECT ... FOR UPDATE) para que, bajo carga, jamás dos requests
// reserven el mismo asiento. Devuelve false si no se pudo reservar
// porque ya no estaba disponible (-> 409 en la capa HTTP).
type SeatRepository interface {
	FindByID(ctx context.Context, id shared.ID) (*seat.Seat, error)
	ListByEvent(ctx context.Context, eventID shared.ID) ([]*seat.Seat, error)
	TryReserve(ctx context.Context, seatID shared.ID) (bool, error)
	Release(ctx context.Context, seatID shared.ID) error
	MarkSold(ctx context.Context, seatID shared.ID) error
}

type ReservationRepository interface {
	Save(ctx context.Context, r *reservation.Reservation) error
	FindByID(ctx context.Context, id shared.ID) (*reservation.Reservation, error)
	UpdateStatus(ctx context.Context, id shared.ID, status reservation.Status) error
	// FindExpiredPending es la red de seguridad del sistema: las
	// keyspace notifications de Redis son best-effort, así que un
	// barrido periódico contra este método (ver cmd/api/main.go)
	// libera cualquier reserva PENDING cuyo TTL ya pasó pero cuyo
	// evento de expiración se perdió (reinicio de Redis, red, etc.).
	FindExpiredPending(ctx context.Context, before time.Time) ([]shared.ID, error)
	// CountPendingByUser cuenta las reservas PENDING todavía vigentes
	// (expires_at en el futuro) de un usuario en un evento. Se usa para
	// aplicar reservation.MaxPendingPerUser.
	CountPendingByUser(ctx context.Context, eventID, userID shared.ID) (int, error)
}

// EventResetResult describe qué cambió al reiniciar un evento, para que
// el caso de uso pueda limpiar Redis y avisar por WebSocket.
type EventResetResult struct {
	// ReleasedSeatIDs son los asientos que NO estaban AVAILABLE y
	// ahora lo están (reservados o vendidos hasta ese momento).
	ReleasedSeatIDs []shared.ID
	// PendingReservationIDs son las reservas PENDING que se borraron;
	// sus claves de TTL en Redis ya no tienen sentido.
	PendingReservationIDs []shared.ID
}

// EventResetRepository es el puerto de la herramienta de desarrollo
// "reiniciar mapa": deja el evento como recién sembrado. Debe ser una
// sola transacción: o queda todo limpio o no cambia nada.
type EventResetRepository interface {
	ResetEvent(ctx context.Context, eventID shared.ID) (*EventResetResult, error)
}

type TicketRepository interface {
	Save(ctx context.Context, t *ticket.Ticket) error
	FindByID(ctx context.Context, id shared.ID) (*ticket.Ticket, error)
}
