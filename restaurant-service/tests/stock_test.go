package tests

import (
	"context"
	"database/sql"
	"testing"

	"restaurant-service/internal/usecase"
)

func TestReserveStock(t *testing.T) {
	db := setupTestDB(t) // skips automatically when INTEGRATION is not set
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {

		}
	}(db)

	_, err := db.Exec(`INSERT INTO categories (id, name) VALUES (9001, 'Test Category') ON CONFLICT (id) DO NOTHING`)
	if err != nil {
		t.Fatalf("seed category: %v", err)
	}
	_, err = db.Exec(`INSERT INTO restaurants (id, category_id, name, description) VALUES (9001, 9001, 'Test Restaurant', 'test') ON CONFLICT (id) DO NOTHING`)
	if err != nil {
		t.Fatalf("seed restaurant: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO menu_items (id, restaurant_id, name, description, price, stock)
		VALUES (9001, 9001, 'Test Item', 'test item', 10.00, 5)
		ON CONFLICT (id) DO UPDATE SET stock = 5
	`)
	if err != nil {
		t.Fatalf("seed menu_item: %v", err)
	}

	t.Cleanup(func() {
		db.Exec("DELETE FROM menu_items  WHERE id = 9001")
		db.Exec("DELETE FROM restaurants WHERE id = 9001")
		db.Exec("DELETE FROM categories  WHERE id = 9001")
	})

	stockUsecase := usecase.NewStockUsecase(db)
	ctx := context.Background()

	err = stockUsecase.ReserveStock(ctx, 9001, 2)
	if err != nil {
		t.Fatalf("ReserveStock(9001, 2) unexpected error: %v", err)
	}

	var stockAfter int
	db.QueryRow("SELECT stock FROM menu_items WHERE id = 9001").Scan(&stockAfter)
	if stockAfter != 3 {
		t.Errorf("expected stock=3 after reserving 2 from 5, got %d", stockAfter)
	}

	err = stockUsecase.ReserveStock(ctx, 9001, 10)
	if err == nil {
		t.Error("expected 'not enough stock' error, got nil")
	}

	db.QueryRow("SELECT stock FROM menu_items WHERE id = 9001").Scan(&stockAfter)
	if stockAfter != 3 {
		t.Errorf("stock should still be 3 after failed reservation, got %d", stockAfter)
	}
}
