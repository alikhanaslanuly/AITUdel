package domain

type Review struct {
	ID           int64
	RestaurantID int64
	UserID       int64
	Rating       int
	Comment      string
}
