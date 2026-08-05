<script setup lang="ts">
const props = defineProps<{ ticketId: string }>();
const emit = defineEmits<{ close: [] }>();

const copied = ref(false);

function copyCode() {
  navigator.clipboard.writeText(props.ticketId).then(() => {
    copied.value = true;
    setTimeout(() => (copied.value = false), 1800);
  });
}
function saveAsPdf() {
  window.print();
}
function onKeydown(e: KeyboardEvent) {
  if (e.key === "Escape") emit("close");
}
onMounted(() => window.addEventListener("keydown", onKeydown));
onUnmounted(() => window.removeEventListener("keydown", onKeydown));
</script>

<template>
  <Teleport to="body">
    <div class="modal-overlay active" role="dialog" aria-modal="true" @click.self="emit('close')">
      <div class="modal-box ticket-print-area">
        <h2>✔ CONFIRMADO</h2>
        <p>Tu boleto está listo</p>
        <p class="ticket-code">{{ ticketId }}</p>
        <p class="footnote no-print">Guarda este código — es tu pase de entrada</p>

        <div class="modal-actions no-print">
          <button class="btn-confirm" @click="saveAsPdf">GUARDAR / IMPRIMIR</button>
          <button class="btn-confirm alt" @click="copyCode">{{ copied ? "¡COPIADO!" : "COPIAR CÓDIGO" }}</button>
        </div>
        <button class="btn-cancel no-print" @click="emit('close')">VOLVER A LOS ASIENTOS</button>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.7);
  backdrop-filter: blur(4px);
  z-index: 10000;
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 20px;
}

.modal-box {
  background: var(--bg-panel);
  border: var(--border-heavy);
  border-color: var(--cyan);
  border-radius: var(--radius);
  padding: 2.5rem 2.5rem 2rem;
  max-width: 460px;
  width: 100%;
  box-shadow: 8px 8px 0px var(--text-dark);
  text-align: center;
}
.modal-box h2 {
  font-family: var(--font-display);
  color: var(--cyan);
  margin: 0 0 0.5rem;
  font-size: 2rem;
  text-shadow: 0 0 20px var(--cyan);
}
.modal-box p {
  font-size: 1rem;
  color: #ccc;
  margin: 0.3rem 0;
}
.ticket-code {
  font-family: var(--font-mono);
  font-size: 1.3rem;
  color: var(--yellow);
  font-weight: 700;
  margin: 1rem 0;
  text-shadow: var(--yellow-glow);
  word-break: break-all;
}
.footnote {
  font-size: 0.8rem;
  color: #888;
}

.modal-actions {
  display: flex;
  gap: 0.8rem;
  justify-content: center;
  margin-top: 1.6rem;
  flex-wrap: wrap;
}
.modal-actions button {
  padding: 0.65rem 1.3rem;
  border: var(--border-heavy);
  border-radius: var(--radius);
  font-family: var(--font-display);
  font-weight: 700;
  text-transform: uppercase;
  cursor: pointer;
  transition: var(--transition);
  font-size: 0.75rem;
  letter-spacing: 1px;
}
.btn-confirm {
  background: var(--cyan);
  color: var(--text-dark);
  box-shadow: 4px 4px 0px var(--text-dark);
}
.btn-confirm.alt {
  background: var(--yellow);
}
.btn-confirm:hover {
  background: #fff;
  transform: translate(-2px, -2px);
  box-shadow: 6px 6px 0px var(--text-dark);
}

.btn-cancel {
  display: block;
  margin: 1.4rem auto 0;
  background: transparent;
  color: var(--pink);
  border: none;
  text-decoration: underline;
  font-family: var(--font-body);
  font-weight: 700;
  font-size: 0.8rem;
  cursor: pointer;
}
.btn-cancel:hover {
  color: #fff;
}

@media print {
  .no-print {
    display: none !important;
  }
}
</style>
