package usecase

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
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

type OrderItem struct {
	ItemID   string `json:"item_id"`
	Quantity int    `json:"quantity"`
}

func (u *StockUsecase) ReserveStock(
	ctx context.Context,
	items []OrderItem,
) error {

	tx, err := u.db.BeginTx(
		ctx,
		nil,
	)

	if err != nil {
		return err
	}

	defer tx.Rollback()

	query := `
		SELECT stock
		FROM menu_items
		WHERE id = $1
		FOR UPDATE
	`

	updateQuery := `
		UPDATE menu_items
		SET stock = stock - $1
		WHERE id = $2
	`

	for _, item := range items {
		itemIDInt, err := strconv.ParseInt(item.ItemID, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid item ID format: %w", err)
		}

		var stock int
		err = tx.QueryRowContext(
			ctx,
			query,
			itemIDInt,
		).Scan(&stock)

		if err != nil {
			return err
		}

		if stock < item.Quantity {
			return fmt.Errorf("not enough stock for item %s", item.ItemID)
		}

		_, err = tx.ExecContext(
			ctx,
			updateQuery,
			item.Quantity,
			itemIDInt,
		)

		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
