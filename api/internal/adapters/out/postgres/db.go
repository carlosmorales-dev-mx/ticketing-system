package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPool crea un pool de conexiones con límites explícitos.
//
// Por qué límites explícitos: sin MaxConns, una ráfaga de tráfico
// puede abrir cientos de conexiones y tumbar Postgres. MaxConns=25
// es un valor conservador razonable para una instancia de desarrollo/
// portfolio; en producción se calcula según núcleos de la BD.
func NewPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("configuración de postgres inválida: %w", err)
	}

	cfg.MaxConns = 25
	cfg.MinConns = 2
	cfg.MaxConnLifetime = 30 * time.Minute
	cfg.MaxConnIdleTime = 5 * time.Minute
	cfg.HealthCheckPeriod = 30 * time.Second

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("no se pudo crear el pool de postgres: %w", err)
	}

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("no se pudo conectar a postgres: %w", err)
	}

	return pool, nil
}
