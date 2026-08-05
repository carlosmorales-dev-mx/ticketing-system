-- 001_init_schema.sql
-- Esquema base del sistema de venta de entradas.
-- Se ejecuta automáticamente al crear el contenedor de Postgres
-- (ver deploy/podman-compose.yaml -> docker-entrypoint-initdb.d).

CREATE EXTENSION IF NOT EXISTS "pgcrypto"; -- para gen_random_uuid()

-- ---------------------------------------------------------------------
-- events
-- ---------------------------------------------------------------------
CREATE TABLE events (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       TEXT NOT NULL,
    venue      TEXT NOT NULL,
    starts_at  TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ---------------------------------------------------------------------
-- seats
-- ---------------------------------------------------------------------
-- La combinación (event_id, row_number, label) es única: no puede
-- haber dos asientos "12C" para el mismo evento.
CREATE TABLE seats (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id   UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    row_number INTEGER NOT NULL CHECK (row_number > 0),
    label      TEXT NOT NULL,
    status     TEXT NOT NULL DEFAULT 'AVAILABLE'
               CHECK (status IN ('AVAILABLE', 'RESERVED', 'SOLD')),
    version    INTEGER NOT NULL DEFAULT 0, -- optimistic locking (ver nota abajo)
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_seat_per_event UNIQUE (event_id, row_number, label)
);

-- Índice crítico: es EXACTAMENTE el patrón de consulta de TryReserve
-- (UPDATE ... WHERE id = $1 AND status = 'AVAILABLE'), así que un
-- índice parcial sobre asientos disponibles acelera muchísimo las
-- lecturas de disponibilidad bajo carga.
CREATE INDEX idx_seats_event_available
    ON seats (event_id)
    WHERE status = 'AVAILABLE';

-- ---------------------------------------------------------------------
-- reservations
-- ---------------------------------------------------------------------
-- seat_id es UNIQUE entre reservas "vivas" (ver índice parcial abajo):
-- ese es el segundo cinturón de seguridad, además del UPDATE atómico
-- en seats, para que sea estructuralmente imposible tener dos
-- reservas PENDING sobre el mismo asiento al mismo tiempo.
CREATE TABLE reservations (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id    UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    seat_id     UUID NOT NULL REFERENCES seats(id) ON DELETE CASCADE,
    user_id     UUID NOT NULL,
    status      TEXT NOT NULL DEFAULT 'PENDING'
                CHECK (status IN ('PENDING', 'CONFIRMED', 'EXPIRED', 'CANCELLED')),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at  TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_reservations_seat ON reservations (seat_id);
CREATE INDEX idx_reservations_user ON reservations (user_id);

-- Solo puede existir UNA reserva PENDING por asiento a la vez.
CREATE UNIQUE INDEX uq_seat_pending_reservation
    ON reservations (seat_id)
    WHERE status = 'PENDING';

-- Útil para el job/consulta de barrido de reservas vencidas (fallback
-- si el evento de expiración de Redis se pierde por cualquier motivo).
CREATE INDEX idx_reservations_expires_at
    ON reservations (expires_at)
    WHERE status = 'PENDING';

-- ---------------------------------------------------------------------
-- tickets
-- ---------------------------------------------------------------------
CREATE TABLE tickets (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reservation_id UUID NOT NULL UNIQUE REFERENCES reservations(id) ON DELETE CASCADE,
    seat_id        UUID NOT NULL REFERENCES seats(id),
    user_id        UUID NOT NULL,
    issued_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ---------------------------------------------------------------------
-- Trigger: mantener updated_at en seats
-- ---------------------------------------------------------------------
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_seats_updated_at
    BEFORE UPDATE ON seats
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();
