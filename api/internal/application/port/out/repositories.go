package out

import (
	"context"

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
}

type TicketRepository interface {
	Save(ctx context.Context, t *ticket.Ticket) error
	FindByID(ctx context.Context, id shared.ID) (*ticket.Ticket, error)
}
