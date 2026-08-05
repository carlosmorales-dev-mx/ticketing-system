package usecase

import (
	"context"
	"fmt"

	usecasein "github.com/carlosmorales-dev-mx/ticketing-system/api/internal/application/port/in"
	"github.com/carlosmorales-dev-mx/ticketing-system/api/internal/application/port/out"
	"github.com/carlosmorales-dev-mx/ticketing-system/api/internal/domain/shared"
)

type listAvailableSeatsUseCase struct {
	seatRepo out.SeatRepository
}

func NewListAvailableSeatsUseCase(seatRepo out.SeatRepository) usecasein.ListAvailableSeatsUseCase {
	return &listAvailableSeatsUseCase{seatRepo: seatRepo}
}

func (uc *listAvailableSeatsUseCase) Execute(ctx context.Context, eventID shared.ID) ([]usecasein.SeatSummary, error) {
	seats, err := uc.seatRepo.ListByEvent(ctx, eventID)
	if err != nil {
		return nil, fmt.Errorf("error listando asientos del evento: %w", err)
	}

	summaries := make([]usecasein.SeatSummary, 0, len(seats))
	for _, s := range seats {
		summaries = append(summaries, usecasein.SeatSummary{
			SeatID: s.ID(),
			Row:    s.Row(),
			Label:  s.Label(),
			Status: string(s.Status()),
		})
	}
	return summaries, nil
}
