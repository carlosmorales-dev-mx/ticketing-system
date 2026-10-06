<script setup lang="ts">
import type { Seat } from "~/types/ticketing";

const props = defineProps<{
  seat: Seat;
  disabled: boolean;
  isMine?: boolean;
  /** Desplazamiento vertical en px para curvar la fila hacia el escenario. */
  offset?: number;
}>();
const emit = defineEmits<{ select: [seat: Seat]; focus: [] }>();

const STATUS_TEXT: Record<Seat["status"], string> = {
  AVAILABLE: "libre",
  RESERVED: "apartado",
  SOLD: "vendido",
};

const description = computed(() => {
  const base = `Fila ${props.seat.row}, asiento ${props.seat.label}`;
  return props.isMine
    ? `${base}, tu lugar. Clic para ir al pago`
    : `${base}, ${STATUS_TEXT[props.seat.status]}`;
});

function handleClick() {
  if (props.isMine) {
    emit("focus"); // no se puede re-reservar tu propio asiento: llevamos al panel de pago
    return;
  }
  emit("select", props.seat);
}

// Un destello breve cuando el asiento cambia de estado (por ejemplo,
// cuando llega una actualización por WebSocket de otra persona).
const flashing = ref(false);
let flashTimer: ReturnType<typeof setTimeout> | null = null;
watch(
  () => props.seat.status,
  () => {
    flashing.value = true;
    if (flashTimer) clearTimeout(flashTimer);
    flashTimer = setTimeout(() => (flashing.value = false), 1000);
  }
);
onUnmounted(() => {
  if (flashTimer) clearTimeout(flashTimer);
});
</script>

<template>
  <button
    class="seat"
    :class="[props.seat.status.toLowerCase(), { mine: props.isMine, flash: flashing }]"
    :style="{ '--o': `${props.offset ?? 0}px` }"
    :disabled="!props.isMine && (props.disabled || props.seat.status !== 'AVAILABLE')"
    :title="description"
    :aria-label="description"
    type="button"
    @click="handleClick"
  >
    <span v-if="props.seat.status !== 'SOLD'" aria-hidden="true">{{ props.seat.label }}</span>
  </button>
</template>

<style scoped>
.seat {
  width: var(--seat, 34px);
  height: var(--seat, 34px);
  border-radius: 50%;
  border: 2px solid transparent;
  background: var(--seat-free);
  color: var(--bg);
  font-family: var(--font-body);
  font-size: 0.72rem;
  font-weight: 700;
  display: grid;
  place-items: center;
  padding: 0;
  cursor: pointer;
  transform: translateY(var(--o, 0px));
  transition: background-color 0.15s, transform 0.15s, box-shadow 0.15s;
}
.seat:hover:not(:disabled) {
  background: var(--gold);
  transform: translateY(calc(var(--o, 0px) - 2px)) scale(1.1);
}
.seat:disabled {
  cursor: default;
}

/* Apartado por otra persona */
.seat.reserved {
  background: transparent;
  border: 2px dashed rgba(233, 189, 60, 0.7);
  color: var(--gold-soft);
}
.seat.sold {
  background: var(--seat-sold);
  color: transparent;
}
/* Tu lugar */
.seat.mine {
  background: var(--red-hi);
  border-color: transparent;
  color: var(--text);
  cursor: pointer;
  box-shadow: 0 0 0 4px rgba(225, 58, 73, 0.3), 0 0 18px rgba(225, 58, 73, 0.75);
}

.seat.flash {
  animation: flash 1s ease-out;
}
@keyframes flash {
  0% {
    outline: 3px solid var(--gold-soft);
    outline-offset: 3px;
  }
  100% {
    outline: 3px solid transparent;
    outline-offset: 10px;
  }
}

@media (prefers-reduced-motion: reduce) {
  .seat {
    transition: none;
  }
  .seat.flash {
    animation: none;
  }
}
</style>
