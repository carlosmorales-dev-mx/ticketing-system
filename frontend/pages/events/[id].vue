<script setup lang="ts">
import { ApiError } from "~/types/ticketing";
import type { Seat } from "~/types/ticketing";

const route = useRoute();
const eventId = route.params.id as string;

const api = useApi();
const { seats, setSeats } = useSeats();
const { history, load: loadHistory, add: addToHistory, clear: historyClear } = useTicketHistory(eventId);
const { save: saveActiveReservation, clear: clearActiveReservation, restore: restoreActiveReservation } = useActiveReservationStorage(eventId);
const { entries: logEntries, log: addLog } = useActivityLog();
useSeatSocket(eventId); // actualiza `seats` en vivo por WebSocket

const userId = ref("");
const loading = ref(true);
const errorMessage = ref<string | null>(null);

const activeReservation = ref<{ reservationId: string; expiresInSec: number; seatLabel: string } | null>(null);
const confirmedTicketId = ref<string | null>(null);

onMounted(async () => {
  userId.value = useUserId();
  loadHistory();

  // Si había una reserva viva antes de recargar la página, la
  // recuperamos aquí — el countdown sigue desde el tiempo real
  // restante, no desde cero.
  const restored = restoreActiveReservation();
  if (restored) {
    activeReservation.value = restored;
    addLog(`Reserva recuperada: asiento ${restored.seatLabel}`, "info");
  }

  await loadSeats();
  addLog("Sistema listo", "success");
});

async function loadSeats() {
  try {
    loading.value = true;
    const data = await api.listSeats(eventId);
    setSeats(data);
  } catch {
    errorMessage.value = "No se pudo cargar el mapa de asientos.";
  } finally {
    loading.value = false;
  }
}

async function onSelectSeat(seat: Seat) {
  errorMessage.value = null;
  try {
    const res = await api.reserveSeat(eventId, seat.seat_id, userId.value);
    const seatLabel = `${seat.row}${seat.label}`;
    activeReservation.value = {
      reservationId: res.reservation_id,
      expiresInSec: res.expires_in,
      seatLabel,
    };
    saveActiveReservation(res.reservation_id, seatLabel, res.expires_in);
    addLog(`Reservado ${seatLabel}`, "success");
  } catch (e) {
    if (e instanceof ApiError && e.code === "SEAT_ALREADY_RESERVED") {
      errorMessage.value = "Justo se adelantaron — ese asiento ya no está disponible.";
      addLog(`Conflicto: ${seat.row}${seat.label} ya no disponible`, "warning");
    } else if (e instanceof ApiError && e.code === "RATE_LIMITED") {
      errorMessage.value = "Demasiadas peticiones, espera un momento.";
    } else {
      errorMessage.value = "No se pudo reservar el asiento.";
    }
  }
}

async function onConfirm() {
  if (!activeReservation.value) return;
  try {
    const res = await api.confirmPayment(activeReservation.value.reservationId);
    confirmedTicketId.value = res.ticket_id;
    addToHistory({
      ticketId: res.ticket_id,
      seatLabel: activeReservation.value.seatLabel,
      purchasedAt: new Date().toISOString(),
    });
    activeReservation.value = null;
    clearActiveReservation();
    addLog(`Compra confirmada — ticket ${res.ticket_id.slice(0, 8)}...`, "success");
  } catch {
    errorMessage.value = "No se pudo confirmar el pago. La reserva pudo haber expirado.";
    activeReservation.value = null;
    clearActiveReservation();
    addLog("Fallo al confirmar el pago", "warning");
  }
}

function onExpired() {
  errorMessage.value = "Se acabó el tiempo — la reserva expiró y el asiento se liberó.";
  activeReservation.value = null;
  clearActiveReservation();
  addLog("Reserva expirada, asiento liberado", "warning");
}

async function onCancel() {
  if (!activeReservation.value) return;
  try {
    await api.cancelReservation(activeReservation.value.reservationId);
  } catch {
    // Aunque falle la llamada (ej. ya había expirado un instante
    // antes), igual soltamos el estado local: el WebSocket ya
    // actualizará el color del asiento cuando corresponda.
    addLog("Reserva cancelada por el usuario", "info");
  } finally {
    activeReservation.value = null;
    clearActiveReservation();
  }
}

function viewTicket(ticketId: string) {
  confirmedTicketId.value = ticketId;
}

function onClearHistory() {
  historyClear();
  addLog("Historial de boletos vaciado", "info");
}

const ticketHistoryEl = ref<{ $el: HTMLElement } | null>(null);
function focusReservation() {
  ticketHistoryEl.value?.$el.scrollIntoView({ behavior: "smooth", block: "center" });
}
</script>

<template>
  <main class="page">
    <header class="header">
      <p class="kicker">✦ Sistema de venta de entradas ✦</p>
      <h1>Mapa de asientos</h1>
    </header>

    <p v-if="errorMessage" class="alert">{{ errorMessage }}</p>
    <TicketModal v-if="confirmedTicketId" :ticket-id="confirmedTicketId" @close="confirmedTicketId = null" />

    <p v-if="loading" class="loading">Cargando asientos...</p>

    <div v-else class="dashboard">
      <SeatMap
        :seats="seats"
        :disabled="!!activeReservation"
        :active-seat-label="activeReservation?.seatLabel ?? null"
        @select="onSelectSeat"
        @focus-reservation="focusReservation"
      />

      <TicketHistory
        ref="ticketHistoryEl"
        :tickets="history"
        :active-reservation="activeReservation"
        @view="viewTicket"
        @clear="onClearHistory"
        @confirm="onConfirm"
        @cancel="onCancel"
        @expired="onExpired"
      >
        <TerminalLog :entries="logEntries" />
      </TicketHistory>
    </div>
  </main>
</template>

<style scoped>
.page {
  max-width: 1300px;
  margin: 0 auto;
  padding: 2.5rem 1.5rem 4rem;
}

.header {
  text-align: center;
  margin-bottom: 2.5rem;
}
.kicker {
  font-family: var(--font-body);
  font-weight: 500;
  font-size: 1.1rem;
  letter-spacing: 4px;
  text-transform: uppercase;
  color: var(--pink);
  text-shadow: var(--pink-glow);
  margin: 0 0 0.2rem;
}
h1 {
  font-family: var(--font-display);
  font-size: 3.6rem;
  font-weight: 900;
  text-transform: uppercase;
  color: var(--cyan);
  -webkit-text-stroke: 2px var(--text-dark);
  text-shadow: 4px 4px 0 var(--pink), 6px 6px 0 var(--text-dark);
  letter-spacing: 2px;
  margin: 0;
  display: inline-block;
  transition: var(--transition);
}
h1:hover {
  transform: scale(1.02);
  text-shadow: 4px 4px 0 var(--yellow), 8px 8px 0 var(--text-dark);
}

.dashboard {
  display: flex;
  gap: 2rem;
  align-items: stretch;
  justify-content: center;
  flex-wrap: wrap;
}

@media (max-width: 900px) {
  .dashboard {
    flex-direction: column;
    align-items: center;
  }
  h1 {
    font-size: 2.4rem;
  }
}

.loading {
  text-align: center;
  font-family: var(--font-mono);
  font-size: 1.2rem;
  color: var(--text-muted);
}

.alert {
  background: var(--pink);
  border: var(--border-heavy);
  box-shadow: var(--shadow-hard);
  color: #fff;
  padding: 10px 16px;
  border-radius: var(--radius);
  font-family: var(--font-body);
  font-weight: 700;
  font-size: 0.95rem;
  max-width: 520px;
  margin: 0 auto 20px;
  text-align: center;
}
</style>
