package repository

import (
	"context"
	"database/sql"
	"restaurant-service/internal/domain"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{
		db: db,
	}
}

func (r *PostgresRepository) ListRestaurants(
	ctx context.Context,
) ([]domain.Restaurant, error) {

	query := `
		SELECT id, name, description, rating
		FROM restaurants
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var restaurants []domain.Restaurant

	for rows.Next() {

		var restaurant domain.Restaurant

		err := rows.Scan(
			&restaurant.ID,
			&restaurant.Name,
			&restaurant.Description,
			&restaurant.Rating,
		)

		if err != nil {
			return nil, err
		}

		restaurants = append(restaurants, restaurant)
	}

	return restaurants, nil
}

func (r *PostgresRepository) GetRestaurant(
	ctx context.Context,
	id int64,
) (*domain.Restaurant, error) {

	query := `
		SELECT id, name, description, rating
		FROM restaurants
		WHERE id = $1
	`

	var restaurant domain.Restaurant

	err := r.db.QueryRowContext(
		ctx,
		query,
		id,
	).Scan(
		&restaurant.ID,
		&restaurant.Name,
		&restaurant.Description,
		&restaurant.Rating,
	)

	if err != nil {
		return nil, err
	}

	return &restaurant, nil
}

func (r *PostgresRepository) SearchItems(
	ctx context.Context,
	search string,
) ([]domain.MenuItem, error) {

	query := `
		SELECT id, name, description, price, stock
		FROM menu_items
		WHERE name ILIKE '%' || $1 || '%'
	`

	rows, err := r.db.QueryContext(
		ctx,
		query,
		search,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var items []domain.MenuItem

	for rows.Next() {

		var item domain.MenuItem

		err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.Description,
			&item.Price,
			&item.Stock,
		)

		if err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	return items, nil
}

func (r *PostgresRepository) GetItem(
	ctx context.Context,
	id int64,
) (*domain.MenuItem, error) {

	query := `
		SELECT id, name, description, price, stock
		FROM menu_items
		WHERE id = $1
	`

	var item domain.MenuItem

	err := r.db.QueryRowContext(
		ctx,
		query,
		id,
	).Scan(
		&item.ID,
		&item.Name,
		&item.Description,
		&item.Price,
		&item.Stock,
	)

	if err != nil {
		return nil, err
	}

	return &item, nil
}

func (r *PostgresRepository) CreateReview(
	ctx context.Context,
	review domain.Review,
) error {

	query := `
		INSERT INTO reviews (
			restaurant_id,
			user_id,
			rating,
			comment
		)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		review.RestaurantID,
		review.UserID,
		review.Rating,
		review.Comment,
	)

	return err
}

func (r *PostgresRepository) UpdateRestaurantRating(
	ctx context.Context,
	restaurantID int64,
) error {

	var avgRating float64

	query := `
		SELECT AVG(rating)
		FROM reviews
		WHERE restaurant_id = $1
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		restaurantID,
	).Scan(&avgRating)

	if err != nil {
		return err
	}

	updateQuery := `
		UPDATE restaurants
		SET rating = $1
		WHERE id = $2
	`

	_, err = r.db.ExecContext(
		ctx,
		updateQuery,
		avgRating,
		restaurantID,
	)

	return err
}
