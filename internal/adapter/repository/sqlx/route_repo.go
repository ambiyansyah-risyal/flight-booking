package sqlxrepo

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/ambiyansyah-risyal/flight-booking/internal/domain"
	"github.com/jmoiron/sqlx"
)

// RouteRepository persists routes using sqlx.
type RouteRepository struct {
	db *sqlx.DB
}

func NewRouteRepository(db *sqlx.DB) *RouteRepository {
	return &RouteRepository{db: db}
}

func (r *RouteRepository) Create(ctx context.Context, route *domain.Route) error {
	query := `INSERT INTO routes (code, origin_code, destination_code) VALUES ($1, $2, $3) RETURNING id, created_at`
	var createdAt time.Time
	if err := r.db.QueryRowContext(ctx, query, route.Code, route.OriginCode, route.DestinationCode).Scan(&route.ID, &createdAt); err != nil {
		if isUniqueViolation(err) {
			return domain.ErrRouteExists
		}
		return err
	}
	route.CreatedAt = createdAt.Format(time.RFC3339)
	return nil
}

func (r *RouteRepository) GetByCode(ctx context.Context, code string) (*domain.Route, error) {
	var out domain.Route
	row := r.db.QueryRowxContext(ctx, `SELECT id, code, origin_code, destination_code, created_at FROM routes WHERE code=$1`, code)
	var createdAt time.Time
	if err := row.Scan(&out.ID, &out.Code, &out.OriginCode, &out.DestinationCode, &createdAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrRouteNotFound
		}
		return nil, err
	}
	out.CreatedAt = createdAt.Format(time.RFC3339)
	return &out, nil
}

func (r *RouteRepository) List(ctx context.Context, limit, offset int) ([]domain.Route, error) {
	rows, err := r.db.QueryxContext(ctx, `SELECT id, code, origin_code, destination_code, created_at FROM routes ORDER BY code LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []domain.Route
	for rows.Next() {
		var route domain.Route
		var createdAt time.Time
		if err := rows.Scan(&route.ID, &route.Code, &route.OriginCode, &route.DestinationCode, &createdAt); err != nil {
			return nil, err
		}
		route.CreatedAt = createdAt.Format(time.RFC3339)
		out = append(out, route)
	}
	return out, rows.Err()
}

func (r *RouteRepository) Delete(ctx context.Context, code string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM routes WHERE code=$1`, code)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrRouteNotFound
	}
	return nil
}
