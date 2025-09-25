-- +goose Up
-- +goose StatementBegin
-- Initial schema: airports table
CREATE TABLE IF NOT EXISTS airports (
    id SERIAL PRIMARY KEY,
    code VARCHAR(8) NOT NULL UNIQUE,
    city VARCHAR(128) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
-- Grant DML privileges to app role
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE airports TO flight_app;
-- Seed a couple of airports (idempotent)
INSERT INTO airports (code, city)
    VALUES ('CGK', 'Jakarta'), ('DPS', 'Denpasar')
ON CONFLICT (code) DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS airports;
-- +goose StatementEnd
