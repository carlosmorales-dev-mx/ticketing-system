package usecase

import (
	"context"
	"fmt"

	usecasein "github.com/carlosmorales-dev-mx/ticketing-system/api/internal/application/port/in"
	"github.com/carlosmorales-dev-mx/ticketing-system/api/internal/application/port/out"
	"github.com/carlosmorales-dev-mx/ticketing-system/api/internal/domain/reservation"
	"github.com/carlosmorales-dev-mx/ticketing-system/api/internal/domain/shared"
)

// releaseExpiredReservationUseCase es invocado por el adaptador que
// escucha las notificaciones de expiración de Redis
// (adapters/in/... -> ver ARCHITECTURE.md, "flujo de expiración").
type releaseExpiredReservationUseCase struct {
	reservationRepo out.ReservationRepository
	seatRepo        out.SeatRepository
	notifier        out.RealtimeNotifier
}

func NewReleaseExpiredReservationUseCase(
	reservationRepo out.ReservationRepository,
	seatRepo out.SeatRepository,
	notifier out.RealtimeNotifier,
) usecasein.ReleaseExpiredReservationUseCase {
	return &releaseExpiredReservationUseCase{
		reservationRepo: reservationRepo,
		seatRepo:        seatRepo,
		notifier:        notifier,
	}
}

func (uc *releaseExpiredReservationUseCase) Execute(ctx context.Context, reservationID shared.ID) error {
	res, err := uc.reservationRepo.FindByID(ctx, reservationID)
	if err != nil {
		return fmt.Errorf("error buscando la reserva a expirar: %w", err)
	}

	// Si ya no está PENDING (por ejemplo, el usuario pagó justo antes
	// de que llegara el evento de expiración), no hacemos nada: es una
	// carrera legítima que el propio dominio resuelve devolviendo error,
	// que aquí simplemente ignoramos como no-op.
	if err := res.Expire(); err != nil {
		return nil
	}

	if err := uc.reservationRepo.UpdateStatus(ctx, res.ID(), reservation.StatusExpired); err != nil {
		return fmt.Errorf("error actualizando estado de la reserva expirada: %w", err)
	}

	if err := uc.seatRepo.Release(ctx, res.SeatID()); err != nil {
		return fmt.Errorf("error liberando el asiento tras expiración: %w", err)
	}

	_ = uc.notifier.BroadcastSeatUpdate(ctx, res.EventID(), res.SeatID(), "AVAILABLE")

	return nil
}
