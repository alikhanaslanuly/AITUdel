package usecase

import (
	"context"
	"fmt"
	"order-service/internal/domain"
	"order-service/internal/repository"
	"order-service/pkg/cache"
	"order-service/pkg/messaging"
	"time"

	"github.com/google/uuid"
)

type OrderEvent struct {
	OrderID      string  `json:"order_id"`
	UserID       string  `json:"user_id"`
	RestaurantID string  `json:"restaurant_id"`
	TotalPrice   float64 `json:"total_price"`
	Status       string  `json:"status"`
}

type OrderUsecase struct {
	orderRepo repository.OrderRepository
	promoRepo repository.PromoRepository
	cache     *cache.RedisClient
	nats      *messaging.NatsClient
}

func NewOrderUsecase(
	orderRepo repository.OrderRepository,
	promoRepo repository.PromoRepository,
	cache *cache.RedisClient,
	nats *messaging.NatsClient,
) *OrderUsecase {
	return &OrderUsecase{
		orderRepo: orderRepo,
		promoRepo: promoRepo,
		cache:     cache,
		nats:      nats,
	}
}

func (u *OrderUsecase) CreateOrder(ctx context.Context, req *domain.Order, promoCode string, isStudent bool) (*domain.Order, error) {
	if len(req.Items) == 0 {
		return nil, fmt.Errorf("order must have at least one item")
	}

	req.ID = uuid.New().String()
	req.Status = domain.StatusPending
	req.CreatedAt = time.Now()
	req.UpdatedAt = time.Now()

	for i := range req.Items {
		req.Items[i].ID = uuid.New().String()
	}

	req.TotalPrice = req.CalcTotal()

	if promoCode != "" {
		promo, err := u.promoRepo.GetByCode(ctx, promoCode)
		if err == nil && promo.IsValid() {
			if promo.StudentOnly && !isStudent {
				return nil, fmt.Errorf("promo code is for students only")
			}

			alreadyUsed, err := u.promoRepo.HasUserUsed(ctx, promo.ID, req.UserID)
			if err != nil {
				return nil, fmt.Errorf("check promo usage: %w", err)
			}
			if alreadyUsed {
				return nil, fmt.Errorf("you already used this promo code")
			}

			discount := promo.CalcDiscount(req.TotalPrice)
			req.DiscountAmount = discount
			req.TotalPrice = req.TotalPrice - discount
			req.PromoCode = promoCode

			defer func() {
				u.promoRepo.IncrementUsage(ctx, promo.ID, req.UserID, req.ID)
			}()
		}
	}

	if err := u.orderRepo.Create(ctx, req); err != nil {
		return nil, fmt.Errorf("create order: %w", err)
	}

	event := OrderEvent{
		OrderID:      req.ID,
		UserID:       req.UserID,
		RestaurantID: req.RestaurantID,
		TotalPrice:   req.TotalPrice,
		Status:       req.Status,
	}
	u.nats.Publish("order.created", event)

	u.cache.Delete(ctx, fmt.Sprintf("cart:%s", req.UserID))

	return req, nil
}

func (u *OrderUsecase) GetOrder(ctx context.Context, orderID string) (*domain.Order, error) {
	cacheKey := fmt.Sprintf("order:%s", orderID)

	var cached domain.Order
	if err := u.cache.Get(ctx, cacheKey, &cached); err == nil {
		return &cached, nil
	}

	order, err := u.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	u.cache.Set(ctx, cacheKey, order, 5*time.Minute)

	return order, nil
}

func (u *OrderUsecase) ListUserOrders(ctx context.Context, userID string, page, limit int) ([]*domain.Order, int, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 50 {
		limit = 10
	}
	return u.orderRepo.GetByUserID(ctx, userID, page, limit)
}

func (u *OrderUsecase) CancelOrder(ctx context.Context, orderID, userID string) error {
	order, err := u.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return err
	}

	if order.UserID != userID {
		return fmt.Errorf("access denied")
	}

	if !order.CanBeCancelled() {
		return fmt.Errorf("order cannot be cancelled in status: %s", order.Status)
	}

	if err := u.orderRepo.UpdateStatus(ctx, orderID, domain.StatusCancelled); err != nil {
		return fmt.Errorf("cancel order: %w", err)
	}

	u.cache.Delete(ctx, fmt.Sprintf("order:%s", orderID))

	event := OrderEvent{
		OrderID: orderID,
		UserID:  userID,
		Status:  domain.StatusCancelled,
	}
	u.nats.Publish("order.cancelled", event)

	return nil
}

func (u *OrderUsecase) ApplyPromo(ctx context.Context, orderID, userID, promoCode string, isStudent bool, orderTotal float64) (float64, float64, error) {
	promo, err := u.promoRepo.GetByCode(ctx, promoCode)
	if err != nil {
		return 0, 0, fmt.Errorf("promo code not found")
	}

	if !promo.IsValid() {
		return 0, 0, fmt.Errorf("promo code expired or exhausted")
	}

	if promo.StudentOnly && !isStudent {
		return 0, 0, fmt.Errorf("promo code is for students only")
	}

	alreadyUsed, err := u.promoRepo.HasUserUsed(ctx, promo.ID, userID)
	if err != nil {
		return 0, 0, err
	}
	if alreadyUsed {
		return 0, 0, fmt.Errorf("promo already used")
	}

	discount := promo.CalcDiscount(orderTotal)
	newTotal := orderTotal - discount

	if err := u.orderRepo.UpdatePromo(ctx, orderID, promoCode, discount, newTotal); err != nil {
		return 0, 0, err
	}

	u.promoRepo.IncrementUsage(ctx, promo.ID, userID, orderID)
	u.cache.Delete(ctx, fmt.Sprintf("order:%s", orderID))

	return discount, newTotal, nil
}
