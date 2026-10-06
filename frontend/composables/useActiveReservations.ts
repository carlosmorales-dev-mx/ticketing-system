import { MAX_SEATS_PER_USER } from "~/utils/seat";

export interface ActiveReservation {
  reservationId: string;
  seatLabel: string;
  // Momento absoluto de expiración (epoch ms), no segundos restantes:
  // así el contador sigue bien aunque pase tiempo entre guardar y leer.
  expiresAt: number;
}

// Persiste las reservas activas (hasta MAX_SEATS_PER_USER, cada una con su
// propio reloj) en localStorage. Sin esto, recargar la página haría perder
// los reservationId de reservas que siguen vivas en el servidor.
export function useActiveReservationsStorage(eventId: string) {
  const key = `ticketing_active_reservations_${eventId}`;
  const legacyKey = `ticketing_active_reservation_${eventId}`; // versión de una sola reserva

  function save(list: ActiveReservation[]) {
    if (import.meta.server) return;
    try {
      if (list.length) localStorage.setItem(key, JSON.stringify(list));
      else localStorage.removeItem(key);
    } catch {
      /* sin almacenamiento disponible: la app sigue funcionando sin persistir */
    }
  }

  function restore(): ActiveReservation[] {
    if (import.meta.server) return [];
    const found: ActiveReservation[] = [];
    try {
      const raw = localStorage.getItem(key);
      if (raw) {
        const parsed: unknown = JSON.parse(raw);
        if (Array.isArray(parsed)) found.push(...(parsed as ActiveReservation[]));
      }

      // Migración: la versión anterior guardaba una sola reserva.
      const legacyRaw = localStorage.getItem(legacyKey);
      if (legacyRaw) {
        const old = JSON.parse(legacyRaw) as Partial<ActiveReservation>;
        if (old.reservationId && old.seatLabel && typeof old.expiresAt === "number") {
          found.push(old as ActiveReservation);
        }
        localStorage.removeItem(legacyKey);
      }
    } catch {
      /* JSON corrupto: se ignora y se parte de cero */
    }

    const now = Date.now();
    const seen = new Set<string>();
    const valid = found.filter((r) => {
      const ok =
        !!r &&
        typeof r.reservationId === "string" &&
        typeof r.seatLabel === "string" &&
        typeof r.expiresAt === "number" &&
        r.expiresAt > now &&
        !seen.has(r.reservationId);
      if (ok) seen.add(r.reservationId);
      return ok;
    });
    return valid.slice(0, MAX_SEATS_PER_USER);
  }

  return { save, restore };
}
