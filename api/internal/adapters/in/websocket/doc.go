// Package websocket implementa RealtimeNotifier (application/port/out)
// vía un hub de conexiones WebSocket. Cuando un asiento cambia de
// estado, el hub retransmite el cambio a todos los clientes suscritos
// al evento correspondiente, para que el frontend Nuxt pinte el mapa
// de asientos en tiempo real sin hacer polling.
package websocket
