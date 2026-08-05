# Arquitectura

Arquitectura Hexagonal (Puertos y Adaptadores) + Domain-Driven Design.

## Principio

Las reglas de negocio (`domain`) no dependen de ningún framework, driver
de base de datos, ni protocolo de red. Todo lo externo se comunica con
el núcleo a través de **interfaces (puertos)**, y cada tecnología
concreta (Postgres, Redis, RabbitMQ, HTTP, WebSocket) es un
**adaptador** intercambiable que implementa esos puertos.

```mermaid
flowchart TB
    subgraph EXT_IN["Adaptadores de ENTRADA"]
        HTTP["REST API<br/>(handlers)"]
        WS["WebSocket Hub"]
        AMQP_IN["Consumidor RabbitMQ<br/>(sala de espera)"]
    end

    subgraph CORE["Núcleo hexagonal"]
        direction TB
        PORT_IN["Puertos de entrada<br/>(interfaces: ReserveSeatUseCase...)"]
        UC["Casos de uso<br/>(application/usecase)"]
        DOMAIN["Dominio<br/>(Seat, Reservation, Event, Ticket)"]
        PORT_OUT["Puertos de salida<br/>(interfaces: SeatRepository...)"]

        PORT_IN --> UC
        UC --> DOMAIN
        UC --> PORT_OUT
    end

    subgraph EXT_OUT["Adaptadores de SALIDA"]
        PG[("Postgres")]
        REDIS[("Redis")]
        RMQ_OUT["RabbitMQ<br/>(publisher)"]
    end

    HTTP --> PORT_IN
    WS -.notifica.-> EXT_IN
    AMQP_IN --> PORT_IN

    PORT_OUT --> PG
    PORT_OUT --> REDIS
    PORT_OUT --> RMQ_OUT
```

## Regla de dependencias

```
adapters  →  application  →  domain
   (nunca al revés)
```

- `domain` no importa nada fuera de Go estándar.
- `application` solo importa `domain` y define interfaces (`port/in`, `port/out`).
- `adapters` implementa esas interfaces con tecnología concreta.
- `cmd/api/main.go` es el único punto que conoce todas las capas a la vez (composition root / inyección de dependencias manual).

## Estructura de carpetas

```
api/
├── cmd/api/main.go              # composition root
├── internal/
│   ├── domain/                  # entidades + reglas de negocio puras
│   │   ├── seat/                # Seat: AVAILABLE -> RESERVED -> SOLD
│   │   ├── reservation/         # Reservation + TTL de 10 min
│   │   ├── event/
│   │   ├── ticket/
│   │   └── shared/               # Value Objects (ID) + DomainError
│   ├── application/
│   │   ├── port/in/             # interfaces de casos de uso
│   │   ├── port/out/            # interfaces hacia infraestructura
│   │   └── usecase/             # implementación de los casos de uso
│   └── adapters/
│       ├── in/http/             # REST (handlers, middleware, dto)
│       ├── in/websocket/        # tiempo real
│       ├── in/amqpconsumer/     # sala de espera virtual
│       └── out/{postgres,redis,rabbitmq}/
└── openapi.yaml
```

## Flujo de negocio: Reservar → Esperar 10 min → Pagar

```mermaid
sequenceDiagram
    actor Cliente
    participant API as API (Go)
    participant PG as Postgres
    participant R as Redis

    Cliente->>API: POST /reservations {event_id, seat_id}
    API->>PG: TryReserve (UPDATE ... WHERE status='AVAILABLE')
    alt asiento ya reservado
        PG-->>API: 0 filas afectadas
        API-->>Cliente: 409 SEAT_ALREADY_RESERVED
    else asiento disponible
        PG-->>API: 1 fila afectada
        API->>PG: INSERT reservation (status=PENDING)
        API->>R: SET reservation:{id} EX=600
        API-->>Cliente: 201 { reservation_id, expires_in: 600 }

        par el cliente paga a tiempo
            Cliente->>API: POST /reservations/{id}/confirm
            API->>PG: UPDATE seat SET status='SOLD'
            API->>R: DEL reservation:{id}
            API-->>Cliente: 200 { ticket_id }
        and el cliente NO paga a tiempo
            R--)API: evento "expired" (keyspace notification)
            API->>PG: UPDATE seat SET status='AVAILABLE'
            API-->>Cliente: WebSocket: seat.status = AVAILABLE
        end
    end
```

Este mismo diagrama debería vivir también en el `README.md` principal
del repo (ver punto 4 del checklist de documentación).
