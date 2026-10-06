package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/carlosmorales-dev-mx/ticketing-system/api/internal/adapters/in/http/dto"
	"github.com/carlosmorales-dev-mx/ticketing-system/api/internal/domain/shared"
)

// writeError es el ÚNICO lugar donde un shared.DomainError se traduce
// a un status HTTP. Centralizarlo aquí evita que cada handler
// reinvente su propio mapeo (y se le olvide algún caso).
func writeError(w http.ResponseWriter, err error) {
	var domainErr *shared.DomainError
	if errors.As(err, &domainErr) {
		status := statusForCode(domainErr.Code)
		writeJSON(w, status, dto.ErrorResponse{Error: domainErr.Code})
		return
	}

	// Cualquier error que NO sea un DomainError es un fallo interno
	// (BD caída, bug, etc.) — nunca se expone el mensaje real al
	// cliente, solo se loguea en el servidor.
	log.Printf("error interno no controlado: %v", err)
	writeJSON(w, http.StatusInternalServerError, dto.ErrorResponse{Error: "INTERNAL_ERROR"})
}

func statusForCode(code string) int {
	switch code {
	case "SEAT_ALREADY_RESERVED", "SEAT_NOT_RESERVED", "RESERVATION_NOT_PENDING", "MAX_SEATS_PER_USER":
		return http.StatusConflict // 409
	case "SEAT_NOT_FOUND", "RESERVATION_NOT_FOUND", "TICKET_NOT_FOUND", "EVENT_NOT_FOUND":
		return http.StatusNotFound // 404
	case "INVALID_ID", "INVALID_EVENT_NAME", "INVALID_EVENT_DATE", "VALIDATION_ERROR":
		return http.StatusBadRequest // 400
	default:
		return http.StatusInternalServerError
	}
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Printf("error serializando respuesta JSON: %v", err)
	}
}
