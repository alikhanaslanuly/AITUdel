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
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		dsn = "host=localhost port=5432 user=postgres password=postgres dbname=order_db sslmode=disable"
	}

	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	require.NoError(t, db.Ping())
	return db
}

func TestOrderRepo_CreateAndGet(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("skipping integration test")
	}

	db := getTestDB(t)
	defer db.Close()

	repo := repository.NewOrderRepository(db)
	ctx := context.Background()

	order := &domain.Order{
		ID:              "test-order-123",
		UserID:          "test-user-456",
		RestaurantID:    "test-rest-789",
		Status:          domain.StatusPending,
		TotalPrice:      2500,
		DeliveryAddress: "Astana, AITU, 5 floor",
		Items: []domain.OrderItem{
			{
				ID:       "item-1",
				ItemID:   "menu-item-1",
				Name:     "Burger",
				Quantity: 2,
				Price:    1000,
			},
			{
				ID:       "item-2",
				ItemID:   "menu-item-2",
				Name:     "Cola",
				Quantity: 1,
				Price:    500,
			},
		},
	}

	err := repo.Create(ctx, order)
	require.NoError(t, err)

	got, err := repo.GetByID(ctx, order.ID)
	require.NoError(t, err)

	assert.Equal(t, order.ID, got.ID)
	assert.Equal(t, order.UserID, got.UserID)
	assert.Equal(t, order.Status, got.Status)
	assert.Len(t, got.Items, 2)

	db.Exec("DELETE FROM order_items WHERE order_id = $1", order.ID)
	db.Exec("DELETE FROM orders WHERE id = $1", order.ID)
}
