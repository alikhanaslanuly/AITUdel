package usecase

import (
	"context"
	"fmt"
	"order-service/internal/domain"
	"order-service/pkg/cache"
	"time"
)

type CartUsecase struct {
	cache *cache.RedisClient
}

func NewCartUsecase(cache *cache.RedisClient) *CartUsecase {
	return &CartUsecase{cache: cache}
}

func (u *CartUsecase) AddItem(ctx context.Context, userID string, item domain.CartItem) error {
	key := fmt.Sprintf("cart:%s", userID)

	var items []domain.CartItem
	u.cache.Get(ctx, key, &items)

	found := false
	for i, existing := range items {
		if existing.ItemID == item.ItemID {
			items[i].Quantity += item.Quantity
			found = true
			break
		}
	}
	if !found {
		items = append(items, item)
	}

	return u.cache.Set(ctx, key, items, 30*time.Minute)
}

func (u *CartUsecase) GetCart(ctx context.Context, userID string) ([]domain.CartItem, float64, error) {
	key := fmt.Sprintf("cart:%s", userID)

	var items []domain.CartItem
	if err := u.cache.Get(ctx, key, &items); err != nil {
		return []domain.CartItem{}, 0, nil
	}

	total := 0.0
	for _, item := range items {
		total += item.Price * float64(item.Quantity)
	}

	return items, total, nil
}

func (u *CartUsecase) ClearCart(ctx context.Context, userID string) error {
	return u.cache.Delete(ctx, fmt.Sprintf("cart:%s", userID))
}
