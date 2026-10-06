<script setup lang="ts">
import { ApiError } from "~/types/ticketing";

const props = defineProps<{ eventId: string; disabled?: boolean }>();
const emit = defineEmits<{ done: [released: number] }>();

const api = useApi();
const open = ref(false);
const busy = ref(false);
const message = ref<string | null>(null);
const cancelBtn = ref<HTMLButtonElement | null>(null);
const toggleBtn = ref<HTMLButtonElement | null>(null);

async function openPanel() {
  message.value = null;
  open.value = true;
  await nextTick();
  cancelBtn.value?.focus();
}
function closePanel() {
  open.value = false;
  toggleBtn.value?.focus();
}

async function run() {
  busy.value = true;
  message.value = null;
  try {
    const res = await api.resetEvent(props.eventId);
    open.value = false;
    emit("done", res.released_seats);
  } catch (e) {
    message.value =
      e instanceof ApiError && e.code === "MAP_RESET_DISABLED"
        ? "El servidor tiene el reinicio apagado (ENABLE_MAP_RESET)."
        : "No se pudo reiniciar el mapa. Intenta de nuevo.";
  } finally {
    busy.value = false;
  }
}

function onKeydown(e: KeyboardEvent) {
  if (e.key === "Escape" && open.value) closePanel();
}
onMounted(() => window.addEventListener("keydown", onKeydown));
onUnmounted(() => window.removeEventListener("keydown", onKeydown));
</script>

<template>
  <div class="reset">
    <div v-if="open" class="panel" role="dialog" aria-labelledby="reset-title">
      <p id="reset-title" class="panel-title">¿Reiniciar el mapa?</p>
      <p class="panel-text">
        Libera todos los asientos y borra las reservas y los boletos de este evento, para todas las personas conectadas.
      </p>
      <p v-if="message" class="panel-error" role="alert">{{ message }}</p>
      <div class="panel-actions">
        <button ref="cancelBtn" class="btn btn--ghost btn--small" type="button" :disabled="busy" @click="closePanel">
          Cancelar
        </button>
        <button class="btn btn--small" type="button" :disabled="busy" @click="run">
          {{ busy ? "Reiniciando…" : "Sí, reiniciar" }}
        </button>
      </div>
    </div>

    <button
      ref="toggleBtn"
      class="trigger"
      type="button"
      title="Reiniciar mapa (herramienta de desarrollo)"
      :aria-expanded="open"
      :disabled="props.disabled"
      @click="open ? closePanel() : openPanel()"
    >
      <svg viewBox="0 0 24 24" aria-hidden="true">
        <path d="M20 12a8 8 0 1 1-2.6-5.9" />
        <path d="M20 4v5h-5" />
      </svg>
      Reiniciar mapa
    </button>
  </div>
</template>

<style scoped>
.reset {
  position: relative;
  display: flex;
  justify-content: flex-end;
}
.trigger {
  appearance: none;
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.25rem 0.55rem;
  border: 0;
  border-radius: 999px;
  background: transparent;
  color: var(--muted);
  font-family: var(--font-body);
  font-size: 0.75rem;
  cursor: pointer;
  opacity: 0.75;
  transition: opacity 0.15s, color 0.15s;
}
.trigger svg {
  width: 13px;
  height: 13px;
  fill: none;
  stroke: currentColor;
  stroke-width: 2;
  stroke-linecap: round;
  stroke-linejoin: round;
}
.trigger:hover:not(:disabled),
.trigger[aria-expanded="true"] {
  opacity: 1;
  color: var(--text);
}
.trigger:disabled {
  cursor: not-allowed;
  opacity: 0.35;
}

.panel {
  position: absolute;
  right: 0;
  bottom: calc(100% + 8px);
  z-index: 5;
  width: min(300px, calc(100vw - 2rem));
  display: grid;
  gap: 0.6rem;
  padding: 0.9rem 1rem;
  border: 1px solid var(--line);
  border-radius: 10px;
  background: var(--panel-2);
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.45);
}
.panel-title {
  font-family: var(--font-display);
  font-weight: 800;
  font-size: 1.25rem;
  line-height: 1.1;
}
.panel-text {
  color: var(--muted);
  font-size: 0.82rem;
  line-height: 1.4;
}
.panel-error {
  padding: 0.4rem 0.6rem;
  border-left: 3px solid var(--red-hi);
  background: rgba(200, 34, 47, 0.16);
  font-size: 0.8rem;
}
.panel-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
}
</style>
