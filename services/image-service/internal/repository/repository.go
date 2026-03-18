package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateListingImage(ctx context.Context, listingID string, url string, publicID string, isPrimary bool, position int) (string, error) {
	var id string
	query := `
		INSERT INTO listing_images (listing_id, url, public_id, is_primary, position)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	err := r.db.QueryRow(ctx, query, listingID, url, publicID, isPrimary, position).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("failed to insert listing image: %w", err)
	}
	return id, nil
}

func (r *Repository) GetListingImages(ctx context.Context, listingID string) ([]struct {
	ID        string
	URL       string
	PublicID  string
	IsPrimary bool
}, error) {
	query := `
		SELECT id, url, public_id, is_primary
		FROM listing_images
		WHERE listing_id = $1
		ORDER BY position ASC
	`
	rows, err := r.db.Query(ctx, query, listingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []struct {
		ID        string
		URL       string
		PublicID  string
		IsPrimary bool
	}

	for rows.Next() {
		var img struct {
			ID        string
			URL       string
			PublicID  string
			IsPrimary bool
		}
		if err := rows.Scan(&img.ID, &img.URL, &img.PublicID, &img.IsPrimary); err != nil {
			return nil, err
		}
		images = append(images, img)
	}
	return images, nil
}

func (r *Repository) DeleteListingImage(ctx context.Context, imageID string) (string, string, error) {
	var url, publicID string
	query := `
		DELETE FROM listing_images
		WHERE id = $1
		RETURNING url, public_id
	`
	err := r.db.QueryRow(ctx, query, imageID).Scan(&url, &publicID)
	if err != nil {
		return "", "", fmt.Errorf("failed to delete listing image: %w", err)
	}
	return url, publicID, nil
}

func (r *Repository) GetImageOwner(ctx context.Context, imageID string) (string, error) {
	var userID string
	query := `
		SELECT l.user_id
		FROM listing_images li
		JOIN car_listings l ON li.listing_id = l.id
		WHERE li.id = $1
	`
	err := r.db.QueryRow(ctx, query, imageID).Scan(&userID)
	if err != nil {
		return "", fmt.Errorf("failed to get image owner: %w", err)
	}
	return userID, nil
}

func (r *Repository) GetListingOwner(ctx context.Context, listingID string) (string, error) {
	var userID string
	query := `SELECT user_id FROM car_listings WHERE id = $1`
	err := r.db.QueryRow(ctx, query, listingID).Scan(&userID)
	if err != nil {
		return "", fmt.Errorf("failed to get listing owner: %w", err)
	}
	return userID, nil
}

func (r *Repository) GetUserAvatarPublicID(ctx context.Context, userID string) (string, error) {
	var publicID *string
	query := `SELECT avatar_public_id FROM users WHERE id = $1`
	err := r.db.QueryRow(ctx, query, userID).Scan(&publicID)
	if err != nil {
		return "", err
	}
	if publicID == nil {
		return "", nil
	}
	return *publicID, nil
}

func (r *Repository) UpdateUserAvatar(ctx context.Context, userID string, url string, publicID string) error {
	query := `UPDATE users SET avatar_url = $1, avatar_public_id = $2 WHERE id = $3`
	_, err := r.db.Exec(ctx, query, url, publicID, userID)
	if err != nil {
		return fmt.Errorf("failed to update user avatar: %w", err)
	}
	return nil
}

func (r *Repository) ReorderListingImages(ctx context.Context, imageIDs []string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for i, id := range imageIDs {
		// Set display_order = i + 1
		// Also ensure primary is set correctly? usually reordering might affect primary if the first one implies primary.
		// For now, just update order.
		_, err := tx.Exec(ctx, `UPDATE listing_images SET position = $1 WHERE id = $2`, i+1, id)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *Repository) SetCoverImage(ctx context.Context, listingID, imageID string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Unset primary for all images of this listing
	_, err = tx.Exec(ctx, `UPDATE listing_images SET is_primary = false WHERE listing_id = $1`, listingID)
	if err != nil {
		return err
	}

	// Set primary for the specific image
	_, err = tx.Exec(ctx, `UPDATE listing_images SET is_primary = true WHERE id = $1`, imageID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *Repository) GetNextDisplayOrder(ctx context.Context, listingID string) (int, error) {
	var maxOrder *int
	err := r.db.QueryRow(ctx, `SELECT MAX(position) FROM listing_images WHERE listing_id=$1`, listingID).Scan(&maxOrder)
	if err != nil {
		// If no rows, it might return null/error depending on driver. pgx usually returns nil for null scan if destination is pointer
		// but Scan might return error if no rows? No, MAX returns a row with NULL if empty.
		// If error is sql.ErrNoRows, return 1.
		// Actually pgx handles it.
		// If table is empty MAX is NULL. `maxOrder` will be nil.
	}
	if maxOrder == nil {
		return 1, nil
	}
	return *maxOrder + 1, nil
}
