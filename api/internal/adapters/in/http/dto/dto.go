// Package dto define los contratos JSON de la API. Nunca se filtran
// hacia application/domain: el handler los traduce a Commands y de
// vuelta.
package dto

type SeatResponse struct {
	SeatID string `json:"seat_id"`
	Row    int    `json:"row"`
	Label  string `json:"label"`
	Status string `json:"status"`
}

type ReserveSeatRequest struct {
	EventID string `json:"event_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	SeatID  string `json:"seat_id" example:"7c9e6679-7425-40de-944b-e07fc1f90ae7"`
	UserID  string `json:"user_id" example:"9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d"`
}

type ReserveSeatResponse struct {
	ReservationID string `json:"reservation_id"`
	ExpiresInSec  int    `json:"expires_in"`
}

type ConfirmPaymentResponse struct {
	TicketID string `json:"ticket_id"`
}

// ErrorResponse es el formato uniforme de error de toda la API.
// Este es el shape que se documenta en el 409/400/429 de OpenAPI.
type ErrorResponse struct {
	Error string `json:"error"`
}

type ResetEventResponse struct {
	ReleasedSeats int `json:"released_seats"`
}
