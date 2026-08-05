export type SeatStatus = "AVAILABLE" | "RESERVED" | "SOLD";

export interface Seat {
  seat_id: string;
  row: number;
  label: string;
  status: SeatStatus;
}

export interface ReserveSeatResponse {
  reservation_id: string;
  expires_in: number;
}

export interface ConfirmPaymentResponse {
  ticket_id: string;
}

export interface ApiErrorResponse {
  error: string;
}

// Payload que emite el backend por WebSocket cuando un asiento cambia
// de estado (ver adapters/in/websocket/hub.go en el backend).
export interface SeatUpdateMessage {
  event_id: string;
  seat_id: string;
  status: SeatStatus;
}

// Error tipado que lanzan los composables de la API: envuelve el
// código de error del backend (SEAT_ALREADY_RESERVED, etc.) para que
// los componentes puedan reaccionar sin parsear strings.
export class ApiError extends Error {
  constructor(
    public code: string,
    public status: number,
  ) {
    super(code);
    this.name = "ApiError";
  }
}
