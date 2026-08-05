package usecase

import (
	"context"
	"fmt"

	usecasein "github.com/carlosmorales-dev-mx/ticketing-system/api/internal/application/port/in"
	"github.com/carlosmorales-dev-mx/ticketing-system/api/internal/application/port/out"
	"github.com/carlosmorales-dev-mx/ticketing-system/api/internal/domain/reservation"
)

// cancelReservationUseCase le da al usuario una salida voluntaria: si
// se arrepiente de un asiento reservado, no tiene que esperar los 10
// minutos completos del TTL para que quede libre otra vez.
type cancelReservationUseCase struct {
	reservationRepo  out.ReservationRepository
	seatRepo         out.SeatRepository
	reservationCache out.ReservationCache
	notifier         out.RealtimeNotifier
}

func NewCancelReservationUseCase(
	reservationRepo out.ReservationRepository,
	seatRepo out.SeatRepository,
	reservationCache out.ReservationCache,
	notifier out.RealtimeNotifier,
) usecasein.CancelReservationUseCase {
	return &cancelReservationUseCase{
		reservationRepo:  reservationRepo,
		seatRepo:         seatRepo,
		reservationCache: reservationCache,
		notifier:         notifier,
	}
}

func (uc *cancelReservationUseCase) Execute(ctx context.Context, cmd usecasein.CancelReservationCommand) error {
	res, err := uc.reservationRepo.FindByID(ctx, cmd.ReservationID)
	if err != nil {
		return fmt.Errorf("error buscando la reserva a cancelar: %w", err)
	}

	// La regla "solo PENDING puede cancelarse" vive en el dominio
	// (reservation.Cancel): si ya se confirmó, expiró, o se canceló
	// antes, devuelve RESERVATION_NOT_PENDING (409) y no seguimos.
	if err := res.Cancel(); err != nil {
		return err
	}

	if err := uc.reservationRepo.UpdateStatus(ctx, res.ID(), reservation.StatusCancelled); err != nil {
		return fmt.Errorf("error actualizando estado de la reserva cancelada: %w", err)
	}

	if err := uc.seatRepo.Release(ctx, res.SeatID()); err != nil {
		return fmt.Errorf("error liberando el asiento tras la cancelación: %w", err)
	}

	// Best-effort: ya la reserva quedó CANCELLED en Postgres (la
	// fuente de verdad); si esto falla solo queda un TTL fantasma en
	// Redis que expirará solo y no hará nada porque Expire() en un
	// dominio ya no-PENDING es un no-op.
	_ = uc.reservationCache.Delete(ctx, res.ID())
	_ = uc.notifier.BroadcastSeatUpdate(ctx, res.EventID(), res.SeatID(), "AVAILABLE")

	return nil
}
