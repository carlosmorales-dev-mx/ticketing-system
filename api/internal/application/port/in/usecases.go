package in

import (
	"context"

	"github.com/carlosmorales-dev-mx/ticketing-system/api/internal/domain/shared"
)

// Estos son los puertos de ENTRADA: la "puerta" por la que el mundo
// exterior (un handler HTTP, un test, un futuro CLI) invoca al
// dominio. El adaptador HTTP depende de estas interfaces, no al
// revés — así el handler no sabe (ni le importa) si por debajo hay
// Postgres, Redis o una implementación en memoria para tests.

type ReserveSeatCommand struct {
	EventID shared.ID
	SeatID  shared.ID
	UserID  shared.ID
}

type ReserveSeatResult struct {
	ReservationID shared.ID
	ExpiresInSec  int
}

type ReserveSeatUseCase interface {
	Execute(ctx context.Context, cmd ReserveSeatCommand) (*ReserveSeatResult, error)
}

type ConfirmPaymentCommand struct {
	ReservationID shared.ID
}

type ConfirmPaymentResult struct {
	TicketID shared.ID
}

type ConfirmPaymentUseCase interface {
	Execute(ctx context.Context, cmd ConfirmPaymentCommand) (*ConfirmPaymentResult, error)
}

type ReleaseExpiredReservationUseCase interface {
	Execute(ctx context.Context, reservationID shared.ID) error
}

type SeatSummary struct {
	SeatID shared.ID
	Row    int
	Label  string
	Status string
}

type ListAvailableSeatsUseCase interface {
	Execute(ctx context.Context, eventID shared.ID) ([]SeatSummary, error)
}
