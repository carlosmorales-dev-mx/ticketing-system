<!--
  Escenario animado: trusses, reflectores en movimiento (sweep),
  luces colgantes con balanceo, tiras LED parpadeantes y estrellas de
  fondo. Es puramente decorativo — vive separado del SeatMap para no
  mezclar la parte "de negocio" (asientos reales) con la ambientación.
-->
<template>
  <div class="stage-wrapper">
    <svg viewBox="0 0 800 500" xmlns="http://www.w3.org/2000/svg" role="img" aria-label="Escenario del concierto">
      <defs>
        <pattern id="truss-pattern" x="0" y="0" width="24" height="24" patternUnits="userSpaceOnUse">
          <rect width="24" height="24" fill="#1a1a1a" />
          <line x1="0" y1="0" x2="24" y2="24" stroke="#444" stroke-width="2.5" />
          <line x1="24" y1="0" x2="0" y2="24" stroke="#444" stroke-width="2.5" />
          <rect width="24" height="24" fill="none" stroke="#111" stroke-width="1" />
        </pattern>
        <linearGradient id="beam-cyan" x1="0%" y1="100%" x2="0%" y2="0%">
          <stop offset="0%" stop-color="#01e1ee" stop-opacity="0.8" />
          <stop offset="100%" stop-color="#01e1ee" stop-opacity="0.0" />
        </linearGradient>
        <linearGradient id="beam-pink" x1="0%" y1="100%" x2="0%" y2="0%">
          <stop offset="0%" stop-color="#fc2c93" stop-opacity="0.8" />
          <stop offset="100%" stop-color="#fc2c93" stop-opacity="0.0" />
        </linearGradient>
        <linearGradient id="beam-yellow" x1="0%" y1="100%" x2="0%" y2="0%">
          <stop offset="0%" stop-color="#fed52f" stop-opacity="0.8" />
          <stop offset="100%" stop-color="#fed52f" stop-opacity="0.0" />
        </linearGradient>
      </defs>

      <rect x="0" y="0" width="800" height="500" fill="#0a0515" />

      <g opacity="0.6">
        <circle cx="120" cy="80" r="1.5" fill="#fff" class="star" />
        <circle cx="300" cy="45" r="2" fill="#fff" class="star" />
        <circle cx="500" cy="70" r="1.5" fill="#fff" class="star" />
        <circle cx="680" cy="50" r="2" fill="#fff" class="star" />
        <circle cx="750" cy="120" r="1.5" fill="#fff" class="star" />
        <circle cx="80" cy="150" r="1" fill="#fff" class="star" />
        <circle cx="220" cy="120" r="2" fill="#fff" class="star" />
        <circle cx="600" cy="130" r="1.5" fill="#fff" class="star" />
      </g>

      <rect x="40" y="420" width="720" height="20" fill="#111" />
      <rect x="80" y="440" width="640" height="10" fill="#222" />
      <rect x="20" y="450" width="760" height="10" fill="#0f0f0f" />

      <rect x="100" y="260" width="30" height="160" fill="url(#truss-pattern)" />
      <rect x="670" y="260" width="30" height="160" fill="url(#truss-pattern)" />
      <rect x="80" y="230" width="640" height="30" fill="url(#truss-pattern)" />
      <rect x="100" y="200" width="40" height="30" fill="url(#truss-pattern)" />
      <rect x="660" y="200" width="40" height="30" fill="url(#truss-pattern)" />
      <polygon points="70,230 400,130 730,230" fill="#222" />
      <polygon points="70,230 400,140 730,230" fill="#1a1a1a" />
      <line x1="70" y1="230" x2="730" y2="230" stroke="#2a2a2a" stroke-width="4" />

      <g>
        <rect x="110" y="232" width="20" height="6" fill="#01e1ee" class="led-strip" />
        <rect x="150" y="232" width="20" height="6" fill="#fc2c93" class="led-strip" />
        <rect x="190" y="232" width="20" height="6" fill="#fed52f" class="led-strip" />
        <rect x="230" y="232" width="20" height="6" fill="#01e1ee" class="led-strip" />
        <rect x="270" y="232" width="20" height="6" fill="#fc2c93" class="led-strip" />
        <rect x="310" y="232" width="20" height="6" fill="#fed52f" class="led-strip" />
        <rect x="350" y="232" width="20" height="6" fill="#01e1ee" class="led-strip" />
        <rect x="390" y="232" width="20" height="6" fill="#fc2c93" class="led-strip" />
        <rect x="430" y="232" width="20" height="6" fill="#fed52f" class="led-strip" />
        <rect x="470" y="232" width="20" height="6" fill="#01e1ee" class="led-strip" />
        <rect x="510" y="232" width="20" height="6" fill="#fc2c93" class="led-strip" />
        <rect x="550" y="232" width="20" height="6" fill="#fed52f" class="led-strip" />
        <rect x="590" y="232" width="20" height="6" fill="#01e1ee" class="led-strip" />
        <rect x="630" y="232" width="20" height="6" fill="#fc2c93" class="led-strip" />
        <rect x="670" y="232" width="20" height="6" fill="#fed52f" class="led-strip" />
      </g>

      <g class="hanging-light-left">
        <line x1="250" y1="230" x2="250" y2="255" stroke="#555" stroke-width="6" />
        <polygon points="235,255 265,255 270,285 230,285" fill="#333" />
        <ellipse cx="250" cy="285" rx="20" ry="5" fill="#fff" class="bulb-glow" />
      </g>
      <g class="hanging-light-right">
        <line x1="550" y1="230" x2="550" y2="255" stroke="#555" stroke-width="6" />
        <polygon points="535,255 565,255 570,285 530,285" fill="#333" />
        <ellipse cx="550" cy="285" rx="20" ry="5" fill="#fff" class="bulb-glow" />
      </g>

      <g class="light-left">
        <polygon points="250,460 50,200 450,200" fill="url(#beam-cyan)" class="beam-fill" />
        <path d="M225,460 L275,460 L260,420 L240,420 Z" fill="#222" />
        <path d="M240,420 L260,420 L270,400 L230,400 Z" fill="#111" />
        <ellipse cx="250" cy="460" rx="35" ry="8" fill="#333" />
      </g>
      <g class="light-center">
        <polygon points="400,460 280,200 520,200" fill="url(#beam-pink)" class="beam-fill" />
        <path d="M375,460 L425,460 L410,420 L390,420 Z" fill="#222" />
        <path d="M390,420 L410,420 L420,400 L380,400 Z" fill="#111" />
        <ellipse cx="400" cy="460" rx="35" ry="8" fill="#333" />
      </g>
      <g class="light-right">
        <polygon points="550,460 350,200 750,200" fill="url(#beam-yellow)" class="beam-fill" />
        <path d="M525,460 L575,460 L560,420 L540,420 Z" fill="#222" />
        <path d="M540,420 L560,420 L570,400 L530,400 Z" fill="#111" />
        <ellipse cx="550" cy="460" rx="35" ry="8" fill="#333" />
      </g>

      <g>
        <rect x="120" y="415" width="40" height="4" fill="#01e1ee" opacity="0.8" class="floor-light" />
        <rect x="180" y="415" width="40" height="4" fill="#fc2c93" opacity="0.8" class="floor-light" />
        <rect x="240" y="415" width="40" height="4" fill="#fed52f" opacity="0.8" class="floor-light" />
        <rect x="520" y="415" width="40" height="4" fill="#01e1ee" opacity="0.8" class="floor-light" />
        <rect x="580" y="415" width="40" height="4" fill="#fc2c93" opacity="0.8" class="floor-light" />
        <rect x="640" y="415" width="40" height="4" fill="#fed52f" opacity="0.8" class="floor-light" />
      </g>
    </svg>
  </div>
