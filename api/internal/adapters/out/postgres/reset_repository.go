package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/carlosmorales-dev-mx/ticketing-system/api/internal/application/port/out"
	"github.com/carlosmorales-dev-mx/ticketing-system/api/internal/domain/shared"
)

type eventResetRepository struct {
	pool *pgxpool.Pool
}

func NewEventResetRepository(pool *pgxpool.Pool) out.EventResetRepository {
	return &eventResetRepository{pool: pool}
}

// ResetEvent deja el evento como recién sembrado, todo en UNA transacción:
//  1. anota qué reservas estaban PENDING y qué asientos no estaban libres,
//  2. borra las reservas del evento (los tickets caen en cascada por
//     tickets.reservation_id ... ON DELETE CASCADE),
//  3. devuelve todos los asientos a AVAILABLE.
func (r *eventResetRepository) ResetEvent(ctx context.Context, eventID shared.ID) (*out.EventResetResult, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("error abriendo transacción de reinicio: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }() // no-op si ya hubo Commit

	// Bloquea la fila del evento: dos reinicios simultáneos se hacen fila.
	var exists bool
	err = tx.QueryRow(ctx, `SELECT true FROM events WHERE id = $1 FOR UPDATE`, eventID.String()).Scan(&exists)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, shared.NewDomainError("EVENT_NOT_FOUND", "el evento no existe")
	}
	if err != nil {
		return nil, fmt.Errorf("error comprobando el evento: %w", err)
	}

	pending, err := collectIDs(ctx, tx,
		`SELECT id::text FROM reservations WHERE event_id = $1 AND status = 'PENDING'`, eventID)
	if err != nil {
		return nil, err
	}
	released, err := collectIDs(ctx, tx,
		`SELECT id::text FROM seats WHERE event_id = $1 AND status <> 'AVAILABLE'`, eventID)
	if err != nil {
		return nil, err
	}

	if _, err := tx.Exec(ctx, `DELETE FROM reservations WHERE event_id = $1`, eventID.String()); err != nil {
		return nil, fmt.Errorf("error borrando reservas del evento: %w", err)
	}
	if _, err := tx.Exec(ctx,
		`UPDATE seats SET status = 'AVAILABLE' WHERE event_id = $1 AND status <> 'AVAILABLE'`, eventID.String()); err != nil {
		return nil, fmt.Errorf("error liberando asientos del evento: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("error confirmando el reinicio: %w", err)
	}
	return &out.EventResetResult{ReleasedSeatIDs: released, PendingReservationIDs: pending}, nil
}

func collectIDs(ctx context.Context, tx pgx.Tx, query string, eventID shared.ID) ([]shared.ID, error) {
	rows, err := tx.Query(ctx, query, eventID.String())
	if err != nil {
		return nil, fmt.Errorf("error consultando ids del evento: %w", err)
	}
	defer rows.Close()

	var ids []shared.ID
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			return nil, fmt.Errorf("error leyendo id: %w", err)
		}
		id, err := shared.ParseID(raw)
		if err != nil {
			return nil, fmt.Errorf("id corrupto en base de datos: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterando ids: %w", err)
	}
	return ids, nil
}
