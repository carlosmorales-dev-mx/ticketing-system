<script setup lang="ts">
import { ApiError } from "~/types/ticketing";
import type { Seat } from "~/types/ticketing";
import type { ActiveReservation } from "~/composables/useActiveReservations";
import type { TicketRecord } from "~/composables/useTicketHistory";
import { MAX_SEATS_PER_USER, describeSeat } from "~/utils/seat";

const route = useRoute();
const eventId = route.params.id as string;
const config = useRuntimeConfig();
const showReset = !!config.public.enableReset;

const api = useApi();
const { seats, setSeats } = useSeats();
const { history, load: loadHistory, add: addToHistory, clear: historyClear } = useTicketHistory(eventId);
const { save: saveActive, restore: restoreActive } = useActiveReservationsStorage(eventId);
const { entries: logEntries, log: addLog } = useActivityLog();
useSeatSocket(eventId); // actualiza `seats` en vivo por WebSocket

const userId = ref("");
const errorMessage = ref<string | null>(null);

// --- Carga del mapa, con reintento automático ---------------------------
const ready = ref(false); // ya se cargó el plano al menos una vez
const failed = ref(false); // la última carga falló
const retrying = ref(false); // hay una petición de carga en curso
const retryIn = ref(0); // segundos para el próximo reintento automático
const RETRY_DELAYS = [2, 3, 5, 5, 8, 8]; // luego deja de insistir solo
let attempt = 0;
let retryTimer: ReturnType<typeof setInterval> | null = null;

function clearRetry() {
  if (retryTimer) {
    clearInterval(retryTimer);
    retryTimer = null;
  }
  retryIn.value = 0;
}

function scheduleRetry() {
  if (attempt >= RETRY_DELAYS.length) return; // dejamos de insistir; queda el botón manual
  retryIn.value = RETRY_DELAYS[attempt++];
  retryTimer = setInterval(() => {
    retryIn.value -= 1;
    if (retryIn.value <= 0) {
      clearRetry();
      loadSeats();
    }
  }, 1000);
}

async function loadSeats(manual = false) {
  clearRetry();
  if (manual) attempt = 0;
  retrying.value = true;
  try {
    setSeats(await api.listSeats(eventId));
    if (failed.value) addLog("Conexión con el servidor recuperada", "success");
    ready.value = true;
    failed.value = false;
    attempt = 0;
  } catch {
    if (!ready.value) {
      failed.value = true;
      scheduleRetry();
    } else {
      errorMessage.value = "No se pudo actualizar el plano. Revisa tu conexión.";
    }
  } finally {
    retrying.value = false;
  }
}

const retryText = computed(() => {
  if (retrying.value) return "Conectando con el servidor…";
  if (retryIn.value > 0) return `Reintentamos solos en ${retryIn.value} s.`;
  return "Revisa que el servidor esté encendido y vuelve a intentarlo.";
});

// --- Reservas (hasta MAX_SEATS_PER_USER, cada una con su reloj) -----------
const reservations = ref<ActiveReservation[]>([]);
const reserving = ref(0); // peticiones de reserva en vuelo
const busy = ref(false); // pagando o liberando
const limitReached = computed(() => reservations.value.length + reserving.value >= MAX_SEATS_PER_USER);
const mineLabels = computed(() => reservations.value.map((r) => r.seatLabel));

watch(reservations, (list) => saveActive(list), { deep: true });

// Boletos que se están viendo en el modal (recién comprados, o uno del historial).
const modalTickets = ref<TicketRecord[]>([]);

onMounted(async () => {
  userId.value = useUserId();
  loadHistory();

  // Si había reservas vivas antes de recargar la página, las
  // recuperamos: cada contador sigue desde su tiempo real restante.
  const restored = restoreActive();
  if (restored.length) {
    reservations.value = restored;
    addLog(`Reservas recuperadas: ${restored.map((r) => r.seatLabel).join(", ")}`, "info");
  }

  await loadSeats();
  if (ready.value) addLog("Sistema listo", "success");
});
onUnmounted(clearRetry);

function removeReservation(id: string) {
  reservations.value = reservations.value.filter((r) => r.reservationId !== id);
}

