package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/carlosmorales-dev-mx/ticketing-system/api/internal/adapters/in/http/dto"
	usecasein "github.com/carlosmorales-dev-mx/ticketing-system/api/internal/application/port/in"
	"github.com/carlosmorales-dev-mx/ticketing-system/api/internal/domain/shared"
)

// ReservationEnqueuer es lo que el handler necesita para reservar un
// asiento. Su firma es idéntica a usecasein.ReserveSeatUseCase.Execute
// a propósito: el handler no necesita saber si, por debajo, la
// petición se ejecuta directo o se encola en la sala de espera
// virtual de RabbitMQ (adapters/out/rabbitmq.ReservationRPCClient) —
// eso es una decisión de wiring que se toma en cmd/api/main.go.
type ReservationEnqueuer interface {
	Enqueue(ctx context.Context, cmd usecasein.ReserveSeatCommand) (*usecasein.ReserveSeatResult, error)
}

type ReservationHandler struct {
	reserveSeat       ReservationEnqueuer
	confirmPayment    usecasein.ConfirmPaymentUseCase
	cancelReservation usecasein.CancelReservationUseCase
}

func NewReservationHandler(
	reserveSeat ReservationEnqueuer,
	confirmPayment usecasein.ConfirmPaymentUseCase,
	cancelReservation usecasein.CancelReservationUseCase,
) *ReservationHandler {
	return &ReservationHandler{
		reserveSeat:       reserveSeat,
		confirmPayment:    confirmPayment,
		cancelReservation: cancelReservation,
	}
}

// Reserve -> POST /reservations
// Body: {"event_id": "...", "seat_id": "...", "user_id": "..."}
// 201 -> reserva creada
// 400 -> algún ID no es un UUID válido
// 409 -> SEAT_ALREADY_RESERVED (alguien más lo reservó primero)
func (h *ReservationHandler) Reserve(w http.ResponseWriter, r *http.Request) {
	var req dto.ReserveSeatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{Error: "INVALID_JSON_BODY"})
		return
	}

	eventID, err := shared.ParseID(req.EventID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{Error: "INVALID_EVENT_ID"})
		return
	}
	seatID, err := shared.ParseID(req.SeatID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{Error: "INVALID_SEAT_ID"})
		return
	}
	userID, err := shared.ParseID(req.UserID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{Error: "INVALID_USER_ID"})
		return
	}

	// Esta petición ahora pasa por la sala de espera virtual: se
	// encola en RabbitMQ y un pool fijo de workers la procesa contra
	// Postgres a un ritmo controlado (ver adapters/in/amqpconsumer).
	result, err := h.reserveSeat.Enqueue(r.Context(), usecasein.ReserveSeatCommand{
		EventID: eventID,
		SeatID:  seatID,
		UserID:  userID,
	})
	if err != nil {
		writeError(w, err) // aquí es donde SEAT_ALREADY_RESERVED se convierte en 409
		return
	}

	writeJSON(w, http.StatusCreated, dto.ReserveSeatResponse{
		ReservationID: result.ReservationID.String(),
		ExpiresInSec:  result.ExpiresInSec,
	})
}

// Confirm -> POST /reservations/{reservationID}/confirm
// 200 -> pago confirmado, ticket emitido
// 404 -> la reserva no existe
// 409 -> RESERVATION_NOT_PENDING (ya expiró, ya se confirmó, o se canceló)
func (h *ReservationHandler) Confirm(w http.ResponseWriter, r *http.Request) {
	reservationID, err := shared.ParseID(r.PathValue("reservationID"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{Error: "INVALID_RESERVATION_ID"})
		return
	}

	result, err := h.confirmPayment.Execute(r.Context(), usecasein.ConfirmPaymentCommand{
		ReservationID: reservationID,
	})
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dto.ConfirmPaymentResponse{TicketID: result.TicketID.String()})
}

// Cancel -> POST /reservations/{reservationID}/cancel
// 204 -> cancelada, el asiento vuelve a estar AVAILABLE
// 404 -> la reserva no existe
// 409 -> RESERVATION_NOT_PENDING (ya se confirmó, expiró, o se canceló antes)
func (h *ReservationHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	reservationID, err := shared.ParseID(r.PathValue("reservationID"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{Error: "INVALID_RESERVATION_ID"})
		return
	}

	if err := h.cancelReservation.Execute(r.Context(), usecasein.CancelReservationCommand{
		ReservationID: reservationID,
	}); err != nil {
		writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
