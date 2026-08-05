import type { Seat, SeatStatus } from "~/types/ticketing";

// useState de Nuxt: estado reactivo compartido entre componentes,
// sin necesidad de Pinia para un caso tan simple como este. Se
// hidrata igual en SSR y cliente porque Nuxt lo serializa entre ambos.
export function useSeats() {
  const seats = useState<Seat[]>("seats", () => []);

  function setSeats(newSeats: Seat[]) {
    seats.value = newSeats;
  }

  // Aplica una actualización puntual (la que llega por WebSocket) sin
  // tener que volver a pedir la lista completa al backend.
  function applySeatUpdate(seatId: string, status: SeatStatus) {
    const seat = seats.value.find((s) => s.seat_id === seatId);
    if (seat) seat.status = status;
  }

  return { seats, setSeats, applySeatUpdate };
}
