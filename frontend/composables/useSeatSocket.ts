import type { SeatUpdateMessage } from "~/types/ticketing";

// Se conecta a GET /ws/events/{eventId} (ver adapters/in/websocket/hub.go
// en el backend) y aplica cada actualización al estado compartido de
// useSeats en tiempo real — así el mapa de asientos cambia de color
// solo, sin que el usuario recargue la página.
export function useSeatSocket(eventId: string) {
  const config = useRuntimeConfig();
  const { applySeatUpdate } = useSeats();

  let socket: WebSocket | null = null;
  let reconnectTimer: ReturnType<typeof setTimeout> | null = null;

  function connect() {
    if (import.meta.server) return;

    socket = new WebSocket(`${config.public.wsBase}/ws/events/${eventId}`);

    socket.onmessage = (event) => {
      const msg: SeatUpdateMessage = JSON.parse(event.data);
      applySeatUpdate(msg.seat_id, msg.status);
    };

    // Reconexión simple con backoff fijo: si el backend se reinicia
    // (ej. durante desarrollo con hot-reload), el frontend no se
    // queda con una conexión muerta indefinidamente.
    socket.onclose = () => {
      reconnectTimer = setTimeout(connect, 2000);
    };

    socket.onerror = () => {
      socket?.close();
    };
  }

  function disconnect() {
    if (reconnectTimer) clearTimeout(reconnectTimer);
    socket?.close();
    socket = null;
  }

  onMounted(connect);
  onUnmounted(disconnect);
}
