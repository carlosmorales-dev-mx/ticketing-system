import type {
  Seat,
  ReserveSeatResponse,
  ConfirmPaymentResponse,
  ResetEventResponse,
  ApiErrorResponse,
} from "~/types/ticketing";
import { ApiError } from "~/types/ticketing";

export function useApi() {
  const config = useRuntimeConfig();
  const base = config.public.apiBase;

  // Wrapper central: cualquier respuesta no-2xx se traduce a
  // ApiError con el código exacto que documenta el openapi.yaml
  // (SEAT_ALREADY_RESERVED, MAX_SEATS_PER_USER, RATE_LIMITED,
  // RESERVATION_NOT_PENDING...). Así los componentes hacen
  // `catch (e) { if (e.code === '...') }` en vez de parsear mensajes.
  async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
    const res = await fetch(`${base}${path}`, {
      ...options,
      headers: { "Content-Type": "application/json", ...options.headers },
    });

    if (!res.ok) {
      const body = (await res.json().catch(() => ({}))) as ApiErrorResponse;
      throw new ApiError(body.error || "UNKNOWN_ERROR", res.status);
    }

    return res.json() as Promise<T>;
  }

  return {
    listSeats(eventId: string) {
      return request<Seat[]>(`/events/${eventId}/seats`);
    },

    reserveSeat(eventId: string, seatId: string, userId: string) {
      return request<ReserveSeatResponse>("/reservations", {
        method: "POST",
        body: JSON.stringify({ event_id: eventId, seat_id: seatId, user_id: userId }),
      });
    },

    confirmPayment(reservationId: string) {
      return request<ConfirmPaymentResponse>(`/reservations/${reservationId}/confirm`, {
        method: "POST",
      });
    },

    // Herramienta de desarrollo: deja el mapa como nuevo. El backend
    // responde 403 MAP_RESET_DISABLED si no se arrancó con ENABLE_MAP_RESET=true.
    resetEvent(eventId: string) {
      return request<ResetEventResponse>(`/events/${eventId}/reset`, { method: "POST" });
    },

    // El backend responde 204 sin body, así que no usamos el helper
    // `request<T>` genérico (que siempre intenta parsear JSON).
    async cancelReservation(reservationId: string) {
      const res = await fetch(`${base}/reservations/${reservationId}/cancel`, { method: "POST" });
      if (!res.ok) {
        const body = (await res.json().catch(() => ({}))) as ApiErrorResponse;
        throw new ApiError(body.error || "UNKNOWN_ERROR", res.status);
      }
    },
  };
}
