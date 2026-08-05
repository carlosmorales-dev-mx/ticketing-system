<script setup lang="ts">
import type { Seat } from "~/types/ticketing";

const props = defineProps<{ seats: Seat[]; disabled: boolean; activeSeatLabel?: string | null }>();
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
</script>

<template>
  <section class="map-section" aria-label="Mapa de asientos">
    <div class="legend" role="list">
      <div class="legend-item" role="listitem"><span class="legend-color disp" /> Disponible</div>
      <div class="legend-item" role="listitem"><span class="legend-color resv" /> Reservado</div>
      <div class="legend-item" role="listitem"><span class="legend-color vend" /> Agotado</div>
    </div>

    <StageSVG />

    <div class="seat-grid" role="grid" aria-label="Asientos">
      <template v-for="r in rows" :key="r.row">
        <div class="row-label" role="rowheader">{{ r.row }}</div>
        <SeatCell
          v-for="seat in r.seats"
          :key="seat.seat_id"
          :seat="seat"
          :disabled="props.disabled"
          :is-mine="`${seat.row}${seat.label}` === props.activeSeatLabel"
          @select="emit('select', $event)"
          @focus="emit('focus-reservation')"
        />
      </template>
    </div>
  </section>
</template>

<style scoped>
.map-section {
  background-color: var(--yellow);
  border: var(--border-heavy);
  border-radius: var(--radius);
  padding: 1.8rem 1.5rem;
  flex: 1.2;
  min-width: 480px;
  box-shadow: var(--shadow-hard);
}

.legend {
  display: flex;
  justify-content: center;
  gap: 2rem;
  margin-bottom: 1.8rem;
  padding: 0.6rem 1rem;
  background: var(--bg-deep);
  border: var(--border-heavy);
  border-radius: var(--radius);
  box-shadow: 4px 4px 0px var(--text-dark);
  flex-wrap: wrap;
}
.legend-item {
  display: flex;
  align-items: center;
  gap: 10px;
  color: var(--text-main);
  font-weight: 700;
  font-size: 0.9rem;
  text-transform: uppercase;
  letter-spacing: 1px;
}
.legend-color {
  width: 18px;
  height: 18px;
  border: 2px solid var(--text-dark);
  border-radius: 4px;
  box-shadow: 2px 2px 0px var(--text-dark);
  display: inline-block;
}
.legend-color.disp { background: var(--cyan); }
.legend-color.resv { background: var(--pink); }
.legend-color.vend { background: var(--grey); }

.seat-grid {
  background: var(--bg-panel);
  border: var(--border-heavy);
  border-radius: var(--radius);
  padding: 1rem;
  display: grid;
  grid-template-columns: auto repeat(8, 1fr);
  gap: 10px;
  align-items: center;
  justify-items: center;
  overflow-x: auto;
  box-shadow: inset 0 0 20px rgba(1, 225, 238, 0.1);
}
.row-label {
  color: var(--cyan);
  font-family: var(--font-display);
  font-weight: 700;
  font-size: 1rem;
  text-align: right;
  padding-right: 0.8rem;
  text-shadow: 0 0 8px var(--cyan);
  justify-self: end;
}

@media (max-width: 900px) {
  .map-section {
    min-width: unset;
    width: 100%;
  }
  .seat-grid {
    gap: 6px;
    padding: 0.5rem;
  }
}
</style>
