<!--
  Escenario: una línea punteada dorada curva y su etiqueta. Es solo
  ambientación; vive separado del SeatMap para no mezclarla con los
  asientos reales.
-->
<template>
  <div class="stage" aria-hidden="true">
    <svg viewBox="0 0 520 40" preserveAspectRatio="none">
      <path d="M12 34 Q260 -14 508 34" />
    </svg>
    <span>Escenario</span>
  </div>
</template>

<style scoped>
.stage {
  /* width:0 + min-width:100%: el arco se estira al ancho de las filas pero no
     las ensancha (su ancho intrínseco empujaba el plano fuera de pantallas chicas). */
  width: 0;
  min-width: 100%;
  display: grid;
  justify-items: center;
  gap: 2px;
  margin-bottom: 0.9rem;
}
svg {
  width: 100%;
  height: 40px;
  display: block;
}
path {
  fill: none;
  stroke: var(--gold);
  stroke-width: 5;
  stroke-linecap: round;
  stroke-dasharray: 0 11;
  vector-effect: non-scaling-stroke;
  filter: drop-shadow(0 0 4px rgba(233, 189, 60, 0.6));
}
span {
  color: var(--muted);
  font-size: 0.8rem;
  letter-spacing: 0.08em;
}
</style>
