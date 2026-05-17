package usecase

import (
	"context"
	"database/sql"
	"errors"
)

type StockUsecase struct {
	db *sql.DB
}

func NewStockUsecase(
	db *sql.DB,
) *StockUsecase {

	return &StockUsecase{
		db: db,
	}
}

func (u *StockUsecase) ReserveStock(
	ctx context.Context,
	itemID int64,
	quantity int,
) error {

	tx, err := u.db.BeginTx(
		ctx,
		nil,
	)

	if err != nil {
		return err
	}

	defer tx.Rollback()

	var stock int

	query := `
		SELECT stock
		FROM menu_items
		WHERE id = $1
		FOR UPDATE
	`

	err = tx.QueryRowContext(
		ctx,
		query,
		itemID,
	).Scan(&stock)

	if err != nil {
		return err
	}

	if stock < quantity {
		return errors.New(
			"not enough stock",
		)
	}

	updateQuery := `
		UPDATE menu_items
		SET stock = stock - $1
		WHERE id = $2
	`

	_, err = tx.ExecContext(
		ctx,
		updateQuery,
		quantity,
		itemID,
	)

	if err != nil {
		return err
	}

	return tx.Commit()
}
