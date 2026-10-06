<script setup lang="ts">
import qrcode from "qrcode-generator";
import { describeSeat } from "~/utils/seat";

const props = defineProps<{
  ticketId: string;
  eventId: string;
  seatLabel?: string;
  purchasedAt?: string;
}>();

// Código QR del boleto, dibujado como un solo <path> SVG (sin canvas ni
// imágenes externas). Si falla por cualquier motivo, el boleto sigue
// mostrando el código escrito.
const qr = computed<{ size: number; path: string } | null>(() => {
  try {
    const q = qrcode(0, "M");
    q.addData(props.ticketId);
    q.make();
    const size = q.getModuleCount();
    let path = "";
    for (let r = 0; r < size; r++) {
      for (let c = 0; c < size; c++) {
        if (q.isDark(r, c)) path += `M${c} ${r}h1v1h-1z`;
      }
    }
    return { size, path };
  } catch {
    return null;
  }
});

const shortCode = computed(() => props.ticketId.slice(0, 8).toUpperCase());
const seatText = computed(() => (props.seatLabel ? describeSeat(props.seatLabel) : null));
const purchased = computed(() => {
  if (!props.purchasedAt) return null;
  const d = new Date(props.purchasedAt);
  if (Number.isNaN(d.getTime())) return null;
  return {
    date: d.toLocaleDateString("es-MX", { day: "numeric", month: "long", year: "numeric" }),
    time: d.toLocaleTimeString("es-MX", { hour: "2-digit", minute: "2-digit", hour12: false }),
  };
});
</script>

<template>
  <article class="ticket">
    <div class="pass">
      <div class="pass-qr">
        <svg v-if="qr" :viewBox="`0 0 ${qr.size} ${qr.size}`" role="img" aria-label="Código QR del boleto" shape-rendering="crispEdges">
          <path :d="qr.path" fill="currentColor" />
        </svg>
        <div v-else class="qr-fallback">QR no disponible</div>
        <code>{{ shortCode }}</code>
      </div>

      <ul class="pass-info">
        <li v-if="seatText">
          <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M3 9a2 2 0 0 0 0 6v3h18v-3a2 2 0 0 1 0-6V6H3z" /></svg>
          <span>{{ seatText }}</span>
        </li>
        <li>
          <svg viewBox="0 0 24 24" aria-hidden="true"><circle cx="12" cy="8" r="4" /><path d="M4 21c0-4 4-6 8-6s8 2 8 6" /></svg>
          <span>1 boleto</span>
        </li>
        <template v-if="purchased">
          <li>
            <svg viewBox="0 0 24 24" aria-hidden="true"><rect x="3" y="5" width="18" height="16" rx="2" /><path d="M3 10h18M8 3v4M16 3v4" /></svg>
            <span>{{ purchased.date }}</span>
          </li>
          <li>
            <svg viewBox="0 0 24 24" aria-hidden="true"><circle cx="12" cy="12" r="9" /><path d="M12 7v5l3 2" /></svg>
            <span>{{ purchased.time }}</span>
          </li>
        </template>
        <li>
          <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M12 21s7-6.2 7-11a7 7 0 1 0-14 0c0 4.8 7 11 7 11z" /><circle cx="12" cy="10" r="2.5" /></svg>
          <span>Evento {{ props.eventId.slice(0, 8) }}</span>
        </li>
      </ul>
    </div>

    <p class="full">
      Muestra el QR o este código en la entrada
      <code>{{ props.ticketId }}</code>
    </p>
  </article>
</template>

<style scoped>
.ticket {
  display: grid;
  gap: 0.6rem;
  break-inside: avoid;
}

.pass {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr);
  position: relative;
  overflow: hidden;
  background: #fff;
  border-radius: 10px;
  box-shadow: 0 1px 0 var(--rule);
}
.pass-qr {
  display: grid;
  gap: 0.5rem;
  justify-items: center;
  align-content: center;
  padding: 1.1rem 1rem 0.9rem;
  border-right: 2px dashed var(--rule);
  color: var(--ink);
}
.pass-qr svg,
.qr-fallback {
  width: 104px;
  height: 104px;
  display: block;
}
.qr-fallback {
  display: grid;
  place-items: center;
  padding: 0.5rem;
  border: 2px dashed var(--rule);
  text-align: center;
  font-size: 0.72rem;
  color: var(--ink-muted);
}
.pass-qr code {
  font-family: var(--font-code);
  font-size: 0.78rem;
}

.pass-info {
  position: relative;
  display: grid;
  gap: 0.65rem;
  align-content: center;
  padding: 1.1rem 1rem;
  font-size: 0.9rem;
}
/* Muescas del talón */
.pass-info::before,
.pass-info::after {
  content: "";
  position: absolute;
  left: -10px;
  width: 18px;
  height: 18px;
  border-radius: 50%;
  background: var(--paper);
}
.pass-info::before {
  top: -9px;
}
.pass-info::after {
  bottom: -9px;
}
.pass-info li {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  min-width: 0;
}
.pass-info svg {
  width: 16px;
  height: 16px;
  flex: none;
  fill: none;
  stroke: var(--red);
  stroke-width: 1.8;
  stroke-linecap: round;
  stroke-linejoin: round;
}

.full {
  text-align: center;
  font-size: 0.75rem;
  color: var(--ink-muted);
}
.full code {
  display: block;
  margin-top: 2px;
  font-family: var(--font-code);
  word-break: break-all;
}

@media (max-width: 480px) {
  .pass {
    grid-template-columns: minmax(0, 1fr);
  }
  .pass-qr {
    border-right: 0;
    border-bottom: 2px dashed var(--rule);
  }
  .pass-info::before {
    top: -9px;
    left: -9px;
  }
  .pass-info::after {
    top: -9px;
    bottom: auto;
    left: auto;
    right: -9px;
  }
}
</style>
