package redis

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/carlosmorales-dev-mx/ticketing-system/api/internal/application/port/out"
	"github.com/carlosmorales-dev-mx/ticketing-system/api/internal/domain/shared"
)

const keyPrefix = "reservation:"

type reservationCache struct {
	client *goredis.Client
}

// NewClient crea un cliente de Redis con timeouts explícitos.
// Sin timeouts, una petición a Redis que se cuelga puede bloquear
// indefinidamente el goroutine de la petición HTTP que la disparó.
func NewClient(addr string) *goredis.Client {
	return goredis.NewClient(&goredis.Options{
		Addr:         addr,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     20,
		MinIdleConns: 2,
	})
}

func NewReservationCache(client *goredis.Client) out.ReservationCache {
	return &reservationCache{client: client}
}

func (c *reservationCache) SetWithTTL(ctx context.Context, reservationID shared.ID, seatID shared.ID, ttl time.Duration) error {
	key := keyPrefix + reservationID.String()
	if err := c.client.Set(ctx, key, seatID.String(), ttl).Err(); err != nil {
		return fmt.Errorf("error escribiendo TTL en redis: %w", err)
	}
	return nil
}

func (c *reservationCache) Delete(ctx context.Context, reservationID shared.ID) error {
	key := keyPrefix + reservationID.String()
	if err := c.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("error borrando clave de reserva en redis: %w", err)
	}
	return nil
}

// SubscribeExpirations escucha el canal de keyspace notifications de
// Redis (`__keyevent@0__:expired`), que se activa gracias a
// `notify-keyspace-events Ex` (ya configurado en podman-compose.yaml).
//
// Nota de robustez: las keyspace notifications de Redis son
// "best-effort" — si Redis se reinicia entre la expiración y la
// entrega del evento, el evento se pierde. Por eso el esquema de
// Postgres incluye un índice sobre `expires_at` (ver migración 001):
// un job de barrido periódico puede usarlo como red de seguridad
// para liberar reservas vencidas que Redis no llegó a notificar.
// Ese barrido no está implementado aquí, pero el índice ya existe
// para cuando se añada.
func (c *reservationCache) SubscribeExpirations(ctx context.Context) (<-chan shared.ID, error) {
	pubsub := c.client.PSubscribe(ctx, "__keyevent@0__:expired")
	if _, err := pubsub.Receive(ctx); err != nil {
		return nil, fmt.Errorf("error suscribiéndose a expiraciones de redis: %w", err)
	}

	out := make(chan shared.ID)
	go func() {
		defer close(out)
		defer pubsub.Close()

		ch := pubsub.Channel()
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}
				// El payload es la clave completa, ej: "reservation:<uuid>"
				raw := strings.TrimPrefix(msg.Payload, keyPrefix)
				id, err := shared.ParseID(raw)
				if err != nil {
					// No es una clave de reserva (podría ser otra
					// clave cualquiera expirando en la misma DB de
					// Redis); la ignoramos sin tumbar el listener.
					log.Printf("aviso: expiración de clave no reconocida en redis: %q", msg.Payload)
					continue
				}
				select {
				case out <- id:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return out, nil
}
