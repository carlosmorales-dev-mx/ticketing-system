package reservation

import (
	"time"

	"github.com/carlosmorales-dev-mx/ticketing-system/api/internal/domain/shared"
)

type Status string

const (
	StatusPending   Status = "PENDING"   // esperando pago
	StatusConfirmed Status = "CONFIRMED" // pago recibido -> se emitió ticket
	StatusExpired   Status = "EXPIRED"   // pasaron los 10 min sin pago
	StatusCancelled Status = "CANCELLED" // cancelada explícitamente por el usuario
)

// Reservation es el "contrato temporal" entre un usuario y un asiento
// mientras completa el pago. Su invariante central es el TTL: una
// reserva PENDING que supera ExpiresAt debe considerarse EXPIRED.
type Reservation struct {
	id        shared.ID
	eventID   shared.ID
	seatID    shared.ID
	userID    shared.ID
	status    Status
	createdAt time.Time
	expiresAt time.Time
}

// DefaultTTL es la ventana de reserva antes de liberar el asiento.
// Vive aquí, en el dominio, porque es una REGLA DE NEGOCIO ("tienes
// 10 minutos para pagar"), no un detalle de infraestructura — aunque
// luego el adaptador de Redis sea quien físicamente implemente el TTL.
const DefaultTTL = 10 * time.Minute

func NewReservation(eventID, seatID, userID shared.ID) *Reservation {
	now := time.Now()
	return &Reservation{
		id:        shared.NewID(),
		eventID:   eventID,
		seatID:    seatID,
		userID:    userID,
		status:    StatusPending,
		createdAt: now,
		expiresAt: now.Add(DefaultTTL),
	}
}

func Reconstruct(id, eventID, seatID, userID shared.ID, status Status, createdAt, expiresAt time.Time) *Reservation {
	return &Reservation{
		id: id, eventID: eventID, seatID: seatID, userID: userID,
		status: status, createdAt: createdAt, expiresAt: expiresAt,
	}
}

func (r *Reservation) ID() shared.ID        { return r.id }
func (r *Reservation) EventID() shared.ID   { return r.eventID }
func (r *Reservation) SeatID() shared.ID    { return r.seatID }
func (r *Reservation) UserID() shared.ID    { return r.userID }
func (r *Reservation) Status() Status       { return r.status }
func (r *Reservation) ExpiresAt() time.Time { return r.expiresAt }

// IsExpired es la regla de negocio pura: no depende de Redis, no
// depende de un cron job. Cualquier capa puede preguntarle a una
// Reservation si ya caducó comparando contra time.Now().
func (r *Reservation) IsExpired(now time.Time) bool {
	return r.status == StatusPending && now.After(r.expiresAt)
}

func (r *Reservation) Confirm() error {
	if r.status != StatusPending {
		return shared.NewDomainError("RESERVATION_NOT_PENDING", "solo una reserva pendiente puede confirmarse")
	}
	r.status = StatusConfirmed
	return nil
}

func (r *Reservation) Expire() error {
	if r.status != StatusPending {
		return shared.NewDomainError("RESERVATION_NOT_PENDING", "solo una reserva pendiente puede expirar")
	}
	r.status = StatusExpired
	return nil
}

func (r *Reservation) Cancel() error {
	if r.status != StatusPending {
		return shared.NewDomainError("RESERVATION_NOT_PENDING", "solo una reserva pendiente puede cancelarse")
	}
	r.status = StatusCancelled
	return nil
}
