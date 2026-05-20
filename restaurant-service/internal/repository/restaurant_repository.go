package repository

import (
	"context"
	"restaurant-service/internal/domain"
)

type RestaurantRepository interface {
	ListRestaurants(
		ctx context.Context,
	) ([]domain.Restaurant, error)

	GetRestaurant(
		ctx context.Context,
		id int64,
	) (*domain.Restaurant, error)

	SearchItems(
		ctx context.Context,
		query string,
	) ([]domain.MenuItem, error)

	GetItem(
		ctx context.Context,
		id int64,
	) (*domain.MenuItem, error)

	CreateReview(
		ctx context.Context,
		review domain.Review,
	) error

	UpdateRestaurantRating(
		ctx context.Context,
		restaurantID int64,
	) error
}
