package domain

type Review struct {
	ID           int64
	RestaurantID int64
	UserID       string
	Rating       int
	Comment      string
}
