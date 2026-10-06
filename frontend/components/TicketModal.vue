<script setup lang="ts">
import type { TicketRecord } from "~/composables/useTicketHistory";

const props = defineProps<{
  tickets: TicketRecord[];
  eventId: string;
}>();
const emit = defineEmits<{ close: [] }>();

const many = computed(() => props.tickets.length > 1);
const copied = ref(false);
const codesEl = ref<HTMLElement | null>(null);
const primaryBtn = ref<HTMLButtonElement | null>(null);

function selectCodes() {
  if (!codesEl.value) return;
  const range = document.createRange();
  range.selectNodeContents(codesEl.value);
  const sel = window.getSelection();
  sel?.removeAllRanges();
  sel?.addRange(range);
}
function copyCodes() {
  const text = props.tickets.map((t) => t.ticketId).join("\n");
  navigator.clipboard
    .writeText(text)
    .then(() => {
      copied.value = true;
      setTimeout(() => (copied.value = false), 1800);
    })
    .catch(selectCodes); // sin permiso de portapapeles: dejamos los códigos seleccionados
}
function saveAsPdf() {
  window.print();
}
function onKeydown(e: KeyboardEvent) {
  if (e.key === "Escape") emit("close");
}

let previousFocus: HTMLElement | null = null;
onMounted(() => {
  previousFocus = document.activeElement as HTMLElement | null;
  window.addEventListener("keydown", onKeydown);
  primaryBtn.value?.focus();
});
onUnmounted(() => {
  window.removeEventListener("keydown", onKeydown);
  previousFocus?.focus?.();
});
</script>

<template>
  <Teleport to="body">
    <div class="overlay" role="dialog" aria-modal="true" aria-labelledby="ticket-title" @click.self="emit('close')">
      <div class="sheet ticket-print-area">
        <h2 id="ticket-title">{{ many ? "¡Listo, ya tienes tus lugares!" : "¡Listo, ya tienes tu lugar!" }}</h2>

        <div ref="codesEl" class="passes">
          <TicketPass
            v-for="t in props.tickets"
            :key="t.ticketId"
            :ticket-id="t.ticketId"
            :event-id="props.eventId"
            :seat-label="t.seatLabel"
            :purchased-at="t.purchasedAt"
          />
        </div>

        <div class="actions no-print">
          <button ref="primaryBtn" class="btn" type="button" @click="saveAsPdf">
            {{ many ? "Guardar o imprimir todos" : "Guardar o imprimir" }}
          </button>
          <button class="btn btn--ghost" type="button" @click="copyCodes">
            {{ copied ? (many ? "Códigos copiados" : "Código copiado") : many ? "Copiar códigos" : "Copiar código" }}
          </button>
          <button class="back" type="button" @click="emit('close')">Volver al plano</button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
.overlay {
  position: fixed;
  inset: 0;
  z-index: 10000;
  display: grid;
  place-items: center;
  padding: 1rem;
  overflow-y: auto;
  background: rgba(5, 8, 18, 0.78);
}

.sheet {
  width: min(420px, 100%);
  display: grid;
  gap: 1.25rem;
  padding: 1.75rem 1.25rem 1.4rem;
  background: var(--paper);
  color: var(--ink);
  border-radius: 18px;
}
h2 {
  font-family: var(--font-display);
  font-weight: 800;
  font-size: 2.2rem;
  line-height: 1;
  text-align: center;
  text-wrap: balance;
}

.passes {
  display: grid;
  gap: 1.1rem;
}

.actions {
  display: grid;
  gap: 0.6rem;
}
.actions .btn {
  width: 100%;
}
.sheet .btn--ghost {
  color: var(--ink);
  border-color: var(--rule);
}
.sheet .btn--ghost:hover:not(:disabled) {
  border-color: var(--ink);
}
.back {
  appearance: none;
  justify-self: center;
  background: none;
  border: 0;
  padding: 0.25rem;
  color: var(--ink-muted);
  font-family: var(--font-body);
  font-size: 0.85rem;
  text-decoration: underline;
  text-underline-offset: 3px;
  cursor: pointer;
}
.back:hover {
  color: var(--ink);
}
.sheet :focus-visible {
  outline-color: var(--red);
}

@media print {
  .no-print {
    display: none !important;
  }
}
</style>
