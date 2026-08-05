package amqpconsumer

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/carlosmorales-dev-mx/ticketing-system/api/internal/adapters/rabbitmqdto"
	usecasein "github.com/carlosmorales-dev-mx/ticketing-system/api/internal/application/port/in"
	"github.com/carlosmorales-dev-mx/ticketing-system/api/internal/domain/shared"
)

// StartWorkers levanta `numWorkers` consumidores independientes de la
// cola de reservas. Cada worker abre su propio canal AMQP y pide
// Qos(prefetch=1): RabbitMQ solo le entrega un mensaje nuevo cuando
// terminó (ack) el anterior. Eso es lo que limita, de verdad, cuántas
// reservas tocan Postgres al mismo tiempo — el número de workers ES
// el ancho de la "puerta" de la sala de espera hacia la base de datos.
func StartWorkers(ctx context.Context, conn *amqp.Connection, reserveSeat usecasein.ReserveSeatUseCase, numWorkers int) error {
	if numWorkers < 1 {
		numWorkers = 1
	}

	for i := 0; i < numWorkers; i++ {
		ch, err := conn.Channel()
		if err != nil {
			return err
		}
		if err := ch.Qos(1, 0, false); err != nil {
			return err
		}
		if _, err := ch.QueueDeclare(rabbitmqdto.ReservationQueueName, true, false, false, false, nil); err != nil {
			return err
		}
		msgs, err := ch.Consume(rabbitmqdto.ReservationQueueName, "", false, false, false, false, nil)
		if err != nil {
			return err
		}

		workerID := i
		go runWorker(ctx, workerID, ch, msgs, reserveSeat)
	}

	log.Printf("sala de espera virtual activa: %d worker(s) escuchando %q", numWorkers, rabbitmqdto.ReservationQueueName)
	return nil
}

func runWorker(ctx context.Context, id int, ch *amqp.Channel, msgs <-chan amqp.Delivery, reserveSeat usecasein.ReserveSeatUseCase) {
	defer ch.Close()
	for {
		select {
		case <-ctx.Done():
			return
		case d, ok := <-msgs:
			if !ok {
				return
			}
			handle(ctx, ch, d, reserveSeat)
		}
	}
}

func handle(ctx context.Context, ch *amqp.Channel, d amqp.Delivery, reserveSeat usecasein.ReserveSeatUseCase) {
	var req rabbitmqdto.ReservationRequest
	if err := json.Unmarshal(d.Body, &req); err != nil {
		log.Printf("mensaje de reserva corrupto, se descarta: %v", err)
		_ = d.Nack(false, false) // no reintenta: un mensaje corrupto nunca se va a arreglar solo
		return
	}

	resp := process(ctx, req, reserveSeat)

	body, err := json.Marshal(resp)
	if err != nil {
		log.Printf("error serializando respuesta de reserva: %v", err)
		_ = d.Nack(false, false)
		return
	}

	if d.ReplyTo != "" {
		err = ch.PublishWithContext(ctx, "", d.ReplyTo, false, false, amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: d.CorrelationId,
			Body:          body,
		})
		if err != nil {
			log.Printf("error publicando respuesta al cliente RPC: %v", err)
		}
	}

	_ = d.Ack(false)
}

func process(ctx context.Context, req rabbitmqdto.ReservationRequest, reserveSeat usecasein.ReserveSeatUseCase) rabbitmqdto.ReservationResponse {
	eventID, err1 := shared.ParseID(req.EventID)
	seatID, err2 := shared.ParseID(req.SeatID)
	userID, err3 := shared.ParseID(req.UserID)
	if err1 != nil || err2 != nil || err3 != nil {
		return rabbitmqdto.ReservationResponse{Error: "INVALID_ID"}
	}

	execCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := reserveSeat.Execute(execCtx, usecasein.ReserveSeatCommand{
		EventID: eventID,
		SeatID:  seatID,
		UserID:  userID,
	})
	if err != nil {
		var domainErr *shared.DomainError
		if errors.As(err, &domainErr) {
			return rabbitmqdto.ReservationResponse{Error: domainErr.Code}
		}
		log.Printf("error interno procesando reserva en worker: %v", err)
		return rabbitmqdto.ReservationResponse{Error: "INTERNAL_ERROR"}
	}

	return rabbitmqdto.ReservationResponse{
		ReservationID: result.ReservationID.String(),
		ExpiresIn:     result.ExpiresInSec,
	}
}
