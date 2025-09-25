package sqlxrepo

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/ambiyansyah-risyal/flight-booking/internal/domain"
	"github.com/jmoiron/sqlx"
)

type AirportRepository struct {
	db *sqlx.DB
}

func NewAirportRepository(db *sqlx.DB) *AirportRepository {
	return &AirportRepository{db: db}
}

func (r *AirportRepository) Create(ctx context.Context, a *domain.Airport) error {
	query := `INSERT INTO airports (code, city) VALUES ($1, $2) RETURNING id, created_at`
	var createdAt time.Time
	if err := r.db.QueryRowContext(ctx, query, a.Code, a.City).Scan(&a.ID, &createdAt); err != nil {
		if isUniqueViolation(err) {
			return domain.ErrAirportExists
		}
		return err
	}
	a.CreatedAt = createdAt.Format(time.RFC3339)
	return nil
}

func (r *AirportRepository) GetByCode(ctx context.Context, code string) (*domain.Airport, error) {
	var out domain.Airport
	row := r.db.QueryRowxContext(ctx, `SELECT id, code, city, created_at FROM airports WHERE code=$1`, code)
	var createdAt time.Time
	if err := row.Scan(&out.ID, &out.Code, &out.City, &createdAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrAirportNotFound
		}
		return nil, err
	}
	out.CreatedAt = createdAt.Format(time.RFC3339)
	return &out, nil
}

func (r *AirportRepository) List(ctx context.Context, limit, offset int) ([]domain.Airport, error) {
	rows, err := r.db.QueryxContext(ctx, `SELECT id, code, city, created_at FROM airports ORDER BY code LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var items []domain.Airport
	for rows.Next() {
		var a domain.Airport
		var createdAt time.Time
		if err := rows.Scan(&a.ID, &a.Code, &a.City, &createdAt); err != nil {
			return nil, err
		}
		a.CreatedAt = createdAt.Format(time.RFC3339)
		items = append(items, a)
	}
	return items, rows.Err()
}

func (r *AirportRepository) Update(ctx context.Context, code string, city string) error {
	res, err := r.db.ExecContext(ctx, `UPDATE airports SET city=$2 WHERE code=$1`, code, city)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrAirportNotFound
	}
	return nil
}

func (r *AirportRepository) Delete(ctx context.Context, code string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM airports WHERE code=$1`, code)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrAirportNotFound
	}
	return nil
}
