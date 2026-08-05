package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/carlosmorales-dev-mx/ticketing-system/api/internal/application/port/out"
	"github.com/carlosmorales-dev-mx/ticketing-system/api/internal/domain/reservation"
	"github.com/carlosmorales-dev-mx/ticketing-system/api/internal/domain/shared"
)

type reservationRepository struct {
	pool *pgxpool.Pool
}

func NewReservationRepository(pool *pgxpool.Pool) out.ReservationRepository {
	return &reservationRepository{pool: pool}
}

func (r *reservationRepository) Save(ctx context.Context, res *reservation.Reservation) error {
	const q = `
		INSERT INTO reservations (id, event_id, seat_id, user_id, status, created_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, now(), $6)`

	_, err := r.pool.Exec(ctx, q,
		res.ID().String(),
		res.EventID().String(),
		res.SeatID().String(),
		res.UserID().String(),
		string(res.Status()),
		res.ExpiresAt(),
	)
	if err != nil {
		// El índice único parcial uq_seat_pending_reservation es el
		// segundo cinturón de seguridad: si por cualquier carrera
		// extraña llegaran dos INSERT con el mismo seat_id en PENDING,
		// Postgres rechaza el segundo con un error de constraint.
		return fmt.Errorf("error guardando reserva (posible violación de unicidad): %w", err)
	}
	return nil
}

func (r *reservationRepository) FindByID(ctx context.Context, id shared.ID) (*reservation.Reservation, error) {
	const q = `
		SELECT id::text, event_id::text, seat_id::text, user_id::text, status, created_at, expires_at
		FROM reservations
		WHERE id = $1`

	var (
		resID, eventID, seatID, userID string
		status                         string
		createdAt, expiresAt           time.Time
	)

	err := r.pool.QueryRow(ctx, q, id.String()).
		Scan(&resID, &eventID, &seatID, &userID, &status, &createdAt, &expiresAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, shared.NewDomainError("RESERVATION_NOT_FOUND", "la reserva no existe")
	}
	if err != nil {
		return nil, fmt.Errorf("error consultando reserva: %w", err)
	}

	rID, err := shared.ParseID(resID)
	if err != nil {
		return nil, fmt.Errorf("id de reserva corrupto en base de datos: %w", err)
	}
	eID, err := shared.ParseID(eventID)
	if err != nil {
		return nil, fmt.Errorf("id de evento corrupto en base de datos: %w", err)
	}
	sID, err := shared.ParseID(seatID)
	if err != nil {
		return nil, fmt.Errorf("id de asiento corrupto en base de datos: %w", err)
	}
	uID, err := shared.ParseID(userID)
	if err != nil {
		return nil, fmt.Errorf("id de usuario corrupto en base de datos: %w", err)
	}

	return reservation.Reconstruct(rID, eID, sID, uID, reservation.Status(status), createdAt, expiresAt), nil
}

func (r *reservationRepository) UpdateStatus(ctx context.Context, id shared.ID, status reservation.Status) error {
	const q = `UPDATE reservations SET status = $2 WHERE id = $1`

	tag, err := r.pool.Exec(ctx, q, id.String(), string(status))
	if err != nil {
		return fmt.Errorf("error actualizando estado de reserva: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return shared.NewDomainError("RESERVATION_NOT_FOUND", "la reserva no existe")
	}
	return nil
}

// FindExpiredPending usa el índice parcial idx_reservations_expires_at
// (ver migración 001_init_schema.sql) para que este barrido sea barato
// incluso con muchas reservas históricas en la tabla.
func (r *reservationRepository) FindExpiredPending(ctx context.Context, before time.Time) ([]shared.ID, error) {
	const q = `
		SELECT id::text FROM reservations
		WHERE status = 'PENDING' AND expires_at < $1`

	rows, err := r.pool.Query(ctx, q, before)
	if err != nil {
		return nil, fmt.Errorf("error buscando reservas vencidas: %w", err)
	}
	defer rows.Close()

	var ids []shared.ID
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			return nil, fmt.Errorf("error leyendo id de reserva vencida: %w", err)
		}
		id, err := shared.ParseID(raw)
		if err != nil {
			return nil, fmt.Errorf("id de reserva corrupto en base de datos: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterando reservas vencidas: %w", err)
	}
	return ids, nil
}
