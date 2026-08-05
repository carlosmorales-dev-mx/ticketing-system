package handler

import (
	"net/http"

	usecasein "github.com/carlosmorales-dev-mx/ticketing-system/api/internal/application/port/in"
	"github.com/carlosmorales-dev-mx/ticketing-system/api/internal/adapters/in/http/dto"
	"github.com/carlosmorales-dev-mx/ticketing-system/api/internal/domain/shared"
)

type SeatHandler struct {
	listSeats usecasein.ListAvailableSeatsUseCase
}

func NewSeatHandler(listSeats usecasein.ListAvailableSeatsUseCase) *SeatHandler {
	return &SeatHandler{listSeats: listSeats}
}

// ListSeats -> GET /events/{eventID}/seats
func (h *SeatHandler) ListSeats(w http.ResponseWriter, r *http.Request) {
	eventID, err := shared.ParseID(r.PathValue("eventID"))
	if err != nil {
		writeError(w, err)
		return
	}

	seats, err := h.listSeats.Execute(r.Context(), eventID)
	if err != nil {
		writeError(w, err)
		return
	}

	response := make([]dto.SeatResponse, 0, len(seats))
	for _, s := range seats {
		response = append(response, dto.SeatResponse{
			SeatID: s.SeatID.String(),
			Row:    s.Row,
			Label:  s.Label,
			Status: s.Status,
		})
	}
	writeJSON(w, http.StatusOK, response)
}
