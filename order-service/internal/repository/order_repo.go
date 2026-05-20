package repository

import (
	"context"
	"database/sql"
	"fmt"
	"order-service/internal/domain"
)

type OrderRepository interface {
	Create(ctx context.Context, order *domain.Order) error
	GetByID(ctx context.Context, id string) (*domain.Order, error)
	GetByUserID(ctx context.Context, userID string, page, limit int) ([]*domain.Order, int, error)
	UpdateStatus(ctx context.Context, id, status string) error
	UpdatePromo(ctx context.Context, id, promoCode string, discount, newTotal float64) error
}

type orderRepo struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) OrderRepository {
	return &orderRepo{db: db}
}

func (r *orderRepo) Create(ctx context.Context, order *domain.Order) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	query := `
		INSERT INTO orders (id, user_id, restaurant_id, status, total_price, promo_code, discount_amount, delivery_address)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err = tx.ExecContext(ctx, query,
		order.ID,
		order.UserID,
		order.RestaurantID,
		order.Status,
		order.TotalPrice,
		order.PromoCode,
		order.DiscountAmount,
		order.DeliveryAddress,
	)
	if err != nil {
		return fmt.Errorf("insert order: %w", err)
	}

	itemQuery := `
		INSERT INTO order_items (id, order_id, item_id, name, quantity, price)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	for _, item := range order.Items {
		_, err = tx.ExecContext(ctx, itemQuery,
			item.ID,
			order.ID,
			item.ItemID,
			item.Name,
			item.Quantity,
			item.Price,
		)
		if err != nil {
			return fmt.Errorf("insert order item: %w", err)
		}
	}

	return tx.Commit()
}

func (r *orderRepo) GetByID(ctx context.Context, id string) (*domain.Order, error) {
	order := &domain.Order{}

	query := `
		SELECT id, user_id, COALESCE(restaurant_id::text, ''), status, total_price,
			COALESCE(promo_code, ''), discount_amount, COALESCE(delivery_address, ''),
			COALESCE(courier_id::text, ''), created_at, updated_at
		FROM orders WHERE id = $1
	`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&order.ID,
		&order.UserID,
		&order.RestaurantID,
		&order.Status,
		&order.TotalPrice,
		&order.PromoCode,
		&order.DiscountAmount,
		&order.DeliveryAddress,
		&order.CourierID,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("order not found")
	}
	if err != nil {
		return nil, fmt.Errorf("get order: %w", err)
	}

	items, err := r.getItems(ctx, id)
	if err != nil {
		return nil, err
	}
	order.Items = items

	return order, nil
}

func (r *orderRepo) getItems(ctx context.Context, orderID string) ([]domain.OrderItem, error) {
	rows, err := r.db.QueryContext(ctx,
		"SELECT id, order_id, item_id, name, quantity, price FROM order_items WHERE order_id = $1",
		orderID,
	)
	if err != nil {
		return nil, fmt.Errorf("get items: %w", err)
	}
	defer rows.Close()

	var items []domain.OrderItem
	for rows.Next() {
		var item domain.OrderItem
		if err := rows.Scan(&item.ID, &item.OrderID, &item.ItemID, &item.Name, &item.Quantity, &item.Price); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *orderRepo) GetByUserID(ctx context.Context, userID string, page, limit int) ([]*domain.Order, int, error) {
	offset := (page - 1) * limit

	var total int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM orders WHERE user_id = $1", userID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, COALESCE(restaurant_id::text, ''), status, total_price,
			COALESCE(promo_code, ''), discount_amount, COALESCE(delivery_address, ''),
			COALESCE(courier_id::text, ''), created_at, updated_at
		FROM orders WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var orders []*domain.Order
	for rows.Next() {
		o := &domain.Order{}
		err := rows.Scan(
			&o.ID, &o.UserID, &o.RestaurantID, &o.Status, &o.TotalPrice,
			&o.PromoCode, &o.DiscountAmount, &o.DeliveryAddress,
			&o.CourierID, &o.CreatedAt, &o.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		orders = append(orders, o)
	}

	return orders, total, nil
}

func (r *orderRepo) UpdateStatus(ctx context.Context, id, status string) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE orders SET status = $1, updated_at = NOW() WHERE id = $2",
		status, id,
	)
	return err
}

func (r *orderRepo) UpdatePromo(ctx context.Context, id, promoCode string, discount, newTotal float64) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE orders SET promo_code = $1, discount_amount = $2, total_price = $3, updated_at = NOW() WHERE id = $4",
		promoCode, discount, newTotal, id,
	)
	return err
}