</template>

<style scoped>
.stage-wrapper {
  background: #0a0515;
  border: var(--border-heavy);
  border-radius: var(--radius);
  margin-bottom: 1.8rem;
  padding: 0.5rem;
  box-shadow: 4px 4px 0px var(--text-dark);
  overflow: hidden;
}
.stage-wrapper svg {
  display: block;
  width: 100%;
  height: auto;
  max-height: 250px;
  margin: 0 auto;
}

@media (max-width: 900px) {
  .stage-wrapper svg {
    max-height: 180px;
  }
}
@media (max-width: 500px) {
  .stage-wrapper svg {
    max-height: 120px;
  }
}

@keyframes sweep-left {
  0% { transform: rotate(-20deg); }
  50% { transform: rotate(20deg); }
  100% { transform: rotate(-20deg); }
}
@keyframes sweep-right {
  0% { transform: rotate(20deg); }
  50% { transform: rotate(-20deg); }
  100% { transform: rotate(20deg); }
}
@keyframes sweep-center {
  0% { transform: rotate(-10deg); }
  50% { transform: rotate(10deg); }
  100% { transform: rotate(-10deg); }
}
.light-left { animation: sweep-left 4s ease-in-out infinite; transform-origin: 250px 460px; }
.light-right { animation: sweep-right 4s ease-in-out infinite; transform-origin: 550px 460px; }
.light-center { animation: sweep-center 3s ease-in-out infinite; transform-origin: 400px 460px; }

