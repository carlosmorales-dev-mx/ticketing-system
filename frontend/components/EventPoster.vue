<script setup lang="ts">
import type { Seat } from "~/types/ticketing";
import { MAX_SEATS_PER_USER } from "~/utils/seat";

const props = defineProps<{ eventId: string; seats: Seat[] }>();

const shortId = computed(() => props.eventId.slice(0, 8));

// Mientras no hay asientos cargados mostramos "–" en vez de ceros falsos.
const ready = computed(() => props.seats.length > 0);
const counts = computed(() => {
  let available = 0;
  let reserved = 0;
  let sold = 0;
  for (const s of props.seats) {
    if (s.status === "AVAILABLE") available++;
    else if (s.status === "RESERVED") reserved++;
    else sold++;
  }
  return { available, reserved, sold };
});
</script>

<template>
  <aside class="poster">
    <p class="event">Evento {{ shortId }}</p>
    <h1>Elige tus asientos</h1>
    <p class="lead">
      Toca hasta {{ MAX_SEATS_PER_USER }} lugares libres y los apartamos para ti mientras pagas. El plano cambia solo cuando alguien más aparta o compra.
    </p>
    <dl class="stats">
      <div class="free">
        <dt>Libres</dt>
        <dd>{{ ready ? counts.available : "–" }}</dd>
      </div>
      <div>
        <dt>Apartados</dt>
        <dd>{{ ready ? counts.reserved : "–" }}</dd>
      </div>
      <div>
        <dt>Vendidos</dt>
        <dd>{{ ready ? counts.sold : "–" }}</dd>
      </div>
    </dl>
  </aside>
</template>

<style scoped>
.poster {
  background: linear-gradient(175deg, #c8212f 0%, #a31a29 55%, var(--red-deep) 100%);
  color: var(--text);
  padding: 2rem 1.75rem 1.75rem;
  display: flex;
  flex-direction: column;
  gap: 1.4rem;
  border-right: 2px dashed rgba(244, 239, 224, 0.4);
}
.event {
  font-family: var(--font-code);
  font-size: 0.78rem;
  color: var(--gold-soft);
}
h1 {
  font-family: var(--font-display);
  font-weight: 800;
  text-transform: uppercase;
  font-size: clamp(3rem, 5vw, 4.3rem);
  line-height: 0.9;
  letter-spacing: 0.01em;
  text-wrap: balance;
}
.lead {
  color: rgba(244, 239, 224, 0.9);
  max-width: 30ch;
}
.stats {
  margin-top: auto;
  display: grid;
}
.stats > div {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 0.75rem;
  padding-block: 0.75rem;
  border-top: 1px solid rgba(244, 239, 224, 0.28);
}
dt {
  font-size: 0.9rem;
}
dd {
  font-family: var(--font-display);
  font-weight: 800;
  font-size: 2.3rem;
  line-height: 1;
  font-variant-numeric: tabular-nums;
}
.free dd {
  color: var(--gold-soft);
}

@media (max-width: 940px) {
  .poster {
    border-right: 0;
    border-bottom: 2px dashed rgba(244, 239, 224, 0.4);
    padding: 1.5rem 1.25rem;
    gap: 1rem;
  }
  .lead {
    max-width: 50ch;
  }
  .stats {
    margin-top: 0.25rem;
    grid-template-columns: repeat(3, 1fr);
    gap: 1rem;
  }
  .stats > div {
    flex-direction: column;
    align-items: flex-start;
    gap: 0.1rem;
  }
}
</style>
