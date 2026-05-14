package repository

import (
	"context"
	"database/sql"
	"fmt"
	"order-service/internal/domain"
)

type PromoRepository interface {
	GetByCode(ctx context.Context, code string) (*domain.PromoCode, error)
	IncrementUsage(ctx context.Context, promoID, userID, orderID string) error
	HasUserUsed(ctx context.Context, promoID, userID string) (bool, error)
}

type promoRepo struct {
	db *sql.DB
}

func NewPromoRepository(db *sql.DB) PromoRepository {
	return &promoRepo{db: db}
}

func (r *promoRepo) GetByCode(ctx context.Context, code string) (*domain.PromoCode, error) {
	promo := &domain.PromoCode{}

	query := `
		SELECT id, code, discount_percent, student_only, max_uses, used_count, expires_at, created_at
		FROM promo_codes WHERE code = $1
	`

	err := r.db.QueryRowContext(ctx, query, code).Scan(
		&promo.ID,
		&promo.Code,
		&promo.DiscountPercent,
		&promo.StudentOnly,
		&promo.MaxUses,
		&promo.UsedCount,
		&promo.ExpiresAt,
		&promo.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("promo not found")
	}
	if err != nil {
		return nil, fmt.Errorf("get promo: %w", err)
	}

	return promo, nil
}

func (r *promoRepo) IncrementUsage(ctx context.Context, promoID, userID, orderID string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx,
		"UPDATE promo_codes SET used_count = used_count + 1 WHERE id = $1",
		promoID,
	)
	if err != nil {
		return fmt.Errorf("increment promo usage: %w", err)
	}

	_, err = tx.ExecContext(ctx,
		"INSERT INTO promo_usages (id, promo_id, user_id, order_id) VALUES (gen_random_uuid(), $1, $2, $3)",
		promoID, userID, orderID,
	)
	if err != nil {
		return fmt.Errorf("insert promo usage: %w", err)
	}

	return tx.Commit()
}

func (r *promoRepo) HasUserUsed(ctx context.Context, promoID, userID string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM promo_usages WHERE promo_id = $1 AND user_id = $2",
		promoID, userID,
	).Scan(&count)
	return count > 0, err
}
