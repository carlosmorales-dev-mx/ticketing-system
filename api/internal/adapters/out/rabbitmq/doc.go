// Package rabbitmq implementa EventPublisher (application/port/out):
// publica eventos de dominio como "seat.reserved" o
// "reservation.confirmed" a un exchange topic, y también expone la
// cola de "sala de espera virtual" que consume in/amqpconsumer.
package rabbitmq
