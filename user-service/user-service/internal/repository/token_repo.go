package repository

import (
	"context"
	"database/sql"
	"user-service/internal/domain"

	"github.com/pkg/errors"
)

type TokenRepository struct {
	db *sql.DB
}

func NewTokenRepository(db *sql.DB) *TokenRepository {
	return &TokenRepository{db: db}
}

func (r *TokenRepository) Save(ctx context.Context, tx *sql.Tx, t *domain.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (id, user_id, token, expires_at, created_at)
		VALUES (gen_random_uuid(), $1, $2, $3, NOW())
		RETURNING id`

	var row *sql.Row
	if tx != nil {
		row = tx.QueryRowContext(ctx, query, t.UserID, t.Token, t.ExpiresAt)
	} else {
		row = r.db.QueryRowContext(ctx, query, t.UserID, t.Token, t.ExpiresAt)
	}
	return row.Scan(&t.ID)
}

func (r *TokenRepository) FindByToken(ctx context.Context, token string) (*domain.RefreshToken, error) {
	query := `SELECT id, user_id, token, expires_at, created_at FROM refresh_tokens WHERE token = $1`
	t := &domain.RefreshToken{}
	err := r.db.QueryRowContext(ctx, query, token).
		Scan(&t.ID, &t.UserID, &t.Token, &t.ExpiresAt, &t.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return t, err
}

func (r *TokenRepository) Delete(ctx context.Context, token string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM refresh_tokens WHERE token = $1`, token)
	return err
}

func (r *TokenRepository) DeleteAllForUser(ctx context.Context, userID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM refresh_tokens WHERE user_id = $1`, userID)
	return err
}
