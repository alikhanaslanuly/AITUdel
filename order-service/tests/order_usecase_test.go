package tests

import (
	"context"
	"order-service/internal/domain"
	"order-service/internal/usecase"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockOrderRepo struct {
	mock.Mock
}

func (m *MockOrderRepo) Create(ctx context.Context, order *domain.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockOrderRepo) GetByID(ctx context.Context, id string) (*domain.Order, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*domain.Order), args.Error(1)
}

func (m *MockOrderRepo) GetByUserID(ctx context.Context, userID string, page, limit int) ([]*domain.Order, int, error) {
	args := m.Called(ctx, userID, page, limit)
	return args.Get(0).([]*domain.Order), args.Int(1), args.Error(2)
}

func (m *MockOrderRepo) UpdateStatus(ctx context.Context, id, status string) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockOrderRepo) UpdatePromo(ctx context.Context, id, promoCode string, discount, newTotal float64) error {
	args := m.Called(ctx, id, promoCode, discount, newTotal)
	return args.Error(0)
}

type MockPromoRepo struct {
	mock.Mock
}

func (m *MockPromoRepo) GetByCode(ctx context.Context, code string) (*domain.PromoCode, error) {
	args := m.Called(ctx, code)
	return args.Get(0).(*domain.PromoCode), args.Error(1)
}

func (m *MockPromoRepo) IncrementUsage(ctx context.Context, promoID, userID, orderID string) error {
	args := m.Called(ctx, promoID, userID, orderID)
	return args.Error(0)
}

func (m *MockPromoRepo) HasUserUsed(ctx context.Context, promoID, userID string) (bool, error) {
	args := m.Called(ctx, promoID, userID)
	return args.Bool(0), args.Error(1)
}

func TestOrderDomain_CalcTotal(t *testing.T) {
	order := &domain.Order{
		Items: []domain.OrderItem{
			{Price: 1500, Quantity: 2},
			{Price: 800, Quantity: 1},
		},
	}
	assert.Equal(t, 3800.0, order.CalcTotal())
}

func TestOrderDomain_CanBeCancelled(t *testing.T) {
	cases := []struct {
		status   string
		expected bool
	}{
		{domain.StatusPending, true},
		{domain.StatusConfirmed, true},
		{domain.StatusCooking, false},
		{domain.StatusDelivered, false},
		{domain.StatusCancelled, false},
	}

	for _, c := range cases {
		order := &domain.Order{Status: c.status}
		assert.Equal(t, c.expected, order.CanBeCancelled(), "status: %s", c.status)
	}
}

func TestPromoCode_IsValid(t *testing.T) {
	promo := &domain.PromoCode{
		MaxUses:   10,
		UsedCount: 5,
	}
	assert.True(t, promo.IsValid())

	promo.UsedCount = 10
	assert.False(t, promo.IsValid())
}

func TestPromoCode_CalcDiscount(t *testing.T) {
	promo := &domain.PromoCode{DiscountPercent: 15}
	discount := promo.CalcDiscount(10000)
	assert.Equal(t, 1500.0, discount)
}

func TestCreateOrder_EmptyItems(t *testing.T) {
	orderRepo := new(MockOrderRepo)
	promoRepo := new(MockPromoRepo)

	uc := usecase.NewOrderUsecase(orderRepo, promoRepo, nil, nil)

	order := &domain.Order{
		UserID: "user-1",
		Items:  []domain.OrderItem{},
	}

	_, err := uc.CreateOrder(context.Background(), order, "", false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least one item")
}

func TestCancelOrder_WrongUser(t *testing.T) {
	orderRepo := new(MockOrderRepo)
	promoRepo := new(MockPromoRepo)

	orderRepo.On("GetByID", mock.Anything, "order-1").Return(&domain.Order{
		ID:     "order-1",
		UserID: "user-1",
		Status: domain.StatusPending,
	}, nil)

	uc := usecase.NewOrderUsecase(orderRepo, promoRepo, nil, nil)

	err := uc.CancelOrder(context.Background(), "order-1", "user-2")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "access denied")
}
