package usecase

import (
	"context"
	"fmt"

	usecasein "github.com/carlosmorales-dev-mx/ticketing-system/api/internal/application/port/in"
	"github.com/carlosmorales-dev-mx/ticketing-system/api/internal/application/port/out"
	"github.com/carlosmorales-dev-mx/ticketing-system/api/internal/domain/reservation"
	"github.com/carlosmorales-dev-mx/ticketing-system/api/internal/domain/ticket"
)

type confirmPaymentUseCase struct {
	reservationRepo out.ReservationRepository
	seatRepo        out.SeatRepository
	ticketRepo      out.TicketRepository
	reservationCache out.ReservationCache
	notifier        out.RealtimeNotifier
}

func NewConfirmPaymentUseCase(
	reservationRepo out.ReservationRepository,
	seatRepo out.SeatRepository,
	ticketRepo out.TicketRepository,
	reservationCache out.ReservationCache,
	notifier out.RealtimeNotifier,
) usecasein.ConfirmPaymentUseCase {
	return &confirmPaymentUseCase{
		reservationRepo:  reservationRepo,
		seatRepo:         seatRepo,
		ticketRepo:       ticketRepo,
		reservationCache: reservationCache,
		notifier:         notifier,
	}
}

func (uc *confirmPaymentUseCase) Execute(ctx context.Context, cmd usecasein.ConfirmPaymentCommand) (*usecasein.ConfirmPaymentResult, error) {
	res, err := uc.reservationRepo.FindByID(ctx, cmd.ReservationID)
	if err != nil {
		return nil, fmt.Errorf("error buscando la reserva: %w", err)
	}

	// La regla "solo PENDING puede confirmarse" vive en el dominio
	// (reservation.Confirm), el caso de uso solo la invoca.
	if err := res.Confirm(); err != nil {
		return nil, err
	}

	if err := uc.seatRepo.MarkSold(ctx, res.SeatID()); err != nil {
		return nil, fmt.Errorf("error marcando el asiento como vendido: %w", err)
	}

	if err := uc.reservationRepo.UpdateStatus(ctx, res.ID(), reservation.StatusConfirmed); err != nil {
		return nil, fmt.Errorf("error actualizando estado de la reserva: %w", err)
	}

	t := ticket.NewTicket(res.ID(), res.SeatID(), res.UserID())
	if err := uc.ticketRepo.Save(ctx, t); err != nil {
		return nil, fmt.Errorf("error guardando el ticket: %w", err)
	}

	// Ya no hace falta el TTL en Redis: la reserva pasó a un estado final.
	_ = uc.reservationCache.Delete(ctx, res.ID())
	_ = uc.notifier.BroadcastSeatUpdate(ctx, res.EventID(), res.SeatID(), "SOLD")

	return &usecasein.ConfirmPaymentResult{TicketID: t.ID()}, nil
}