async function onSelectSeat(seat: Seat) {
  errorMessage.value = null;
  if (limitReached.value) {
    errorMessage.value = `Ya tienes ${MAX_SEATS_PER_USER} asientos apartados. Paga o libera uno para elegir otro.`;
    return;
  }
  const seatLabel = `${seat.row}${seat.label}`;
  reserving.value++;
  try {
    const res = await api.reserveSeat(eventId, seat.seat_id, userId.value);
    reservations.value = [
      ...reservations.value,
      { reservationId: res.reservation_id, seatLabel, expiresAt: Date.now() + res.expires_in * 1000 },
    ];
    addLog(`Reservado ${seatLabel} (${reservations.value.length} de ${MAX_SEATS_PER_USER})`, "success");
  } catch (e) {
    if (e instanceof ApiError && e.code === "SEAT_ALREADY_RESERVED") {
      errorMessage.value = "Justo se adelantaron: ese asiento ya no está disponible.";
      addLog(`Conflicto: ${seatLabel} ya no disponible`, "warning");
    } else if (e instanceof ApiError && e.code === "MAX_SEATS_PER_USER") {
      errorMessage.value = `Solo puedes apartar hasta ${MAX_SEATS_PER_USER} asientos a la vez.`;
      addLog("Límite de asientos alcanzado", "warning");
    } else if (e instanceof ApiError && e.code === "RATE_LIMITED") {
      errorMessage.value = "Demasiadas peticiones, espera un momento.";
    } else {
      errorMessage.value = "No se pudo reservar el asiento.";
    }
  } finally {
    reserving.value--;
  }
}

// Paga todas las reservas, una por una. Si alguna ya había expirado, las
// demás se pagan igual y se avisa de cuáles no.
async function onConfirm() {
  if (!reservations.value.length || busy.value) return;
  busy.value = true;
  errorMessage.value = null;

  const bought: TicketRecord[] = [];
  const lost: string[] = [];

  for (const r of [...reservations.value]) {
    try {
      const res = await api.confirmPayment(r.reservationId);
      const record: TicketRecord = {
        ticketId: res.ticket_id,
        seatLabel: r.seatLabel,
        purchasedAt: new Date().toISOString(),
      };
      addToHistory(record);
      bought.push(record);
      addLog(`Compra confirmada: ${describeSeat(r.seatLabel)}, ticket ${res.ticket_id.slice(0, 8)}`, "success");
    } catch {
      lost.push(r.seatLabel);
      addLog(`No se pudo pagar ${r.seatLabel}`, "warning");
    }
    removeReservation(r.reservationId);
  }

  if (lost.length) {
    errorMessage.value =
      lost.length === 1
        ? `No se pudo pagar ${describeSeat(lost[0])}: la reserva pudo haber expirado.`
        : `No se pudieron pagar ${lost.length} asientos (${lost.join(", ")}): sus reservas pudieron haber expirado.`;
  }
  if (bought.length) modalTickets.value = bought;
  busy.value = false;
}

function onExpired(reservationId: string) {
  const r = reservations.value.find((x) => x.reservationId === reservationId);
  if (!r) return;
  removeReservation(reservationId);
  errorMessage.value = `Se acabó el tiempo: ${describeSeat(r.seatLabel)} se liberó.`;
  addLog(`Reserva expirada: ${r.seatLabel} liberado`, "warning");
}

async function cancelQuietly(reservationId: string) {
  try {
    await api.cancelReservation(reservationId);
  } catch {
    // Aunque falle la llamada (ej. ya había expirado un instante
    // antes), igual soltamos el estado local: el WebSocket ya
    // actualizará el color del asiento cuando corresponda.
  }
}

async function onCancelOne(reservationId: string) {
  const r = reservations.value.find((x) => x.reservationId === reservationId);
  if (!r || busy.value) return;
  busy.value = true;
  await cancelQuietly(reservationId);
  removeReservation(reservationId);
  addLog(`Liberaste ${r.seatLabel}`, "info");
  busy.value = false;
}

async function onCancelAll() {
  if (!reservations.value.length || busy.value) return;
  busy.value = true;
  const all = [...reservations.value];
  await Promise.all(all.map((r) => cancelQuietly(r.reservationId)));
  reservations.value = [];
  addLog(all.length > 1 ? `Liberaste ${all.length} asientos` : `Liberaste ${all[0].seatLabel}`, "info");
  busy.value = false;
}

function viewTicket(ticketId: string) {
  const record = history.value.find((t) => t.ticketId === ticketId);
  if (record) modalTickets.value = [record];
}

function onClearHistory() {
  historyClear();
  addLog("Historial de boletos vaciado", "info");
}

