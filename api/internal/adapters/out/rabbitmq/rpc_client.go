package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/carlosmorales-dev-mx/ticketing-system/api/internal/adapters/rabbitmqdto"
	usecasein "github.com/carlosmorales-dev-mx/ticketing-system/api/internal/application/port/in"
	"github.com/carlosmorales-dev-mx/ticketing-system/api/internal/domain/shared"
)

// ReservationRPCClient es el lado "cliente" de la sala de espera
// virtual: en vez de que el handler HTTP invoque el caso de uso
// directamente, publica la petición en la cola
// rabbitmqdto.ReservationQueueName y espera la respuesta en una cola
// exclusiva y temporal, usando el patrón RPC sobre AMQP (correlación
// por CorrelationId + ReplyTo). Esto es lo que realmente absorbe
// ráfagas de tráfico: los workers del otro lado (adapters/in/amqpconsumer)
// controlan cuántas peticiones tocan Postgres a la vez.
type ReservationRPCClient struct {
	channel *amqp.Channel
	timeout time.Duration
}

func NewReservationRPCClient(ch *amqp.Channel) (*ReservationRPCClient, error) {
	_, err := ch.QueueDeclare(rabbitmqdto.ReservationQueueName, true, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("error declarando la cola de reservas: %w", err)
	}
	return &ReservationRPCClient{channel: ch, timeout: 8 * time.Second}, nil
}

// Enqueue empuja la petición a la sala de espera y bloquea hasta
// recibir la respuesta del worker que la procesó (o hasta que expire
// el timeout, lo que proteja al cliente HTTP de esperar para siempre
// si RabbitMQ o los workers tienen un problema).
func (c *ReservationRPCClient) Enqueue(ctx context.Context, cmd usecasein.ReserveSeatCommand) (*usecasein.ReserveSeatResult, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	// Cola de respuesta: anónima (el broker le genera un nombre único),
	// exclusiva a esta conexión, y auto-delete al desconectar — no deja
	// basura acumulándose en RabbitMQ por cada request.
	replyQueue, err := c.channel.QueueDeclare("", false, true, true, false, nil)
	if err != nil {
		return nil, fmt.Errorf("error declarando cola de respuesta: %w", err)
	}

	msgs, err := c.channel.Consume(replyQueue.Name, "", true, true, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("error suscribiéndose a la cola de respuesta: %w", err)
	}

	correlationID := shared.NewID().String()
	payload, err := json.Marshal(rabbitmqdto.ReservationRequest{
		EventID: cmd.EventID.String(),
		SeatID:  cmd.SeatID.String(),
		UserID:  cmd.UserID.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("error serializando la petición: %w", err)
	}

	err = c.channel.PublishWithContext(ctx, "", rabbitmqdto.ReservationQueueName, false, false, amqp.Publishing{
		ContentType:   "application/json",
		CorrelationId: correlationID,
		ReplyTo:       replyQueue.Name,
		Body:          payload,
	})
	if err != nil {
		return nil, fmt.Errorf("error publicando en la sala de espera: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("tiempo de espera agotado en la sala de espera virtual: %w", ctx.Err())
		case d, ok := <-msgs:
			if !ok {
				return nil, fmt.Errorf("la cola de respuesta se cerró antes de recibir contestación")
			}
			if d.CorrelationId != correlationID {
				continue // respuesta de otra petición (no debería pasar con cola exclusiva, pero por seguridad)
			}
			var resp rabbitmqdto.ReservationResponse
			if err := json.Unmarshal(d.Body, &resp); err != nil {
				return nil, fmt.Errorf("error deserializando respuesta: %w", err)
			}
			if resp.Error != "" {
				return nil, shared.NewDomainError(resp.Error, resp.Error)
			}
			reservationID, err := shared.ParseID(resp.ReservationID)
			if err != nil {
				return nil, fmt.Errorf("reservation_id inválido en la respuesta: %w", err)
			}
			return &usecasein.ReserveSeatResult{ReservationID: reservationID, ExpiresInSec: resp.ExpiresIn}, nil
		}
	}
}
