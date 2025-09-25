package sqlxrepo

import (
    _ "github.com/jackc/pgx/v5/stdlib"
    "github.com/jmoiron/sqlx"
)

func New(dbURL string) (*sqlx.DB, error) {
    return sqlx.Open("pgx", dbURL)
}