// Después de reiniciar el mapa no queda nada del lado del servidor:
// ni reservas ni boletos. Limpiamos también lo guardado en este navegador.
async function onMapReset(released: number) {
  reservations.value = [];
  modalTickets.value = [];
  historyClear();
  errorMessage.value = null;
  await loadSeats();
  addLog(`Mapa reiniciado: ${released} asiento${released === 1 ? "" : "s"} liberado${released === 1 ? "" : "s"}`, "info");
}

const reservationEl = ref<{ $el: HTMLElement } | null>(null);
function focusReservation() {
  reservationEl.value?.$el.scrollIntoView({ behavior: "smooth", block: "center" });
}
</script>

<template>
  <main class="page">
    <TicketModal
      v-if="modalTickets.length"
      :tickets="modalTickets"
      :event-id="eventId"
      @close="modalTickets = []"
    />

    <section class="card" aria-label="Mapa de asientos">
      <EventPoster :event-id="eventId" :seats="seats" />

      <div class="room">
        <p v-if="!ready && !failed" class="loading" role="status">Cargando asientos…</p>

        <div v-else-if="!ready" class="load-error" role="alert">
          <p class="le-title">No pudimos cargar el plano</p>
          <p class="le-text">{{ retryText }}</p>
          <button class="btn" type="button" :disabled="retrying" @click="loadSeats(true)">
            {{ retrying ? "Reintentando…" : "Reintentar ahora" }}
          </button>
        </div>

        <SeatMap
          v-else
          :seats="seats"
          :disabled="limitReached || busy"
          :mine-labels="mineLabels"
          @select="onSelectSeat"
          @focus-reservation="focusReservation"
        />

        <ReservationBar
          ref="reservationEl"
          :reservations="reservations"
          :error="errorMessage"
          :busy="busy"
          @confirm="onConfirm"
          @cancel-one="onCancelOne"
          @cancel-all="onCancelAll"
          @expired="onExpired"
        />

        <div v-if="showReset" class="room-foot">
          <ResetMap :event-id="eventId" :disabled="busy" @done="onMapReset" />
        </div>
      </div>
    </section>

    <div class="lower">
      <TicketHistory :tickets="history" @view="viewTicket" @clear="onClearHistory" />
      <ActivityLog :entries="logEntries" />
    </div>
  </main>
</template>

<style scoped>
.page {
  --poster-w: 340px;
  max-width: 1180px;
  margin: 0 auto;
  padding: 2rem 1.25rem 4rem;
  display: grid;
  gap: 1.5rem;
}

/* Una sola pieza tipo boleto: talón rojo + sala, con muescas en la perforación */
.card {
  position: relative;
  display: grid;
  grid-template-columns: var(--poster-w) minmax(0, 1fr);
  overflow: hidden;
  background: var(--panel);
  border-radius: 12px;
}
.card::before,
.card::after {
  content: "";
  position: absolute;
  left: calc(var(--poster-w) - 15px);
  width: 30px;
  height: 30px;
  border-radius: 50%;
  background: var(--bg);
  z-index: 2;
}
.card::before {
  top: -15px;
}
.card::after {
  bottom: -15px;
}

.room {
  min-width: 0;
  display: grid;
  gap: 1.4rem;
  align-content: start;
  padding: 1.75rem 1.75rem 1.5rem;
  background: radial-gradient(ellipse 70% 50% at 50% 0%, rgba(233, 189, 60, 0.09), transparent 70%), var(--panel);
}
.loading {
  text-align: center;
  color: var(--muted);
  padding-block: 4rem;
}

.load-error {
  display: grid;
  justify-items: center;
  gap: 0.6rem;
  padding-block: 3rem;
  text-align: center;
}
.le-title {
  font-family: var(--font-display);
  font-weight: 800;
  font-size: 1.9rem;
  line-height: 1.05;
}
.le-text {
  max-width: 34ch;
  margin-bottom: 0.4rem;
  color: var(--muted);
  font-size: 0.9rem;
}

/* Botón pequeño de reinicio, abajo a la derecha del plano */
.room-foot {
  margin-top: -0.6rem;
}

.lower {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 1.5rem;
}

@media (max-width: 940px) {
  .card {
    grid-template-columns: minmax(0, 1fr);
  }
  .card::before,
  .card::after {
    display: none;
  }
  .lower {
    grid-template-columns: minmax(0, 1fr);
  }
}
@media (max-width: 480px) {
  .page {
    padding-inline: 1rem;
  }
  .room {
    padding: 1.25rem 0.9rem;
  }
}
</style>
