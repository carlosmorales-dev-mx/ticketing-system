// Package domain contiene el NÚCLEO del negocio: entidades, value objects
// y reglas de negocio puras.
//
// REGLAS DE ORO de esta capa:
//   - CERO dependencias externas. Nada de "database/sql", "net/http",
//     drivers de Postgres/Redis/RabbitMQ, ni siquiera un logger.
//   - No conoce JSON, no conoce HTTP, no conoce SQL.
//   - Solo Go estándar + tipos propios.
//
// Aquí viven las invariantes de negocio, por ejemplo:
//   - "Un asiento no puede reservarse si ya está VENDIDO o RESERVADO"
//   - "Una reserva expira a los 10 minutos de creada"
//   - "Un ticket solo puede emitirse tras un pago confirmado"
//
// Si mañana cambias Postgres por MongoDB, o REST por gRPC, este paquete
// no debería cambiar ni una línea. Esa es la prueba de fuego de que la
// arquitectura hexagonal está bien aplicada.
package domain
