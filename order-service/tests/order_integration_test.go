package tests

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"order-service/internal/domain"
	"order-service/internal/repository"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func startPostgres(t *testing.T) (string, func()) {
	t.Helper()
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "postgres:17-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test",
			"POSTGRES_DB":       "order_db",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(60 * time.Second),
	}
	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	host, _ := c.Host(ctx)
	port, _ := c.MappedPort(ctx, "5432")
	dsn := fmt.Sprintf("host=%s port=%s user=test password=test dbname=order_db sslmode=disable", host, port.Port())

	return dsn, func() {
		_ = c.Terminate(ctx)
	}
}

func applySchema(t *testing.T, db *sql.DB) {
	t.Helper()
	schema := `
    CREATE EXTENSION IF NOT EXISTS pgcrypto;

    CREATE TABLE IF NOT EXISTS orders (
        id UUID PRIMARY KEY,
        user_id UUID NOT NULL,
        restaurant_id UUID NOT NULL,
        status VARCHAR(50) NOT NULL,
        delivery_address TEXT,
        total_price NUMERIC(10, 2) NOT NULL,
        promo_code VARCHAR(50),           -- Добавили колонку для промокода
        is_student BOOLEAN DEFAULT FALSE  -- Добавили флаг студента (на случай, если репо его пишет)
    );

    CREATE TABLE IF NOT EXISTS order_items (
        id UUID PRIMARY KEY,
        order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
        item_id UUID NOT NULL,
        name VARCHAR(255) NOT NULL,
        quantity INT NOT NULL,
        price NUMERIC(10, 2) NOT NULL
    );`

	_, err := db.ExecContext(context.Background(), schema)
	require.NoError(t, err)
}

func TestOrderRepo_CreateAndGet(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	dsn, stopPG := startPostgres(t)
	defer stopPG()

	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	defer db.Close()

	applySchema(t, db)

	repo := repository.NewOrderRepository(db)
	ctx := context.Background()

	orderID := "a0000000-0000-0000-0000-000000000001"
	userID := "b0000000-0000-0000-0000-000000000001"
	restaurantID := "c0000000-0000-0000-0000-000000000001"

	order := &domain.Order{
		ID:              orderID,
		UserID:          userID,
		RestaurantID:    restaurantID,
		Status:          domain.StatusPending,
		DeliveryAddress: "Astana, AITU, 5th floor",
		Items: []domain.OrderItem{
			{
				ID:       "d0000000-0000-0000-0000-000000000001",
				ItemID:   "e0000000-0000-0000-0000-000000000001",
				Name:     "Burger",
				Quantity: 2,
				Price:    1000,
			},
			{
				ID:       "d0000000-0000-0000-0000-000000000002",
				ItemID:   "e0000000-0000-0000-0000-000000000002",
				Name:     "Cola",
				Quantity: 1,
				Price:    500,
			},
		},
	}
	order.TotalPrice = order.CalcTotal()

	err = repo.Create(ctx, order)
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
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	dsn, stopPG := startPostgres(t)
	defer stopPG()

	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	defer db.Close()

	applySchema(t, db)

	repo := repository.NewOrderRepository(db)
	ctx := context.Background()

	orderID := "a0000000-0000-0000-0000-000000000002"
	userID := "b0000000-0000-0000-0000-000000000002"
	restaurantID := "c0000000-0000-0000-0000-000000000002"

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

	err = repo.UpdateStatus(ctx, orderID, domain.StatusConfirmed)
	require.NoError(t, err, "update status")

	got, err := repo.GetByID(ctx, orderID)
	require.NoError(t, err)
	assert.Equal(t, domain.StatusConfirmed, got.Status)
}
