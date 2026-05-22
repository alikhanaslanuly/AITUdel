package domain

import (
	"time"
)

const (
	StatusPending   = "pending"
	StatusConfirmed = "confirmed"
	StatusCooking   = "cooking"
	StatusOnTheWay  = "on_the_way"
	StatusDelivered = "delivered"
	StatusCancelled = "cancelled"
)

type Order struct {
	ID              string
	UserID          string
	RestaurantID    string
	Status          string
	TotalPrice      float64
	PromoCode       string
	DiscountAmount  float64
	DeliveryAddress string
	CourierID       string
	Items           []OrderItem
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type OrderItem struct {
	ID       string
	OrderID  string
	ItemID   string
	Name     string
	Quantity int
	Price    float64
}

type CartItem struct {
	ID        string
	UserID    string
	ItemID    string
	Name      string
	Quantity  int
	Price     float64
	CreatedAt time.Time
}

func (o *Order) CanBeCancelled() bool {
	return o.Status == StatusPending || o.Status == StatusConfirmed
}

func (o *Order) CalcTotal() float64 {
	total := 0.0
	for _, item := range o.Items {
		total += item.Price * float64(item.Quantity)
	}
	return total
}