@keyframes flicker {
  0% { opacity: 0.2; } 20% { opacity: 0.7; } 40% { opacity: 0.3; }
  60% { opacity: 0.9; } 80% { opacity: 0.4; } 100% { opacity: 0.6; }
}
.beam-fill { animation: flicker 0.15s infinite alternate; }

@keyframes sway {
  0% { transform: rotate(-4deg); } 50% { transform: rotate(4deg); } 100% { transform: rotate(-4deg); }
}
.hanging-light-left { animation: sway 5s ease-in-out infinite; transform-origin: 250px 230px; }
.hanging-light-right { animation: sway 5s ease-in-out infinite 1s; transform-origin: 550px 230px; }

@keyframes glow-pulse {
  0% { filter: drop-shadow(0 0 5px #fff); }
  50% { filter: drop-shadow(0 0 20px #fff); }
  100% { filter: drop-shadow(0 0 5px #fff); }
}
.bulb-glow { animation: glow-pulse 1.5s infinite alternate; }

@keyframes led-blink {
  0% { opacity: 0.4; } 50% { opacity: 1; } 100% { opacity: 0.4; }
}
.led-strip { animation: led-blink 1.2s infinite alternate; }
.led-strip:nth-child(2) { animation-delay: 0.2s; }
.led-strip:nth-child(3) { animation-delay: 0.4s; }
.led-strip:nth-child(4) { animation-delay: 0.6s; }
.led-strip:nth-child(5) { animation-delay: 0.8s; }
.led-strip:nth-child(6) { animation-delay: 1.0s; }

@keyframes floor-pulse {
  0% { opacity: 0.3; transform: scaleY(1); }
  50% { opacity: 1; transform: scaleY(1.5); }
  100% { opacity: 0.3; transform: scaleY(1); }
}
.floor-light { animation: floor-pulse 2s infinite alternate; transform-origin: center; }
.floor-light:nth-child(2) { animation-delay: 0.3s; }
.floor-light:nth-child(3) { animation-delay: 0.6s; }
.floor-light:nth-child(4) { animation-delay: 0.9s; }
.floor-light:nth-child(5) { animation-delay: 1.2s; }
.floor-light:nth-child(6) { animation-delay: 1.5s; }

@keyframes twinkle {
  0% { opacity: 0.2; } 50% { opacity: 1; } 100% { opacity: 0.2; }
}
.star { animation: twinkle 3s infinite alternate; }
.star:nth-child(2) { animation-delay: 0.5s; }
.star:nth-child(3) { animation-delay: 1s; }
.star:nth-child(4) { animation-delay: 1.5s; }
.star:nth-child(5) { animation-delay: 2s; }
.star:nth-child(6) { animation-delay: 2.5s; }
.star:nth-child(7) { animation-delay: 0.8s; }
.star:nth-child(8) { animation-delay: 1.8s; }

@media (prefers-reduced-motion: reduce) {
  .light-left, .light-right, .light-center, .beam-fill,
  .hanging-light-left, .hanging-light-right, .bulb-glow,
  .led-strip, .floor-light, .star {
    animation: none;
  }
}
</style>
