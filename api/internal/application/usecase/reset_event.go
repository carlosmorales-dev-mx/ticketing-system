package usecase

import (
	"context"
	"fmt"

	usecasein "github.com/carlosmorales-dev-mx/ticketing-system/api/internal/application/port/in"
	"github.com/carlosmorales-dev-mx/ticketing-system/api/internal/application/port/out"
)

type resetEventUseCase struct {
	resetRepo out.EventResetRepository
	cache     out.ReservationCache
	notifier  out.RealtimeNotifier
}

func NewResetEventUseCase(
	resetRepo out.EventResetRepository,
	cache out.ReservationCache,
	notifier out.RealtimeNotifier,
) usecasein.ResetEventUseCase {
	return &resetEventUseCase{resetRepo: resetRepo, cache: cache, notifier: notifier}
}

func (uc *resetEventUseCase) Execute(ctx context.Context, cmd usecasein.ResetEventCommand) (*usecasein.ResetEventResult, error) {
	res, err := uc.resetRepo.ResetEvent(ctx, cmd.EventID)
	if err != nil {
		return nil, fmt.Errorf("error reiniciando el evento: %w", err)
	}

	// Postgres ya quedó limpio (fuente de verdad). Lo siguiente es
	// best-effort: si falla, un TTL huérfano en Redis expira solo y
	// ReleaseExpired lo ignora porque la reserva ya no existe.
	for _, id := range res.PendingReservationIDs {
		_ = uc.cache.Delete(ctx, id)
	}
	// Avisamos asiento por asiento a los demás navegadores abiertos
	// para que su mapa se limpie en vivo.
	for _, seatID := range res.ReleasedSeatIDs {
		_ = uc.notifier.BroadcastSeatUpdate(ctx, cmd.EventID, seatID, "AVAILABLE")
	}

	return &usecasein.ResetEventResult{ReleasedSeats: len(res.ReleasedSeatIDs)}, nil
}
