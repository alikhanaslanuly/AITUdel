package tests

import (
	"context"
	"database/sql"
	"order-service/internal/domain"
	"order-service/internal/repository"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getTestDB(t *testing.T) *sql.DB {
	t.Helper()
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("skipping integration test; set INTEGRATION=1 to run")
	}

	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		dsn = "host=localhost port=5432 user=postgres password=postgres dbname=order_db sslmode=disable"
	}

	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err, "open db")
	require.NoError(t, db.Ping(), "ping db")
	return db
}

func TestOrderRepo_CreateAndGet(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	repo := repository.NewOrderRepository(db)
	ctx := context.Background()

	// Use valid UUIDs — the orders and order_items tables use UUID primary keys.
	orderID := "a0000000-0000-0000-0000-000000000001"
	userID := "b0000000-0000-0000-0000-000000000001"
	restaurantID := "c0000000-0000-0000-0000-000000000001"
	item1ID := "d0000000-0000-0000-0000-000000000001"
	item2ID := "d0000000-0000-0000-0000-000000000002"
	menuItem1ID := "e0000000-0000-0000-0000-000000000001"
	menuItem2ID := "e0000000-0000-0000-0000-000000000002"

	cleanup := func() {
		db.Exec("DELETE FROM order_items WHERE order_id = $1", orderID)
		db.Exec("DELETE FROM orders WHERE id = $1", orderID)
	}
	cleanup()
	t.Cleanup(cleanup)

	order := &domain.Order{
		ID:              orderID,
		UserID:          userID,
		RestaurantID:    restaurantID,
		Status:          domain.StatusPending,
		DeliveryAddress: "Astana, AITU, 5th floor",
		Items: []domain.OrderItem{
			{
				ID:       item1ID,
				ItemID:   menuItem1ID,
				Name:     "Burger",
				Quantity: 2,
				Price:    1000,
			},
			{
				ID:       item2ID,
				ItemID:   menuItem2ID,
				Name:     "Cola",
				Quantity: 1,
				Price:    500,
			},
		},
	}
	order.TotalPrice = order.CalcTotal() // 2500

	err := repo.Create(ctx, order)
	require.NoError(t, err, "create order")

	got, err := repo.GetByID(ctx, order.ID)
	require.NoError(t, err, "get order by id")

	assert.Equal(t, order.ID, got.ID)
	assert.Equal(t, order.UserID, got.UserID)
	assert.Equal(t, order.RestaurantID, got.RestaurantID)
	assert.Equal(t, order.Status, got.Status)
	assert.Equal(t, order.TotalPrice, got.TotalPrice)
	assert.Equal(t, order.DeliveryAddress, got.DeliveryAddress)
	assert.Len(t, got.Items, 2)
}

func TestOrderRepo_UpdateStatus(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	repo := repository.NewOrderRepository(db)
	ctx := context.Background()

	orderID := "a0000000-0000-0000-0000-000000000002"
	userID := "b0000000-0000-0000-0000-000000000002"
	restaurantID := "c0000000-0000-0000-0000-000000000002"

	cleanup := func() {
		db.Exec("DELETE FROM order_items WHERE order_id = $1", orderID)
		db.Exec("DELETE FROM orders WHERE id = $1", orderID)
	}
	cleanup()
	t.Cleanup(cleanup)

	order := &domain.Order{
		ID:           orderID,
		UserID:       userID,
		RestaurantID: restaurantID,
		Status:       domain.StatusPending,
		Items: []domain.OrderItem{
			{
				ID:       "d0000000-0000-0000-0000-000000000003",
				ItemID:   "e0000000-0000-0000-0000-000000000003",
				Name:     "Pizza",
				Quantity: 1,
				Price:    2000,
			},
		},
	}
	order.TotalPrice = order.CalcTotal()

	require.NoError(t, repo.Create(ctx, order))

	err := repo.UpdateStatus(ctx, orderID, domain.StatusConfirmed)
	require.NoError(t, err, "update status")

	got, err := repo.GetByID(ctx, ordыerID)
	require.NoError(t, err)
	assert.Equal(t, domain.StatusConfirmed, got.Status)
}
