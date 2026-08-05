<script setup lang="ts">
import type { Seat } from "~/types/ticketing";

const props = defineProps<{ seat: Seat; disabled: boolean; isMine?: boolean }>();
const emit = defineEmits<{ select: [seat: Seat]; focus: [] }>();

function handleClick() {
  if (props.isMine) {
    emit("focus"); // no se puede re-reservar tu propio asiento: llevamos al panel de pago
    return;
  }
  emit("select", props.seat);
}
</script>

<template>
  <button
    class="seat"
    :class="[props.seat.status.toLowerCase(), { mine: props.isMine }]"
    :disabled="!props.isMine && (props.disabled || props.seat.status !== 'AVAILABLE')"
    :title="props.isMine ? 'Este es tu asiento reservado — clic para ir al pago' : `Fila ${props.seat.row}, Asiento ${props.seat.label} — ${props.seat.status}`"
    @click="handleClick"
  >
    <span v-if="props.seat.status !== 'SOLD'">{{ props.seat.label }}</span>
  </button>
</template>

<style scoped>
.seat {
  width: 40px;
  height: 40px;
  border-radius: 4px;
  border: 2px solid var(--text-dark);
  font-family: var(--font-mono);
  font-weight: 700;
  font-size: 0.9rem;
  cursor: pointer;
  transition: var(--transition);
  color: var(--text-dark);
  display: flex;
  align-items: center;
  justify-content: center;
  background: transparent;
  box-shadow: 3px 3px 0px rgba(0, 0, 0, 0.3);
  position: relative;
}

.seat.available {
  background: var(--cyan);
  box-shadow: 3px 3px 0px var(--text-dark);
}
.seat.available:hover:not(:disabled) {
  transform: translate(-3px, -3px);
  box-shadow: 6px 6px 0px var(--text-dark), 0 0 20px var(--cyan);
  background: #ffffff;
  z-index: 2;
}
.seat.available:active:not(:disabled) {
  transform: translate(3px, 3px);
  box-shadow: 0px 0px 0px var(--text-dark);
}

.seat.reserved {
  background: var(--pink);
  color: #fff;
  box-shadow: 3px 3px 0px var(--text-dark), 0 0 20px var(--pink);
  transform: scale(1.05);
}

.seat.mine {
  outline: 3px solid var(--cyan);
  outline-offset: 3px;
  cursor: pointer;
}

.seat.sold {
  background: var(--grey);
  color: #666;
  cursor: not-allowed;
  box-shadow: inset 0 0 10px #000;
  background-image: repeating-linear-gradient(45deg, #3a3a4a 0px, #3a3a4a 4px, #2a2a3a 4px, #2a2a3a 8px);
  border-color: #222;
  transform: none !important;
}

@media (prefers-reduced-motion: reduce) {
  .seat {
    transition: none;
  }
}
</style>
