<script setup lang="ts">
import type { Seat } from "~/types/ticketing";

const props = defineProps<{ seats: Seat[]; disabled: boolean; mineLabels?: string[] }>();
const emit = defineEmits<{ select: [seat: Seat]; "focus-reservation": [] }>();

const rows = computed(() => {
  const grouped = new Map<number, Seat[]>();
  for (const seat of props.seats) {
    if (!grouped.has(seat.row)) grouped.set(seat.row, []);
    grouped.get(seat.row)!.push(seat);
  }
  return [...grouped.entries()]
    .sort(([a], [b]) => a - b)
    .map(([row, seats]) => ({ row, seats: seats.sort((a, b) => a.label.localeCompare(b.label)) }));
});

const mine = computed(() => new Set(props.mineLabels ?? []));

// El pasillo queda en el centro de cada fila.
function aisleAt(count: number) {
  return Math.ceil(count / 2);
}

// Curva las filas hacia el escenario: los asientos de los extremos
// suben hasta 11px, los del centro se quedan en su lugar.
const MAX_LIFT = 11;
function lift(index: number, count: number) {
  const center = (count - 1) / 2;
  if (center === 0) return 0;
  const d = index - center;
  return -((d * d) / (center * center)) * MAX_LIFT;
}
</script>

<template>
  <section class="map" aria-label="Plano de la sala">
    <div class="plan-scroll">
      <div class="hall">
        <StageSVG />

        <div v-for="r in rows" :key="r.row" class="row" role="group" :aria-label="`Fila ${r.row}`">
          <span class="row-num" aria-hidden="true">{{ r.row }}</span>
          <template v-for="(seat, i) in r.seats" :key="seat.seat_id">
            <span v-if="i === aisleAt(r.seats.length)" class="aisle" aria-hidden="true" />
            <SeatCell
              :seat="seat"
              :disabled="props.disabled"
              :is-mine="mine.has(`${seat.row}${seat.label}`)"
              :offset="lift(i, r.seats.length)"
              @select="emit('select', $event)"
              @focus="emit('focus-reservation')"
            />
          </template>
          <span class="row-num" aria-hidden="true">{{ r.row }}</span>
        </div>
      </div>
    </div>

    <div class="legend" aria-hidden="true">
      <span><i class="dot free" /> Libre</span>
      <span><i class="dot mine" /> Tu lugar</span>
      <span><i class="dot held" /> Apartado</span>
      <span><i class="dot sold" /> Vendido</span>
    </div>
  </section>
</template>

<style scoped>
.map {
  --seat: 34px;
  --gap: 8px;
  display: grid;
  gap: 1.4rem;
  min-width: 0;
}
.plan-scroll {
  overflow-x: auto;
  padding-bottom: 4px;
}
.hall {
  width: max-content;
  margin: 0 auto;
  display: grid;
  gap: 8px;
  justify-items: center;
}

.row {
  display: flex;
  align-items: flex-end;
  gap: var(--gap);
  height: calc(var(--seat) + 12px);
}
.row-num {
  width: 20px;
  text-align: center;
  align-self: center;
  color: var(--muted);
  font-family: var(--font-display);
  font-weight: 700;
  font-size: 1rem;
}
.aisle {
  width: 22px;
}

.legend {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem 1.4rem;
  justify-content: center;
  color: var(--muted);
  font-size: 0.8rem;
}
.legend span {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}
.dot {
  display: inline-block;
  width: 12px;
  height: 12px;
  border-radius: 50%;
}
.dot.free {
  background: var(--seat-free);
}
.dot.mine {
  background: var(--red-hi);
}
.dot.held {
  border: 1.5px dashed rgba(233, 189, 60, 0.7);
}
.dot.sold {
  background: var(--seat-sold);
}

@media (max-width: 480px) {
  .map {
    --seat: 28px;
    --gap: 4px;
  }
  .row-num {
    width: 12px;
  }
  .aisle {
    width: 12px;
  }
}
</style>
