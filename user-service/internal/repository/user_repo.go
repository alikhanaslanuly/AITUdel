package repository

import (
	"context"
	"database/sql"
	"fmt"
	"user-service/internal/domain"

	"github.com/pkg/errors"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, tx *sql.Tx, u *domain.User) error {
	query := `
		INSERT INTO users (id, email, password_hash, name, phone, role, is_student, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, NOW(), NOW())
		RETURNING id, created_at, updated_at`

	var row *sql.Row
	if tx != nil {
		row = tx.QueryRowContext(ctx, query, u.Email, u.PasswordHash, u.Name, u.Phone, u.Role, u.IsStudent)
	} else {
		row = r.db.QueryRowContext(ctx, query, u.Email, u.PasswordHash, u.Name, u.Phone, u.Role, u.IsStudent)
	}

	return row.Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	query := `SELECT id, email, password_hash, name, phone, role, is_student, created_at, updated_at
	          FROM users WHERE id = $1`
	u := &domain.User{}
	err := r.db.QueryRowContext(ctx, query, id).
		Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.Phone, &u.Role, &u.IsStudent, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("user not found")
	}
	return u, err
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `SELECT id, email, password_hash, name, phone, role, is_student, created_at, updated_at
	          FROM users WHERE email = $1`
	u := &domain.User{}
	err := r.db.QueryRowContext(ctx, query, email).
		Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.Phone, &u.Role, &u.IsStudent, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("user not found")
	}
	return u, err
}

func (r *UserRepository) Update(ctx context.Context, u *domain.User) error {
	query := `UPDATE users SET name = $1, phone = $2, updated_at = NOW() WHERE id = $3`
	_, err := r.db.ExecContext(ctx, query, u.Name, u.Phone, u.ID)
	return err
}

func (r *UserRepository) SetPromoAssigned(ctx context.Context, tx *sql.Tx, userID, promo string) error {
	query := `UPDATE users SET promo_assigned = $1 WHERE id = $2`
	if tx != nil {
		_, err := tx.ExecContext(ctx, query, promo, userID)
		return err
	}
	_, err := r.db.ExecContext(ctx, query, promo, userID)
	return err
}

func (r *UserRepository) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return r.db.BeginTx(ctx, nil)
}
