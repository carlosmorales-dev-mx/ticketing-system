<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from "vue";
import { describeSeat, MAX_SEATS_PER_USER } from "~/utils/seat";
import type { ActiveReservation } from "~/composables/useActiveReservations";

const props = defineProps<{
  reservations: ActiveReservation[];
  error?: string | null;
  /** Mientras se paga o se libera algo, los botones se bloquean. */
  busy?: boolean;
}>();

const emit = defineEmits<{
  confirm: [];
  "cancel-one": [reservationId: string];
  "cancel-all": [];
  expired: [reservationId: string];
}>();

// Un solo reloj para todas las reservas: cada asiento tiene su propia
// hora de vencimiento (expiresAt) y aquí solo se calcula cuánto falta.
const now = ref(Date.now());
let timer: ReturnType<typeof setInterval> | null = null;

function secondsLeft(r: ActiveReservation) {
  return Math.ceil((r.expiresAt - now.value) / 1000);
}

function tick() {
  now.value = Date.now();
  for (const r of props.reservations) {
    if (secondsLeft(r) <= 0) emit("expired", r.reservationId);
  }
}

onMounted(() => {
  timer = setInterval(tick, 1000);
});
onUnmounted(() => {
  if (timer) clearInterval(timer);
});

function label(r: ActiveReservation) {
  const left = Math.max(secondsLeft(r), 0);
  return `${Math.floor(left / 60)}:${String(left % 60).padStart(2, "0")}`;
}

const items = computed(() =>
  props.reservations.map((r) => ({
    r,
    time: label(r),
    urgent: secondsLeft(r) <= 60,
  }))
);

const count = computed(() => props.reservations.length);
const payText = computed(() => (count.value > 1 ? `Pagar ${count.value} asientos` : "Pagar ahora"));
const releaseAllText = computed(() => (count.value > 1 ? "Liberar todos" : "Liberar asiento"));
</script>

<template>
  <div class="reserve">
    <p v-if="props.error" class="alert" role="alert">{{ props.error }}</p>

    <div class="reserve-row">
      <div class="info">
        <template v-if="count">
          <p class="seat-title">
            Tus asientos
            <span class="of">{{ count }} de {{ MAX_SEATS_PER_USER }}</span>
          </p>
          <ul class="chips">
            <li v-for="it in items" :key="it.r.reservationId" class="chip" :class="{ urgent: it.urgent }">
              <span class="chip-seat">{{ describeSeat(it.r.seatLabel) }}</span>
              <b class="chip-time" :aria-label="`Quedan ${it.time} para pagar`">{{ it.time }}</b>
              <button
                class="chip-x"
                type="button"
                :disabled="props.busy"
                :aria-label="`Liberar ${describeSeat(it.r.seatLabel)}`"
                @click="emit('cancel-one', it.r.reservationId)"
              >
                <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M6 6l12 12M18 6L6 18" /></svg>
              </button>
            </li>
          </ul>
          <small>Cada asiento tiene su propio reloj y se libera solo al llegar a 0:00.</small>
        </template>
        <template v-else>
          <p class="seat-none">Sin asientos apartados</p>
          <small>Puedes elegir hasta {{ MAX_SEATS_PER_USER }}. Cada uno se guarda 10 minutos mientras pagas.</small>
        </template>
      </div>

      <div class="actions">
        <button class="btn" type="button" :disabled="!count || props.busy" @click="emit('confirm')">
          {{ props.busy ? "Procesando…" : payText }}
        </button>
        <button v-if="count" class="btn btn--ghost" type="button" :disabled="props.busy" @click="emit('cancel-all')">
          {{ releaseAllText }}
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.reserve {
  border-top: 1px solid var(--line);
  padding-top: 1.25rem;
  display: grid;
  gap: 0.9rem;
}
.reserve-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 1rem 1.5rem;
}
.info {
  min-width: 0;
  flex: 1 1 260px;
  display: grid;
  gap: 0.55rem;
}
small {
  color: var(--muted);
  font-size: 0.85rem;
}
.seat-title {
  display: flex;
  align-items: baseline;
  gap: 0.7rem;
  font-family: var(--font-display);
  font-weight: 800;
  font-size: 1.8rem;
  line-height: 1.05;
}
.of {
  font-family: var(--font-body);
  font-weight: 500;
  font-size: 0.85rem;
  color: var(--muted);
}
.seat-none {
  font-family: var(--font-display);
  font-weight: 700;
  font-size: 1.6rem;
  line-height: 1.1;
  color: var(--muted);
}

.chips {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}
.chip {
  display: inline-flex;
  align-items: center;
  gap: 0.6rem;
  padding: 0.3rem 0.35rem 0.3rem 0.8rem;
  border: 1px solid var(--line);
  border-radius: 999px;
  background: var(--panel-2);
  font-size: 0.9rem;
}
.chip-time {
  font-family: var(--font-code);
  font-weight: 500;
  font-size: 0.85rem;
  color: var(--gold);
  font-variant-numeric: tabular-nums;
}
.chip.urgent {
  border-color: rgba(228, 64, 77, 0.6);
}
.chip.urgent .chip-time {
  color: var(--red-hi);
}
.chip-x {
  appearance: none;
  display: grid;
  place-items: center;
  width: 24px;
  height: 24px;
  padding: 0;
  border: 0;
  border-radius: 50%;
  background: transparent;
  color: var(--muted);
  cursor: pointer;
}
.chip-x svg {
  width: 12px;
  height: 12px;
  fill: none;
  stroke: currentColor;
  stroke-width: 2.4;
  stroke-linecap: round;
}
.chip-x:hover:not(:disabled) {
  background: rgba(255, 255, 255, 0.1);
  color: var(--text);
}
.chip-x:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.6rem;
}
.alert {
  padding: 0.6rem 0.9rem;
  border-left: 3px solid var(--red-hi);
  border-radius: 0 6px 6px 0;
  background: rgba(200, 34, 47, 0.16);
  font-size: 0.88rem;
}

@media (max-width: 480px) {
  .actions,
  .actions .btn {
    width: 100%;
  }
}
</style>
