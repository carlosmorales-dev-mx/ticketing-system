package rabbitmq

import (
	"context"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/carlosmorales-dev-mx/ticketing-system/api/internal/application/port/out"
)

const exchangeName = "ticketing.events"

type publisher struct {
	channel *amqp.Channel
}

// NewConnection abre la conexión AMQP. Se hace en el composition root
// (main.go) para poder cerrarla ordenadamente en el shutdown.
func NewConnection(url string) (*amqp.Connection, error) {
	conn, err := amqp.DialConfig(url, amqp.Config{
		// Sin esto, un RabbitMQ caído/lento puede colgar el arranque
		// del servicio indefinidamente.
		Dial: amqp.DefaultDial(5 * time.Second),
	})
	if err != nil {
		return nil, fmt.Errorf("error conectando a rabbitmq: %w", err)
	}
	return conn, nil
}

// NewEventPublisher declara un exchange topic durable y devuelve el
// adaptador. Un exchange durable sobrevive a reinicios de RabbitMQ;
// sin esto, perderías el exchange (y con él, las rutas de mensajes)
// en cualquier reinicio del broker.
func NewEventPublisher(conn *amqp.Connection) (out.EventPublisher, *amqp.Channel, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, nil, fmt.Errorf("error abriendo canal amqp: %w", err)
	}

	err = ch.ExchangeDeclare(
		exchangeName,
		"topic",
		true,  // durable
		false, // auto-deleted
		false, // internal
		false, // no-wait
		nil,
	)
	if err != nil {
		ch.Close()
		return nil, nil, fmt.Errorf("error declarando exchange: %w", err)
	}

	return &publisher{channel: ch}, ch, nil
}

func (p *publisher) Publish(ctx context.Context, topic string, payload []byte) error {
	err := p.channel.PublishWithContext(ctx,
		exchangeName,
		topic, // routing key, ej: "seat.reserved"
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         payload,
			DeliveryMode: amqp.Persistent, // sobrevive a reinicios del broker
			Timestamp:    time.Now(),
		},
	)
	if err != nil {
		return fmt.Errorf("error publicando evento %q: %w", topic, err)
	}
	return nil
}
