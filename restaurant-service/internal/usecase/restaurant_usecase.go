package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"restaurant-service/internal/domain"
	"restaurant-service/internal/repository"
)

type RestaurantUsecase struct {
	repo  repository.RestaurantRepository
	redis *redis.Client
}

func NewRestaurantUsecase(
	repo repository.RestaurantRepository,
	redis *redis.Client,
) *RestaurantUsecase {

	return &RestaurantUsecase{
		repo:  repo,
		redis: redis,
	}
}

func (u *RestaurantUsecase) ListRestaurants(
	ctx context.Context,
) ([]domain.Restaurant, error) {

	return u.repo.ListRestaurants(ctx)
}

func (u *RestaurantUsecase) GetRestaurant(
	ctx context.Context,
	id int64,
) (*domain.Restaurant, error) {

	cacheKey := fmt.Sprintf(
		"restaurant:%d",
		id,
	)

	cached, err := u.redis.Get(
		ctx,
		cacheKey,
	).Result()

	if err == nil {

		var restaurant domain.Restaurant

		json.Unmarshal(
			[]byte(cached),
			&restaurant,
		)

		fmt.Println("CACHE HIT")

		return &restaurant, nil
	}

	fmt.Println("CACHE MISS")

	restaurant, err := u.repo.GetRestaurant(
		ctx,
		id,
	)

	if err != nil {
		return nil, err
	}

	jsonData, _ := json.Marshal(restaurant)

	u.redis.Set(
		ctx,
		cacheKey,
		jsonData,
		5*time.Minute,
	)

	return restaurant, nil
}

func (u *RestaurantUsecase) SearchItems(
	ctx context.Context,
	query string,
) ([]domain.MenuItem, error) {

	return u.repo.SearchItems(
		ctx,
		query,
	)
}

func (u *RestaurantUsecase) GetItem(
	ctx context.Context,
	id int64,
) (*domain.MenuItem, error) {

	return u.repo.GetItem(
		ctx,
		id,
	)
}

func (u *RestaurantUsecase) RateRestaurant(
	ctx context.Context,
	review domain.Review,
) error {

	err := u.repo.CreateReview(
		ctx,
		review,
	)

	if err != nil {
		return err
	}

	err = u.repo.UpdateRestaurantRating(
		ctx,
		review.RestaurantID,
	)

	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf(
		"restaurant:%d",
		review.RestaurantID,
	)

	u.redis.Del(
		ctx,
		cacheKey,
	)

	return nil
}
