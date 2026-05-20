package repository

import (
	"context"
	"database/sql"
	"fmt"
	"user-service/internal/domain"

	"github.com/pkg/errors"
)

type CourierRepository struct {
	db *sql.DB
}

func NewCourierRepository(db *sql.DB) *CourierRepository {
	return &CourierRepository{db: db}
}

func (r *CourierRepository) Create(ctx context.Context, tx *sql.Tx, c *domain.Courier) error {
	query := `
		INSERT INTO couriers (id, user_id, status, latitude, longitude, updated_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, NOW())
		RETURNING id`

	var row *sql.Row
	if tx != nil {
		row = tx.QueryRowContext(ctx, query, c.UserID, c.Status, c.Latitude, c.Longitude)
	} else {
		row = r.db.QueryRowContext(ctx, query, c.UserID, c.Status, c.Latitude, c.Longitude)
	}
	return row.Scan(&c.ID)
}

func (r *CourierRepository) FindByID(ctx context.Context, id string) (*domain.Courier, error) {
	query := `SELECT id, user_id, status, latitude, longitude, updated_at FROM couriers WHERE id = $1`
	c := &domain.Courier{}
	err := r.db.QueryRowContext(ctx, query, id).
		Scan(&c.ID, &c.UserID, &c.Status, &c.Latitude, &c.Longitude, &c.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("courier not found")
	}
	return c, err
}

func (r *CourierRepository) FindByUserID(ctx context.Context, userID string) (*domain.Courier, error) {
	query := `SELECT id, user_id, status, latitude, longitude, updated_at FROM couriers WHERE user_id = $1`
	c := &domain.Courier{}
	err := r.db.QueryRowContext(ctx, query, userID).
		Scan(&c.ID, &c.UserID, &c.Status, &c.Latitude, &c.Longitude, &c.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("courier not found")
	}
	return c, err
}

func (r *CourierRepository) UpdateStatus(ctx context.Context, courierID, status string, lat, lng float64) error {
	query := `UPDATE couriers SET status = $1, latitude = $2, longitude = $3, updated_at = NOW() WHERE id = $4`
	_, err := r.db.ExecContext(ctx, query, status, lat, lng, courierID)
	return err
}
