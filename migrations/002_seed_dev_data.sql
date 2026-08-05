-- 002_seed_dev_data.sql
-- Datos de ejemplo para desarrollo local: un evento con una sala
-- pequeña de 5 filas x 8 asientos (40 asientos), suficiente para
-- probar visualmente el mapa de asientos en tiempo real.

INSERT INTO events (id, name, venue, starts_at)
VALUES (
    '123e4567-e89b-12d3-a456-426614174000',
    'Concierto de Prueba',
    'Sala CachyOS Arena',
    now() + interval '30 days'
);

DO $$
DECLARE
    r INTEGER;
    l TEXT;
BEGIN
    FOR r IN 1..5 LOOP
        FOREACH l IN ARRAY ARRAY['A','B','C','D','E','F','G','H'] LOOP
            INSERT INTO seats (event_id, row_number, label, status)
            VALUES ('123e4567-e89b-12d3-a456-426614174000', r, l, 'AVAILABLE');
        END LOOP;
    END LOOP;
END $$;
