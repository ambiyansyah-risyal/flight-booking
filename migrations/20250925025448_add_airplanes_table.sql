-- +goose Up
-- +goose StatementBegin
-- Create airplanes table
CREATE TABLE IF NOT EXISTS airplanes (
    id SERIAL PRIMARY KEY,
    code VARCHAR(16) NOT NULL UNIQUE,
    seat_capacity INT NOT NULL CHECK (seat_capacity > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
-- Grant DML to app role
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE airplanes TO flight_app;
GRANT USAGE, SELECT ON SEQUENCE airplanes_id_seq TO flight_app;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS airplanes;
-- +goose StatementEnd
