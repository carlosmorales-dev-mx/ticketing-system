<script setup lang="ts">
// Redirige directo al evento de prueba que crea la migración
// 002_seed_dev_data.sql, para poder probar el flujo sin buscar el ID.
await navigateTo("/events/123e4567-e89b-12d3-a456-426614174000");
</script>

<template>
  <div />
</template>
