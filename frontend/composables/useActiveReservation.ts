interface StoredReservation {
  reservationId: string;
  seatLabel: string;
  expiresAt: number; // epoch ms — guardamos el momento absoluto de
  // expiración, no los segundos restantes, para poder recalcular
  // correctamente aunque pase tiempo entre que se guarda y se lee.
}

export interface RestoredReservation {
  reservationId: string;
  seatLabel: string;
  expiresInSec: number;
}

// Persiste la reserva activa en localStorage. Sin esto, recargar la
// página (F5, cambiar de pestaña y volver, etc.) hace que el frontend
// pierda el reservationId de una reserva que sigue viva en el
// servidor — el usuario queda viendo un asiento "suyo" reservado sin
// forma de pagarlo ni cancelarlo hasta que expire solo.
export function useActiveReservationStorage(eventId: string) {
  const storageKey = `ticketing_active_reservation_${eventId}`;

  function save(reservationId: string, seatLabel: string, expiresInSec: number) {
    if (import.meta.server) return;
    const record: StoredReservation = {
      reservationId,
      seatLabel,
      expiresAt: Date.now() + expiresInSec * 1000,
    };
    localStorage.setItem(storageKey, JSON.stringify(record));
  }

  function clear() {
    if (import.meta.server) return;
    localStorage.removeItem(storageKey);
  }

  function restore(): RestoredReservation | null {
    if (import.meta.server) return null;
    const raw = localStorage.getItem(storageKey);
    if (!raw) return null;

    try {
      const record = JSON.parse(raw) as StoredReservation;
      const remainingMs = record.expiresAt - Date.now();
      if (remainingMs <= 0) {
        localStorage.removeItem(storageKey);
        return null;
      }
      return {
        reservationId: record.reservationId,
        seatLabel: record.seatLabel,
        expiresInSec: Math.ceil(remainingMs / 1000),
      };
    } catch {
      localStorage.removeItem(storageKey);
      return null;
    }
  }

  return { save, clear, restore };
}
