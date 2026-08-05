<script setup lang="ts">
import { computed, onUnmounted, ref, watch } from "vue";
import type { TicketRecord } from "~/composables/useTicketHistory";

interface ActiveReservationView {
  reservationId: string;
  seatLabel: string;
  expiresInSec: number;
}

const props = defineProps<{
  tickets: TicketRecord[];
  activeReservation: ActiveReservationView | null;
}>();

const emit = defineEmits<{
  view: [ticketId: string];
  clear: [];
  confirm: [];
  cancel: [];
  expired: [];
}>();

// El countdown vive aquí, dentro del propio panel "MIS BOLETOS" — no
// como una tarjeta flotante aparte. Se reinicia cada vez que llega una
// reserva nueva (comparando reservationId) y se detiene si desaparece.
const remaining = ref(0);
let timer: ReturnType<typeof setInterval> | null = null;

function stopTimer() {
  if (timer) {
    clearInterval(timer);
    timer = null;
  }
}

watch(
  () => props.activeReservation?.reservationId,
  () => {
    stopTimer();
    if (props.activeReservation) {
      remaining.value = props.activeReservation.expiresInSec;
      timer = setInterval(() => {
        remaining.value -= 1;
        if (remaining.value <= 0) {
          stopTimer();
          emit("expired");
        }
      }, 1000);
    }
  },
  { immediate: true }
);
onUnmounted(stopTimer);

const timeLabel = computed(() => {
  const m = Math.floor(remaining.value / 60);
  const s = String(remaining.value % 60).padStart(2, "0");
  return `${m}:${s}`;
});
const urgent = computed(() => remaining.value <= 60);
</script>

<template>
  <aside class="inventory-aside" aria-label="Mis boletos">
    <div class="inventory-header">
      <h2 class="inventory-title">MIS BOLETOS</h2>
      <div class="header-right">
        <div class="inventory-badge" aria-live="polite">{{ props.tickets.length }}</div>
        <button v-if="props.tickets.length" class="btn-clear" type="button" @click="emit('clear')">
          VACIAR
        </button>
      </div>
    </div>

    <ul class="ticket-list" role="list">
      <li v-if="!props.tickets.length" class="empty">Ninguno comprado todavía</li>
      <li v-for="t in props.tickets" :key="t.ticketId" class="ticket-item" role="listitem">
        <div class="ticket-info">Asiento <span>{{ t.seatLabel }}</span></div>
        <button class="btn-action" type="button" @click="emit('view', t.ticketId)">VER</button>
      </li>
    </ul>

    <slot />

    <!-- Checkout: igual que en el HTML original, la última sección
         dentro del mismo panel, no una tarjeta aparte — pero con su
         propio marco y resplandor para que destaque como la acción
         principal del panel. -->
    <div
      class="checkout-wrapper"
      :class="{
        'checkout-wrapper--active': props.activeReservation,
        'checkout-wrapper--urgent': props.activeReservation && urgent,
      }"
    >
      <div class="total-display" :class="{ urgent: props.activeReservation && urgent }">
        <span>
          <span v-if="props.activeReservation" class="live-dot" aria-hidden="true" />
          {{ props.activeReservation ? `ASIENTO ${props.activeReservation.seatLabel}` : "SIN RESERVA" }}
        </span>
        <span class="total-amount">{{ props.activeReservation ? timeLabel : "--:--" }}</span>
      </div>
      <button class="btn-checkout" type="button" :disabled="!props.activeReservation" @click="emit('confirm')">
        PAGAR AHORA
      </button>
      <button class="btn-reset" type="button" :disabled="!props.activeReservation" @click="emit('cancel')">
        CANCELAR
      </button>
    </div>
  </aside>
</template>

<style scoped>
.inventory-aside {
  background: var(--bg-panel);
  border: var(--border-heavy);
  border-color: var(--pink);
  border-radius: var(--radius);
  padding: 1.8rem 1.5rem;
  flex: 0.8;
  min-width: 300px;
  box-shadow: 6px 6px 0px var(--pink);
  display: flex;
  flex-direction: column;
}

.inventory-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
  padding-bottom: 0.8rem;
  border-bottom: 3px dashed var(--pink);
}
.inventory-title {
  color: var(--pink);
  font-family: var(--font-display);
  font-size: 1.2rem;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  text-shadow: var(--pink-glow);
  margin: 0;
}
.header-right {
  display: flex;
  align-items: center;
  gap: 0.6rem;
}
.inventory-badge {
  background: var(--pink);
  color: #fff;
  padding: 4px 12px;
  border-radius: 20px;
  font-weight: 700;
  font-size: 0.9rem;
  border: 2px solid var(--text-dark);
  box-shadow: 2px 2px 0px var(--text-dark);
}
.btn-clear {
  background: transparent;
  border: none;
  color: var(--text-muted);
  font-family: var(--font-body);
  font-weight: 700;
  font-size: 0.7rem;
  letter-spacing: 0.05em;
  text-decoration: underline;
  cursor: pointer;
}
.btn-clear:hover {
  color: var(--pink);
}

