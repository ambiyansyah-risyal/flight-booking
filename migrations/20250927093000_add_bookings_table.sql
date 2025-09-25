-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS bookings (
    id SERIAL PRIMARY KEY,
    reference VARCHAR(32) NOT NULL UNIQUE,
    schedule_id INTEGER NOT NULL REFERENCES flight_schedules(id) ON DELETE CASCADE,
    passenger_name VARCHAR(128) NOT NULL,
    seat_number INTEGER NOT NULL,
    status VARCHAR(16) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT bookings_schedule_seat_unique UNIQUE (schedule_id, seat_number)
);
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE bookings TO flight_app;
GRANT USAGE, SELECT ON SEQUENCE bookings_id_seq TO flight_app;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS bookings;
-- +goose StatementEnd
