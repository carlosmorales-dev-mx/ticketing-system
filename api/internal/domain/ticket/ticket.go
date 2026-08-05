package ticket

import (
	"time"

	"github.com/carlosmorales-dev-mx/ticketing-system/api/internal/domain/shared"
)

// Ticket se emite SOLO cuando una Reservation pasa a CONFIRMED
// (pago recibido). Es el resultado final del flujo de negocio.
type Ticket struct {
	id            shared.ID
	reservationID shared.ID
	seatID        shared.ID
	userID        shared.ID
	issuedAt      time.Time
}

func NewTicket(reservationID, seatID, userID shared.ID) *Ticket {
	return &Ticket{
		id:            shared.NewID(),
		reservationID: reservationID,
		seatID:        seatID,
		userID:        userID,
		issuedAt:      time.Now(),
	}
}

func Reconstruct(id, reservationID, seatID, userID shared.ID, issuedAt time.Time) *Ticket {
	return &Ticket{id: id, reservationID: reservationID, seatID: seatID, userID: userID, issuedAt: issuedAt}
}

func (t *Ticket) ID() shared.ID            { return t.id }
func (t *Ticket) ReservationID() shared.ID { return t.reservationID }
func (t *Ticket) SeatID() shared.ID        { return t.seatID }
func (t *Ticket) UserID() shared.ID        { return t.userID }
func (t *Ticket) IssuedAt() time.Time      { return t.issuedAt }
