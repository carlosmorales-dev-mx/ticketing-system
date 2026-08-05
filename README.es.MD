# Infraestructura — Sistema de Venta de Entradas

Levanta Postgres, Redis, RabbitMQ y Swagger UI con Podman.

## Arrancar

```bash
cd deploy
cp .env.example .env   # ajusta si hace falta
podman-compose -f podman-compose.yaml up -d
```

## Verificar que todo está sano

```bash
podman-compose -f podman-compose.yaml ps
```

Deberías ver `postgres`, `redis`, `rabbitmq` y `swagger-ui` en estado `healthy` / `running`.

## Accesos

| Servicio         | URL / Host                     | Credenciales                          |
|-------------------|---------------------------------|----------------------------------------|
| Postgres          | `localhost:5432`               | `ticketing` / `ticketing_dev_password` |
| Redis             | `localhost:6379`               | sin auth (solo dev)                    |
| RabbitMQ (AMQP)   | `localhost:5672`               | `ticketing` / `ticketing_dev_password` |
| RabbitMQ (panel)  | http://localhost:15672         | `ticketing` / `ticketing_dev_password` |
| Swagger UI        | http://localhost:8081          | —                                       |

## Notas importantes

- **Redis con `notify-keyspace-events Ex`**: esto activa las notificaciones de expiración de claves. Es el mecanismo que usará el servicio Go para enterarse en tiempo real cuando una reserva de 10 minutos caduca, y así liberar el asiento en Postgres automáticamente.
- **Migraciones automáticas**: cualquier `.sql` que pongas en `../migrations/` se ejecutará automáticamente la primera vez que se cree el volumen de Postgres (comportamiento estándar de la imagen oficial vía `docker-entrypoint-initdb.d`). Si ya existe el volumen y quieres reaplicar migraciones desde cero: `podman-compose down -v`.
- **Swagger UI** lee `../api/openapi.yaml` — cuando generemos ese archivo en el siguiente paso, se reflejará automáticamente al recargar http://localhost:8081.

## Parar / limpiar

```bash
podman-compose -f podman-compose.yaml down       # para los contenedores
podman-compose -f podman-compose.yaml down -v     # + borra los volúmenes (reset total)
```