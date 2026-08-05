// Package application contiene los CASOS DE USO: orquestan entidades
// de dominio y puertos de salida para cumplir una intención concreta
// del negocio (ej: "reservar un asiento").
//
// Esta capa SÍ puede depender de `domain`, pero NUNCA de un adaptador
// concreto (postgres, redis, http...). Solo conoce INTERFACES
// (los "puertos"), definidas en application/port.
//
// Subcarpetas:
//   - port/in  -> interfaces que exponen los casos de uso hacia fuera
//                 (las implementa esta capa, las invocan los adaptadores
//                 de entrada como el handler HTTP).
//   - port/out -> interfaces que esta capa necesita del mundo exterior
//                 (las invoca esta capa, las implementan los adaptadores
//                 de salida como el repositorio Postgres).
//   - usecase  -> implementación concreta de cada port/in.
package application
