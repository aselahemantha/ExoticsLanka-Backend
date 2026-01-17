package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/aselahemantha/exoticsLanka/services/favorites-service/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	AddFavorite(ctx context.Context, userID, listingID uuid.UUID) (*domain.Favorite, error)
	RemoveFavorite(ctx context.Context, userID, listingID uuid.UUID) error
	GetFavorites(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.Favorite, int64, error)
	CheckFavorite(ctx context.Context, userID, listingID uuid.UUID) (bool, *time.Time, error)
	GetFavoritesCount(ctx context.Context, userID uuid.UUID) (int64, error)
	ClearAllFavorites(ctx context.Context, userID uuid.UUID) (int64, error)
}

type postgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) ClearAllFavorites(ctx context.Context, userID uuid.UUID) (int64, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	// 1. Get all listing IDs to decrement counts
	rows, err := tx.Query(ctx, "SELECT listing_id FROM favorites WHERE user_id = $1", userID)
	if err != nil {
		return 0, err
	}

	var listingIDs []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err == nil {
			listingIDs = append(listingIDs, id)
		}
	}
	rows.Close()

	if len(listingIDs) == 0 {
		return 0, nil
	}

	// 2. Delete favorites
	cmdTag, err := tx.Exec(ctx, "DELETE FROM favorites WHERE user_id = $1", userID)
	if err != nil {
		return 0, err
	}
	deletedCount := cmdTag.RowsAffected()

	// 3. Decrement listing counts
	// We can do this in a loop or a single update. A single update with ANY is more efficient but requires array support helper.
	// For simplicity with pgx and standard SQL, iterating is acceptable given the likely small number of favorites per user,
	// OR we can use the ANY($1) syntax if we convert slice to array.
	_, err = tx.Exec(ctx, "UPDATE car_listings SET favorites_count = GREATEST(favorites_count - 1, 0) WHERE id = ANY($1)", listingIDs)
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}

	return deletedCount, nil
}

func (r *postgresRepository) AddFavorite(ctx context.Context, userID, listingID uuid.UUID) (*domain.Favorite, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// 1. Check if listing exists and is active
	var status string
	err = tx.QueryRow(ctx, "SELECT status FROM car_listings WHERE id = $1", listingID).Scan(&status)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("listing not found")
		}
		return nil, err
	}

	if status != "active" {
		return nil, fmt.Errorf("cannot favorite inactive listing")
	}

	// 2. Insert Favorite
	fav := &domain.Favorite{}
	err = tx.QueryRow(ctx, `
		INSERT INTO favorites (user_id, listing_id)
		VALUES ($1, $2)
		RETURNING id, user_id, listing_id, created_at
	`, userID, listingID).Scan(&fav.ID, &fav.UserID, &fav.ListingID, &fav.CreatedAt)

	if err != nil {
		// Unique constraint violation check could be here, but handled by service logic too
		return nil, err
	}

	// 3. Update Listing count
	_, err = tx.Exec(ctx, "UPDATE car_listings SET favorites_count = favorites_count + 1 WHERE id = $1", listingID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return fav, nil
}

func (r *postgresRepository) RemoveFavorite(ctx context.Context, userID, listingID uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// 1. Delete Favorite
	cmdTag, err := tx.Exec(ctx, "DELETE FROM favorites WHERE user_id = $1 AND listing_id = $2", userID, listingID)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("favorite not found")
	}

	// 2. Decrement Listing count
	_, err = tx.Exec(ctx, "UPDATE car_listings SET favorites_count = GREATEST(favorites_count - 1, 0) WHERE id = $1", listingID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *postgresRepository) GetFavorites(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.Favorite, int64, error) {
	// Query for total count
	var total int64
	err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM favorites WHERE user_id = $1", userID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Query for favorites with listing details
	query := `
		SELECT 
			f.id, f.created_at, f.listing_id,
			cl.title, cl.make, cl.model, cl.year, cl.price, cl.mileage, cl.location, 
			cl.status, cl.health_score, cl.views, cl.days_listed,
			COALESCE((SELECT image_url FROM listing_images WHERE listing_id = cl.id AND is_cover = TRUE LIMIT 1), '') as cover_image
		FROM favorites f
		JOIN car_listings cl ON f.listing_id = cl.id
		WHERE f.user_id = $1
		ORDER BY f.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var favorites []domain.Favorite
	for rows.Next() {
		var fav domain.Favorite
		fav.UserID = userID // Explicitly set as it's input
		details := &domain.FavoriteListingDetails{}

		err := rows.Scan(
			&fav.ID, &fav.CreatedAt, &fav.ListingID,
			&details.Title, &details.Make, &details.Model, &details.Year, &details.Price, &details.Mileage, &details.Location,
			&details.Status, &details.HealthScore, &details.Views, &details.DaysListed,
			&details.CoverImage,
		)
		if err != nil {
			return nil, 0, err
		}

		details.ID = fav.ListingID
		fav.Listing = details
		favorites = append(favorites, fav)
	}

	return favorites, total, nil
}

func (r *postgresRepository) CheckFavorite(ctx context.Context, userID, listingID uuid.UUID) (bool, *time.Time, error) {
	var createdAt time.Time
	err := r.db.QueryRow(ctx, "SELECT created_at FROM favorites WHERE user_id = $1 AND listing_id = $2", userID, listingID).Scan(&createdAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil, nil
		}
		return false, nil, err
	}
	return true, &createdAt, nil
}

func (r *postgresRepository) GetFavoritesCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM favorites WHERE user_id = $1", userID).Scan(&count)
	return count, err
}
