export interface TicketRecord {
  ticketId: string;
  seatLabel: string;
  purchasedAt: string;
}

// Guarda los boletos comprados en esta sesión en localStorage, por
// evento. Es intencionalmente simple (sin backend "mis compras" real
// todavía) — suficiente para que el usuario pueda volver a ver un
// ticket que ya confirmó sin tener que anotar el código a mano.
export function useTicketHistory(eventId: string) {
  const storageKey = `ticketing_history_${eventId}`;
  const history = useState<TicketRecord[]>(`ticket-history-${eventId}`, () => []);

  function load() {
    if (import.meta.server) return;
    try {
      const raw = localStorage.getItem(storageKey);
      history.value = raw ? (JSON.parse(raw) as TicketRecord[]) : [];
    } catch {
      history.value = [];
    }
  }

  function add(record: TicketRecord) {
    if (import.meta.server) return;
    history.value = [record, ...history.value];
    localStorage.setItem(storageKey, JSON.stringify(history.value));
  }

  function clear() {
    if (import.meta.server) return;
    history.value = [];
    localStorage.removeItem(storageKey);
  }

  return { history, load, add, clear };
}