.ticket-list {
  list-style: none;
  margin: 0;
  padding: 0 4px 0 0;
  flex-grow: 1;
  min-height: 140px;
  display: flex;
  flex-direction: column;
  gap: 0.6rem;
  max-height: 280px;
  overflow-y: auto;
}
.ticket-list::-webkit-scrollbar {
  width: 4px;
}
.ticket-list::-webkit-scrollbar-thumb {
  background: var(--pink);
  border-radius: 2px;
}
.empty {
  color: #666;
  text-align: center;
  padding: 1.5rem 0;
  font-style: italic;
}
.ticket-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.7rem 1rem;
  background: rgba(252, 44, 147, 0.1);
  border-left: 4px solid var(--pink);
  border: 1px solid transparent;
  border-left: 4px solid var(--pink);
  border-radius: 4px;
  transition: var(--transition);
}
.ticket-item:hover {
  background: rgba(252, 44, 147, 0.2);
  transform: translateX(6px);
  border-color: var(--pink);
  box-shadow: 0 0 15px rgba(252, 44, 147, 0.2);
}
.ticket-info {
  color: #fff;
  font-size: 0.95rem;
  font-weight: 500;
}
.ticket-info span {
  color: var(--cyan);
  font-weight: 700;
  text-shadow: 0 0 8px var(--cyan);
}
.btn-action {
  background: transparent;
  border: 2px solid var(--pink);
  color: var(--pink);
  padding: 0.2rem 0.8rem;
  font-family: var(--font-body);
  font-weight: 700;
  font-size: 0.8rem;
  text-transform: uppercase;
  border-radius: 4px;
  cursor: pointer;
  transition: var(--transition);
  box-shadow: 2px 2px 0px var(--text-dark);
}
.btn-action:hover {
  background: var(--pink);
  color: #fff;
  box-shadow: 4px 4px 0px var(--text-dark);
  transform: translate(-1px, -1px);
}
.btn-action:active {
  transform: translate(2px, 2px);
  box-shadow: 0px 0px 0px var(--text-dark);
}

/* ===== Checkout: mismo contenido que tu .checkout-wrapper original,
   pero con marco propio + resplandor para que destaque como la
   acción principal del panel. ===== */
.checkout-wrapper {
  margin-top: 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 0.9rem;
  padding: 1.1rem 1.1rem 1.3rem;
  border: 2px solid var(--grey);
  border-radius: var(--radius);
  transition: border-color 0.3s ease;
}
.checkout-wrapper--active {
  background: rgba(1, 225, 238, 0.06);
  border-color: var(--cyan);
  animation: checkout-breathe 2.4s ease-in-out infinite;
}
.checkout-wrapper--urgent {
  background: rgba(252, 44, 147, 0.08);
  border-color: var(--pink) !important;
  animation: checkout-breathe-urgent 0.9s ease-in-out infinite;
}
@keyframes checkout-breathe {
  0%, 100% { box-shadow: 0 0 14px rgba(1, 225, 238, 0.25); }
  50% { box-shadow: 0 0 28px rgba(1, 225, 238, 0.55); }
}
@keyframes checkout-breathe-urgent {
  0%, 100% { box-shadow: 0 0 14px rgba(252, 44, 147, 0.35); }
  50% { box-shadow: 0 0 30px rgba(252, 44, 147, 0.7); }
}
@media (prefers-reduced-motion: reduce) {
  .checkout-wrapper--active,
  .checkout-wrapper--urgent {
    animation: none;
  }
}

.live-dot {
  display: inline-block;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--cyan);
  box-shadow: 0 0 8px var(--cyan);
  margin-right: 6px;
  vertical-align: middle;
}
.total-display.urgent .live-dot {
  background: var(--pink);
  box-shadow: 0 0 8px var(--pink);
}

.total-display {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 1.3rem;
  font-weight: 700;
  color: #fff;
  font-family: var(--font-display);
  letter-spacing: 1.5px;
  text-shadow: 0 0 10px rgba(255, 255, 255, 0.2);
}
.total-amount {
  font-size: 1.6rem;
  color: var(--yellow);
  text-shadow: var(--yellow-glow);
  font-variant-numeric: tabular-nums;
}
.total-display.urgent .total-amount {
  color: var(--pink);
  text-shadow: var(--pink-glow);
}

.btn-checkout,
.btn-reset {
  width: 100%;
  border: var(--border-heavy);
  padding: 0.9rem 1rem;
  font-family: var(--font-display);
  font-size: 1rem;
  text-transform: uppercase;
  font-weight: 900;
  cursor: pointer;
  border-radius: var(--radius);
  transition: var(--transition);
  letter-spacing: 2px;
}
.btn-checkout {
  background: var(--cyan);
  color: var(--text-dark);
  box-shadow: 4px 4px 0px var(--text-dark);
}
.btn-checkout:hover:not(:disabled) {
  transform: translate(-3px, -3px);
  box-shadow: 8px 8px 0px var(--text-dark), 0 0 30px var(--cyan);
  background: #ffffff;
}
.btn-checkout:active:not(:disabled) {
  transform: translate(3px, 3px);
  box-shadow: 0px 0px 0px var(--text-dark);
}
.btn-checkout:disabled {
  background: var(--grey);
  color: #555;
  border-color: #222;
  cursor: not-allowed;
  box-shadow: none;
}

.btn-reset {
  background: transparent;
  color: var(--pink);
  border-color: var(--pink);
  box-shadow: 3px 3px 0px var(--text-dark);
  font-size: 0.8rem;
  padding: 0.6rem;
}
.btn-reset:hover:not(:disabled) {
  background: var(--pink);
  color: #fff;
  box-shadow: 5px 5px 0px var(--text-dark);
  transform: translate(-2px, -2px);
}
.btn-reset:disabled {
  opacity: 0.4;
  cursor: not-allowed;
  box-shadow: none;
}
</style>
