-- +goose Up
-- Initial schema for flight booking
CREATE TABLE IF NOT EXISTS airports (
    id SERIAL PRIMARY KEY,
    code VARCHAR(8) NOT NULL UNIQUE,
    city VARCHAR(128) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS airports;

