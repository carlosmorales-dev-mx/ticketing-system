<script setup lang="ts">
import { computed } from "vue";
import type { TicketRecord } from "~/composables/useTicketHistory";
import { describeSeat } from "~/utils/seat";

const props = defineProps<{ tickets: TicketRecord[] }>();
const emit = defineEmits<{ view: [ticketId: string]; clear: [] }>();

const title = computed(() => (props.tickets.length ? `Mis boletos (${props.tickets.length})` : "Mis boletos"));
</script>

<template>
  <section class="panel" aria-labelledby="tickets-title">
    <div class="head">
      <h2 id="tickets-title">{{ title }}</h2>
      <button v-if="props.tickets.length" class="link-btn" type="button" @click="emit('clear')">
        Borrar historial
      </button>
    </div>

    <ul class="list">
      <li v-if="!props.tickets.length" class="empty">Todavía no compras ningún boleto. Aquí aparecerán.</li>
      <li v-for="t in props.tickets" :key="t.ticketId" class="item">
        <div class="item-info">
          <span class="seat">{{ describeSeat(t.seatLabel) }}</span>
          <code>{{ t.ticketId.slice(0, 8) }}</code>
        </div>
        <button class="btn btn--ghost btn--small" type="button" @click="emit('view', t.ticketId)">
          Ver boleto
        </button>
      </li>
    </ul>
  </section>
</template>

<style scoped>
.panel {
  background: var(--panel);
  border-radius: 12px;
  padding: 1.4rem 1.5rem;
  min-width: 0;
}
.head {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  gap: 0.75rem;
}
h2 {
  font-size: 0.95rem;
  font-weight: 600;
  color: var(--muted);
}
.list {
  margin-top: 0.5rem;
  display: grid;
  max-height: 280px;
  overflow-y: auto;
}
.item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  padding: 0.75rem 0;
  border-top: 1px solid var(--line);
}
.item:first-child {
  border-top: 0;
}
.seat {
  font-weight: 600;
}
code {
  display: block;
  font-family: var(--font-code);
  font-size: 0.75rem;
  color: var(--muted);
}
.empty {
  color: var(--muted);
  font-size: 0.9rem;
  padding-top: 0.25rem;
}
</style>
