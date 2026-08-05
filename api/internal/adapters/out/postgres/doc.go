// Package postgres implementa SeatRepository, ReservationRepository y
// TicketRepository (application/port/out) usando database/sql +
// pgx/lib-pq.
//
// SeatRepository.TryReserve es el método más importante de todo el
// proyecto: debe ejecutar algo equivalente a
//
//	UPDATE seats SET status = 'RESERVED'
//	WHERE id = $1 AND status = 'AVAILABLE'
//
// y comprobar RowsAffected(). Si es 0, alguien ganó la carrera antes.
// Esto es más eficiente que un SELECT ... FOR UPDATE explícito y
// evita deadlocks bajo alta concurrencia, aunque ambas estrategias son
// válidas y documentables como decisión de diseño.
package postgres
