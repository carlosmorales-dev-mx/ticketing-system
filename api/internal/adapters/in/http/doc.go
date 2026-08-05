// Package http expone la API REST. Traduce HTTP <-> casos de uso:
//   - handler/    convierte un http.Request en un Command del puerto
//     in, invoca el caso de uso, y traduce el resultado (o el
//     shared.DomainError) a una respuesta HTTP + código de estado.
//   - middleware/ rate limiting, logging, recovery, auth.
//   - dto/        structs de request/response con tags `json`. Estos
//     DTOs NUNCA se filtran hacia application ni domain — el handler
//     los mapea a Commands y de vuelta.
//
// Aquí es donde se traduce shared.DomainError.Code a status HTTP:
//   SEAT_ALREADY_RESERVED -> 409
//   INVALID_ID / campos inválidos -> 400
//   (rate limiter) -> 429 con cabecera Retry-After
package http
