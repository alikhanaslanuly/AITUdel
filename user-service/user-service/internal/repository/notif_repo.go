package repository

import (
	"context"
	"database/sql"
	"user-service/internal/domain"
)

type NotifRepository struct {
	db *sql.DB
}

func NewNotifRepository(db *sql.DB) *NotifRepository {
	return &NotifRepository{db: db}
}

func (r *NotifRepository) Log(ctx context.Context, tx *sql.Tx, n *domain.NotificationLog) error {
	query := `
		INSERT INTO notifications_log (id, user_id, type, subject, success, sent_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, NOW())
		RETURNING id`

	var row *sql.Row
	if tx != nil {
		row = tx.QueryRowContext(ctx, query, n.UserID, n.Type, n.Subject, n.Success)
	} else {
		row = r.db.QueryRowContext(ctx, query, n.UserID, n.Type, n.Subject, n.Success)
	}
	return row.Scan(&n.ID)
}
