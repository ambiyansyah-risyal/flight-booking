-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS routes (
    id SERIAL PRIMARY KEY,
    code VARCHAR(16) NOT NULL UNIQUE,
    origin_code VARCHAR(8) NOT NULL REFERENCES airports(code) ON DELETE RESTRICT,
    destination_code VARCHAR(8) NOT NULL REFERENCES airports(code) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT routes_origin_destination_check CHECK (origin_code <> destination_code)
);
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE routes TO flight_app;
GRANT USAGE, SELECT ON SEQUENCE routes_id_seq TO flight_app;

CREATE TABLE IF NOT EXISTS flight_schedules (
    id SERIAL PRIMARY KEY,
    route_code VARCHAR(16) NOT NULL REFERENCES routes(code) ON DELETE CASCADE,
    airplane_code VARCHAR(16) NOT NULL REFERENCES airplanes(code) ON DELETE RESTRICT,
    departure_date DATE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT flight_schedules_unique UNIQUE (route_code, airplane_code, departure_date)
);
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE flight_schedules TO flight_app;
GRANT USAGE, SELECT ON SEQUENCE flight_schedules_id_seq TO flight_app;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS flight_schedules;
DROP TABLE IF EXISTS routes;
-- +goose StatementEnd
