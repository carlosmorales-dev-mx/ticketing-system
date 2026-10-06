export default defineNuxtConfig({
  compatibilityDate: "2025-01-01",
  devtools: { enabled: true },

  css: ["~/assets/main.css"],

  app: {
    head: {
      title: "Venta de entradas",
      meta: [{ name: "description", content: "Mapa de asientos en tiempo real" }],
      link: [
        { rel: "preconnect", href: "https://fonts.googleapis.com" },
        { rel: "preconnect", href: "https://fonts.gstatic.com", crossorigin: "" },
        {
          rel: "stylesheet",
          href: "https://fonts.googleapis.com/css2?family=Archivo:wght@400;500;600;700&family=Big+Shoulders+Display:wght@700;800&family=IBM+Plex+Mono:wght@500&display=swap",
        },
      ],
    },
  },

  // runtimeConfig.public queda disponible en el cliente. Las URLs
  // vienen de variables de entorno (NUXT_PUBLIC_*) para no hardcodear
  // localhost en el bundle de producción.
  runtimeConfig: {
    public: {
      apiBase: process.env.NUXT_PUBLIC_API_BASE || "http://localhost:8080",
      wsBase: process.env.NUXT_PUBLIC_WS_BASE || "ws://localhost:8080",
      // Botón "Reiniciar mapa": herramienta de desarrollo. Visible en dev,
      // oculto en producción salvo que NUXT_PUBLIC_ENABLE_RESET=true.
      enableReset: process.env.NUXT_PUBLIC_ENABLE_RESET
        ? process.env.NUXT_PUBLIC_ENABLE_RESET === "true"
        : process.env.NODE_ENV !== "production",
    },
  },
});
