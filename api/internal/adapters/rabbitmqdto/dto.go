// Package rabbitmqdto define el formato "sobre el cable" (wire format)
// compartido entre el cliente RPC (adapters/out/rabbitmq) y el
// trabajador consumidor (adapters/in/amqpconsumer). Vive fuera de
// ambos para no duplicar structs ni crear una dependencia circular
// entre el adaptador de entrada y el de salida.
package rabbitmqdto

// ReservationQueueName es la cola de la "sala de espera virtual":
// absorbe ráfagas de peticiones de reserva y las entrega a un pool
// fijo de trabajadores, que las procesan contra Postgres a un ritmo
// controlado en vez de dejar que cada request HTTP golpee la base de
// datos directamente.
const ReservationQueueName = "reservation.requests"

type ReservationRequest struct {
	EventID string `json:"event_id"`
	SeatID  string `json:"seat_id"`
	UserID  string `json:"user_id"`
}

// ReservationResponse: si Error está vacío, la reserva fue exitosa y
// ReservationID/ExpiresIn son válidos. Si Error tiene contenido, es
// el código de shared.DomainError (SEAT_ALREADY_RESERVED, etc.) tal
// como lo documenta el openapi.yaml.
type ReservationResponse struct {
	ReservationID string `json:"reservation_id,omitempty"`
	ExpiresIn     int    `json:"expires_in,omitempty"`
	Error         string `json:"error,omitempty"`
}
