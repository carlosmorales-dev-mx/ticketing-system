// Package amqpconsumer consume mensajes de RabbitMQ (la "sala de
// espera virtual"): peticiones de reserva encoladas bajo alta
// demanda, procesadas a un ritmo controlado hacia
// application/port/in.ReserveSeatUseCase, en lugar de dejar que miles
// de requests concurrentes golpeen Postgres directamente.
package amqpconsumer
