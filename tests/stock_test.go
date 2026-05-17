package tests

import (
	"context"
	"testing"

	"restaurant-service/internal/usecase"
)

func TestReserveStock(t *testing.T) {

	db := setupTestDB()

	stockUsecase := usecase.NewStockUsecase(
		db,
	)

	err := stockUsecase.ReserveStock(
		context.Background(),
		1,
		1,
	)

	if err != nil {
		t.Error(err)
	}
}
