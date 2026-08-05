package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/carlosmorales-dev-mx/ticketing-system/api/internal/application/port/out"
	"github.com/carlosmorales-dev-mx/ticketing-system/api/internal/domain/shared"
	"github.com/carlosmorales-dev-mx/ticketing-system/api/internal/domain/ticket"
)

type ticketRepository struct {
	pool *pgxpool.Pool
}

func NewTicketRepository(pool *pgxpool.Pool) out.TicketRepository {
	return &ticketRepository{pool: pool}
}

func (r *ticketRepository) Save(ctx context.Context, t *ticket.Ticket) error {
	const q = `
		INSERT INTO tickets (id, reservation_id, seat_id, user_id, issued_at)
		VALUES ($1, $2, $3, $4, $5)`

	_, err := r.pool.Exec(ctx, q,
		t.ID().String(), t.ReservationID().String(), t.SeatID().String(), t.UserID().String(), t.IssuedAt(),
	)
	if err != nil {
		return fmt.Errorf("error guardando ticket: %w", err)
	}
	return nil
}

func (r *ticketRepository) FindByID(ctx context.Context, id shared.ID) (*ticket.Ticket, error) {
	const q = `
		SELECT id::text, reservation_id::text, seat_id::text, user_id::text, issued_at
		FROM tickets WHERE id = $1`

	var (
		ticketID, reservationID, seatID, userID string
		issuedAt                                time.Time
	)

	err := r.pool.QueryRow(ctx, q, id.String()).
		Scan(&ticketID, &reservationID, &seatID, &userID, &issuedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, shared.NewDomainError("TICKET_NOT_FOUND", "el ticket no existe")
	}
	if err != nil {
		return nil, fmt.Errorf("error consultando ticket: %w", err)
	}

	tID, err := shared.ParseID(ticketID)
	if err != nil {
		return nil, fmt.Errorf("id de ticket corrupto en base de datos: %w", err)
	}
	resID, err := shared.ParseID(reservationID)
	if err != nil {
		return nil, fmt.Errorf("id de reserva corrupto en base de datos: %w", err)
	}
	sID, err := shared.ParseID(seatID)
	if err != nil {
		return nil, fmt.Errorf("id de asiento corrupto en base de datos: %w", err)
	}
	uID, err := shared.ParseID(userID)
	if err != nil {
		return nil, fmt.Errorf("id de usuario corrupto en base de datos: %w", err)
	}

	return ticket.Reconstruct(tID, resID, sID, uID, issuedAt), nil
}
