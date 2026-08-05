<script setup lang="ts">
import { nextTick, watch } from "vue";
import type { LogEntry } from "~/composables/useActivityLog";

const props = defineProps<{ entries: LogEntry[] }>();
const logEl = ref<HTMLElement | null>(null);

watch(
  () => props.entries.length,
  async () => {
    await nextTick();
    if (logEl.value) logEl.value.scrollTop = logEl.value.scrollHeight;
  }
);
</script>

<template>
  <div ref="logEl" class="terminal-log" role="log" aria-live="polite">
    <div v-if="!props.entries.length" class="log-entry">
      <span class="time">[SYS]</span> Esperando actividad...
    </div>
    <div v-for="e in props.entries" :key="e.id" class="log-entry">
      <span class="time">{{ e.timestamp }}</span>
      <span :class="e.type">{{ e.message }}</span>
    </div>
  </div>
</template>

<style scoped>
.terminal-log {
  background: #08080c;
  border: 2px solid #333;
  border-radius: var(--radius);
  padding: 0.8rem;
  height: 110px;
  overflow-y: auto;
  margin-top: 1.2rem;
  font-family: var(--font-mono);
  font-size: 0.75rem;
  color: var(--text-muted);
  display: flex;
  flex-direction: column;
  gap: 2px;
  box-shadow: inset 0 0 20px rgba(0, 0, 0, 0.8);
}
.terminal-log::-webkit-scrollbar {
  width: 4px;
}
.terminal-log::-webkit-scrollbar-thumb {
  background: var(--pink);
  border-radius: 2px;
}
.log-entry .time {
  color: #555;
  margin-right: 6px;
}
.log-entry .success {
  color: var(--cyan);
}
.log-entry .warning {
  color: var(--pink);
}
.log-entry .info {
  color: var(--yellow);
}
</style>
