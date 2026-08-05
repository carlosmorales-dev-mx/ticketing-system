package out

import (
	"context"
	"time"

	"github.com/carlosmorales-dev-mx/ticketing-system/api/internal/domain/shared"
)

// ReservationCache es el puerto hacia Redis. Se usa para materializar
// el TTL de 10 minutos: al reservar, se escribe una clave con
// expiración; cuando Redis la expira, el adaptador de entrada
// amqpconsumer/redis-subscriber reacciona liberando el asiento.
type ReservationCache interface {
	SetWithTTL(ctx context.Context, reservationID shared.ID, seatID shared.ID, ttl time.Duration) error
	Delete(ctx context.Context, reservationID shared.ID) error
	// SubscribeExpirations escucha el canal de keyspace notifications
	// de Redis (`__keyevent@0__:expired`) y entrega el ID de cada
	// clave de reserva que acaba de expirar.
	SubscribeExpirations(ctx context.Context) (<-chan shared.ID, error)
}

// EventPublisher es el puerto hacia RabbitMQ: publica eventos de
// dominio (ej. "seat.reserved", "reservation.confirmed") para que
// otros componentes (la sala de espera virtual, el hub de WebSockets,
// futuros microservicios) reaccionen de forma asíncrona y desacoplada.
type EventPublisher interface {
	Publish(ctx context.Context, topic string, payload []byte) error
}

// RealtimeNotifier es el puerto hacia el hub de WebSockets: informa a
// los clientes conectados de cambios de estado de asientos en vivo.
type RealtimeNotifier interface {
	BroadcastSeatUpdate(ctx context.Context, eventID, seatID shared.ID, status string) error
}
