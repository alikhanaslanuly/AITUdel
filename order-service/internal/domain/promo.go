package domain

import "time"

type PromoCode struct {
	ID              string
	Code            string
	DiscountPercent int
	StudentOnly     bool
	MaxUses         int
	UsedCount       int
	ExpiresAt       *time.Time
	CreatedAt       time.Time
}

type PromoUsage struct {
	ID      string
	PromoID string
	UserID  string
	OrderID string
	UsedAt  time.Time
}

func (p *PromoCode) IsValid() bool {
	if p.UsedCount >= p.MaxUses {
		return false
	}
	if p.ExpiresAt != nil && time.Now().After(*p.ExpiresAt) {
		return false
	}
	return true
}

func (p *PromoCode) CalcDiscount(total float64) float64 {
	return total * float64(p.DiscountPercent) / 100
}
