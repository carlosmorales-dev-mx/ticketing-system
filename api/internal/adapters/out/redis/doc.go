// Package redis implementa ReservationCache (application/port/out).
// SetWithTTL escribe una clave `reservation:{id}` con EX=600s.
// SubscribeExpirations usa PSUBSCRIBE a `__keyevent@0__:expired`
// (requiere notify-keyspace-events Ex, ya activado en el
// podman-compose) para enterarse de expiraciones en tiempo real.
package redis
