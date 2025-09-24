-- Initial schema: airports table
CREATE TABLE IF NOT EXISTS airports (
    id SERIAL PRIMARY KEY,
    code VARCHAR(8) NOT NULL UNIQUE,
    city VARCHAR(128) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
-- Seed a couple of airports (idempotent)
INSERT INTO airports (code, city)
    VALUES ('CGK', 'Jakarta'), ('DPS', 'Denpasar')
ON CONFLICT (code) DO NOTHING;

DROP TABLE IF EXISTS airports;
