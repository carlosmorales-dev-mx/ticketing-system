<script setup lang="ts">
import { nextTick, watch } from "vue";
import type { LogEntry } from "~/composables/useActivityLog";

const props = defineProps<{ entries: LogEntry[] }>();
const logEl = ref<HTMLElement | null>(null);

// El composable guarda la hora como "[21:23:23]"; aquí se muestra sin corchetes.
function clock(entry: LogEntry) {
  return entry.timestamp.replace(/[[\]]/g, "");
}

watch(
  () => props.entries.length,
  async () => {
    await nextTick();
    if (logEl.value) logEl.value.scrollTop = logEl.value.scrollHeight;
  }
);
</script>

<template>
  <section class="panel" aria-labelledby="activity-title">
    <h2 id="activity-title">Actividad</h2>
    <ul ref="logEl" class="log" role="log" aria-live="polite">
      <li v-if="!props.entries.length" class="empty">Esperando actividad…</li>
      <li v-for="e in props.entries" :key="e.id">
        <time>{{ clock(e) }}</time>
        <span :class="e.type">{{ e.message }}</span>
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
h2 {
  font-size: 0.95rem;
  font-weight: 600;
  color: var(--muted);
}
.log {
  margin-top: 0.6rem;
  display: grid;
  gap: 0.4rem;
  max-height: 168px;
  overflow-y: auto;
  font-size: 0.85rem;
}
.log li {
  display: grid;
  grid-template-columns: 62px 1fr;
  gap: 0.5rem;
}
time {
  color: var(--muted);
  font-variant-numeric: tabular-nums;
}
.success {
  font-weight: 600;
}
.warning {
  color: var(--red-hi);
}
.info {
  color: var(--text);
}
.empty {
  display: block;
  color: var(--muted);
}
</style>
