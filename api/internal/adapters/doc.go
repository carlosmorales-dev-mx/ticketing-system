// Package adapters contiene el mundo exterior: todo lo que habla con
// frameworks, protocolos y tecnologías concretas.
//
//   - adapters/in  -> "impulsan" la aplicación (alguien llama al
//     dominio desde fuera): handlers HTTP, el hub de WebSocket, el
//     consumidor de RabbitMQ, el suscriptor de expiraciones de Redis.
//     Implementan/invocan los puertos de application/port/in.
//
//   - adapters/out -> son "impulsados" por la aplicación (el dominio
//     necesita algo del exterior): repositorios Postgres, cliente
//     Redis, publisher de RabbitMQ. Implementan los puertos de
//     application/port/out.
//
// Ningún paquete de `domain` o `application` importa nada de aquí.
// Es esta capa la que importa hacia adentro, nunca al revés.
package adapters
