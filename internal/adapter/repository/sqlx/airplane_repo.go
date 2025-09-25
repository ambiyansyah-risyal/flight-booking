package sqlxrepo

import (
    "context"
    "database/sql"
    "errors"
    "strings"
    "time"

    "github.com/ambiyansyah-risyal/flight-booking/internal/domain"
    "github.com/jmoiron/sqlx"
)

type AirplaneRepository struct { db *sqlx.DB }

func NewAirplaneRepository(db *sqlx.DB) *AirplaneRepository { return &AirplaneRepository{db: db} }

func (r *AirplaneRepository) Create(ctx context.Context, a *domain.Airplane) error {
    q := `INSERT INTO airplanes (code, seat_capacity) VALUES ($1,$2) RETURNING id, created_at`
    var created time.Time
    if err := r.db.QueryRowContext(ctx, q, a.Code, a.SeatCapacity).Scan(&a.ID, &created); err != nil {
        if localUniqueViolation(err) { return domain.ErrAirplaneExists }
        return err
    }
    a.CreatedAt = created.Format(time.RFC3339)
    return nil
}

func (r *AirplaneRepository) GetByCode(ctx context.Context, code string) (*domain.Airplane, error) {
    var out domain.Airplane
    var created time.Time
    err := r.db.QueryRowContext(ctx, `SELECT id, code, seat_capacity, created_at FROM airplanes WHERE code=$1`, code).
        Scan(&out.ID, &out.Code, &out.SeatCapacity, &created)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) { return nil, domain.ErrAirplaneNotFound }
        return nil, err
    }
    out.CreatedAt = created.Format(time.RFC3339)
    return &out, nil
}

func (r *AirplaneRepository) List(ctx context.Context, limit, offset int) ([]domain.Airplane, error) {
    rows, err := r.db.QueryxContext(ctx, `SELECT id, code, seat_capacity, created_at FROM airplanes ORDER BY code LIMIT $1 OFFSET $2`, limit, offset)
    if err != nil { return nil, err }
    defer func(){ _ = rows.Close() }()
    var items []domain.Airplane
    for rows.Next() {
        var a domain.Airplane
        var created time.Time
        if err := rows.Scan(&a.ID, &a.Code, &a.SeatCapacity, &created); err != nil { return nil, err }
        a.CreatedAt = created.Format(time.RFC3339)
        items = append(items, a)
    }
    return items, rows.Err()
}

func (r *AirplaneRepository) UpdateSeats(ctx context.Context, code string, seats int) error {
    res, err := r.db.ExecContext(ctx, `UPDATE airplanes SET seat_capacity=$2 WHERE code=$1`, code, seats)
    if err != nil { return err }
    n, _ := res.RowsAffected()
    if n == 0 { return domain.ErrAirplaneNotFound }
    return nil
}

func (r *AirplaneRepository) Delete(ctx context.Context, code string) error {
    res, err := r.db.ExecContext(ctx, `DELETE FROM airplanes WHERE code=$1`, code)
    if err != nil { return err }
    n, _ := res.RowsAffected()
    if n == 0 { return domain.ErrAirplaneNotFound }
    return nil
}

// localUniqueViolation avoids colliding with airport repo's helper name in linters/build
func localUniqueViolation(err error) bool {
    s := strings.ToLower(err.Error())
    return strings.Contains(s, "duplicate key") || strings.Contains(s, "unique constraint")
}
