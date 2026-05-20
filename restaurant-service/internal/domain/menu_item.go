package domain

type MenuItem struct {
	ID           int64
	RestaurantID int64
	Name         string
	Description  string
	Price        float64
	Stock        int
}
