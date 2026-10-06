package handler

import (
	"net/http"

	"github.com/carlosmorales-dev-mx/ticketing-system/api/internal/adapters/in/http/dto"
	usecasein "github.com/carlosmorales-dev-mx/ticketing-system/api/internal/application/port/in"
	"github.com/carlosmorales-dev-mx/ticketing-system/api/internal/domain/shared"
)

type EventHandler struct {
	resetEvent usecasein.ResetEventUseCase
	resetOn    bool
}

// resetOn viene de ENABLE_MAP_RESET: es una herramienta de desarrollo
// y por defecto está apagada.
func NewEventHandler(resetEvent usecasein.ResetEventUseCase, resetOn bool) *EventHandler {
	return &EventHandler{resetEvent: resetEvent, resetOn: resetOn}
}

// Reset -> POST /events/{eventID}/reset
// 200 -> mapa reiniciado: {"released_seats": N}
// 400 -> eventID no es un UUID
// 403 -> MAP_RESET_DISABLED (la API no se arrancó con ENABLE_MAP_RESET=true)
// 404 -> EVENT_NOT_FOUND
func (h *EventHandler) Reset(w http.ResponseWriter, r *http.Request) {
	if !h.resetOn {
		writeJSON(w, http.StatusForbidden, dto.ErrorResponse{Error: "MAP_RESET_DISABLED"})
		return
	}

	eventID, err := shared.ParseID(r.PathValue("eventID"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{Error: "INVALID_EVENT_ID"})
		return
	}

	result, err := h.resetEvent.Execute(r.Context(), usecasein.ResetEventCommand{EventID: eventID})
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dto.ResetEventResponse{ReleasedSeats: result.ReleasedSeats})
}
