package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/carlosmorales-dev-mx/ticketing-system/api/internal/application/port/out"
	"github.com/carlosmorales-dev-mx/ticketing-system/api/internal/domain/seat"
	"github.com/carlosmorales-dev-mx/ticketing-system/api/internal/domain/shared"
)

type seatRepository struct {
	pool *pgxpool.Pool
}

func NewSeatRepository(pool *pgxpool.Pool) out.SeatRepository {
	return &seatRepository{pool: pool}
}

func (r *seatRepository) FindByID(ctx context.Context, id shared.ID) (*seat.Seat, error) {
	const q = `
		SELECT id::text, event_id::text, row_number, label, status
		FROM seats
		WHERE id = $1`

	var (
		seatIDStr, eventIDStr string
		row                   int
		label, status         string
	)

	err := r.pool.QueryRow(ctx, q, id.String()).Scan(&seatIDStr, &eventIDStr, &row, &label, &status)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, shared.NewDomainError("SEAT_NOT_FOUND", "el asiento no existe")
	}
	if err != nil {
		return nil, fmt.Errorf("error consultando asiento: %w", err)
	}

	sID, err := shared.ParseID(seatIDStr)
	if err != nil {
		return nil, fmt.Errorf("id de asiento corrupto en base de datos: %w", err)
	}
	eID, err := shared.ParseID(eventIDStr)
	if err != nil {
		return nil, fmt.Errorf("id de evento corrupto en base de datos: %w", err)
	}
	return seat.Reconstruct(sID, eID, row, label, seat.Status(status)), nil
}

func (r *seatRepository) ListByEvent(ctx context.Context, eventID shared.ID) ([]*seat.Seat, error) {
	const q = `
		SELECT id::text, event_id::text, row_number, label, status
		FROM seats
		WHERE event_id = $1
		ORDER BY row_number, label`

	rows, err := r.pool.Query(ctx, q, eventID.String())
	if err != nil {
		return nil, fmt.Errorf("error listando asientos: %w", err)
	}
	defer rows.Close()

	var result []*seat.Seat
	for rows.Next() {
		var (
			seatIDStr, evIDStr string
			row                int
			label, status      string
		)
		if err := rows.Scan(&seatIDStr, &evIDStr, &row, &label, &status); err != nil {
			return nil, fmt.Errorf("error leyendo fila de asiento: %w", err)
		}
		sID, err := shared.ParseID(seatIDStr)
		if err != nil {
			return nil, fmt.Errorf("id de asiento corrupto en base de datos: %w", err)
		}
		eID, err := shared.ParseID(evIDStr)
		if err != nil {
			return nil, fmt.Errorf("id de evento corrupto en base de datos: %w", err)
		}
		result = append(result, seat.Reconstruct(sID, eID, row, label, seat.Status(status)))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterando asientos: %w", err)
	}
	return result, nil
}

// TryReserve es EL método crítico de todo el sistema.
//
// El UPDATE condicional (WHERE status = 'AVAILABLE') es atómico a
// nivel de fila en Postgres: si dos requests concurrentes ejecutan
// este mismo UPDATE al mismo tiempo sobre el mismo seat_id, Postgres
// serializa internamente el acceso a esa fila (row-level lock
// implícito del UPDATE) y solo UNA de las dos transacciones verá
// RowsAffected() == 1. La otra verá 0 filas afectadas porque, para
// cuando su UPDATE se ejecuta, el status ya cambió.
//
// Esto es más simple y con menos riesgo de deadlock que un
// `SELECT ... FOR UPDATE` explícito seguido de un UPDATE separado,
// y es la técnica recomendada por la documentación de Postgres para
// "compare-and-swap" a nivel de fila.
func (r *seatRepository) TryReserve(ctx context.Context, seatID shared.ID) (bool, error) {
	const q = `
		UPDATE seats
		SET status = 'RESERVED'
		WHERE id = $1 AND status = 'AVAILABLE'`

	tag, err := r.pool.Exec(ctx, q, seatID.String())
	if err != nil {
		return false, fmt.Errorf("error ejecutando TryReserve: %w", err)
	}
	return tag.RowsAffected() == 1, nil
}

func (r *seatRepository) Release(ctx context.Context, seatID shared.ID) error {
	const q = `
		UPDATE seats
		SET status = 'AVAILABLE'
		WHERE id = $1 AND status = 'RESERVED'`

	tag, err := r.pool.Exec(ctx, q, seatID.String())
	if err != nil {
		return fmt.Errorf("error liberando asiento: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return shared.NewDomainError("SEAT_NOT_RESERVED", "el asiento no estaba reservado")
	}
	return nil
}

func (r *seatRepository) MarkSold(ctx context.Context, seatID shared.ID) error {
	const q = `
		UPDATE seats
		SET status = 'SOLD'
		WHERE id = $1 AND status = 'RESERVED'`

	tag, err := r.pool.Exec(ctx, q, seatID.String())
	if err != nil {
		return fmt.Errorf("error marcando asiento como vendido: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return shared.NewDomainError("SEAT_NOT_RESERVED", "el asiento debe estar reservado antes de venderse")
	}
	return nil
}
