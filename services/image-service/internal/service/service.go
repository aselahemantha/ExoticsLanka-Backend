package service

import (
	"bytes"
	"context"
	"fmt"
	"image/jpeg"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/aselahemantha/exoticsLanka/services/image-service/internal/domain"
	"github.com/aselahemantha/exoticsLanka/services/image-service/internal/repository"
	"github.com/aselahemantha/exoticsLanka/services/image-service/internal/storage"
	"github.com/disintegration/imaging"
	"github.com/google/uuid"
)

type Service struct {
	repo    *repository.Repository
	storage *storage.S3Client
}

func NewService(repo *repository.Repository, storage *storage.S3Client) *Service {
	return &Service{
		repo:    repo,
		storage: storage,
	}
}

func (s *Service) UploadListingImage(ctx context.Context, listingID string, userID string, file multipart.File, header *multipart.FileHeader) (*domain.ImageUploadResponse, error) {
	// Verify ownership
	ownerID, err := s.repo.GetListingOwner(ctx, listingID)
	if err != nil {
		return nil, err
	}
	if ownerID != userID {
		return nil, fmt.Errorf("unauthorized: user does not own this listing")
	}

	// Decode image
	img, err := imaging.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Resize logic (Generate a high-res version for display, maybe optimized)
	// For MVP, strictly max width of 1920, preserve aspect ratio
	optimisedImg := imaging.Fit(img, 1920, 1080, imaging.Lanczos)

	// Encode to JPEG
	buf := new(bytes.Buffer)
	err = jpeg.Encode(buf, optimisedImg, &jpeg.Options{Quality: 80})
	if err != nil {
		return nil, fmt.Errorf("failed to encode image: %w", err)
	}

	// Generate Key
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext == "" {
		ext = ".jpg"
	}
	key := fmt.Sprintf("listings/%s/%s%s", listingID, uuid.New().String(), ".jpg") // Force jpg for consistency since we encoded to jpeg

	// Upload to S3
	url, err := s.storage.UploadFile(key, buf, "image/jpeg")
	if err != nil {
		return nil, err
	}

	// Determine usage
	// Check if this is the first image, if so make it primary
	nextOrder, err := s.repo.GetNextDisplayOrder(ctx, listingID)
	if err != nil {
		return nil, err
	}
	isPrimary := nextOrder == 1

	// Save to DB
	id, err := s.repo.CreateListingImage(ctx, listingID, url, isPrimary, nextOrder)
	if err != nil {
		// Cleanup S3 if DB fails?
		// For now, ignore cleanup to keep it simple, cleaner job can run later.
		return nil, err
	}

	return &domain.ImageUploadResponse{
		ID:        id,
		URL:       url,
		IsPrimary: isPrimary,
	}, nil
}

func (s *Service) DeleteListingImage(ctx context.Context, imageID string, userID string) error {
	// Verify ownership
	ownerID, err := s.repo.GetImageOwner(ctx, imageID)
	if err != nil {
		return err
	}
	if ownerID != userID {
		return fmt.Errorf("unauthorized")
	}

	// Get URL to extract key
	// We need to fetch the URL first to know what to delete from S3
	// Retrieve image details from DB is tricky if we don't have a Get method for single image
	// But DeleteListingImage returns the URL in our repo implementation!
	url, err := s.repo.DeleteListingImage(ctx, imageID)
	if err != nil {
		return err
	}

	// Extract key from URL
	// Simple assumption: URL ends with the key
	// This might be brittle if URL structure changes.
	// A better way is to store the key or assume standard structure.
	// URL: https://bucket.s3.../listings/listingID/uuid.jpg
	// Key: listings/listingID/uuid.jpg

	// Quick hack for MVP: split by bucket name or known domain parts.
	// Or just store the key in DB? For now, let's try to parse it.
	// If standard AWS URL path style: https://s3.region.amazonaws.com/bucket/key
	// If virtual hosted: https://bucket.s3.region.amazonaws.com/key

	// Let's rely on the fact that we constructed it as `listings/...` which is unique enough?
	// Finding "listings/" index
	idx := strings.Index(url, "listings/")
	if idx != -1 {
		key := url[idx:]
		_ = s.storage.DeleteFile(key) // Ignore error on S3 delete for now
	}

	return nil
}

func (s *Service) ReorderImages(ctx context.Context, listingID string, userID string, imageIDs []string) error {
	ownerID, err := s.repo.GetListingOwner(ctx, listingID)
	if err != nil {
		return err
	}
	if ownerID != userID {
		return fmt.Errorf("unauthorized")
	}

	return s.repo.ReorderListingImages(ctx, imageIDs)
}

func (s *Service) UploadUserAvatar(ctx context.Context, userID string, file multipart.File, header *multipart.FileHeader) (string, error) {
	img, err := imaging.Decode(file)
	if err != nil {
		return "", err
	}

	// Resize to 400x400 square
	thumb := imaging.Fill(img, 400, 400, imaging.Center, imaging.Lanczos)

	buf := new(bytes.Buffer)
	err = jpeg.Encode(buf, thumb, &jpeg.Options{Quality: 80})
	if err != nil {
		return "", err
	}

	key := fmt.Sprintf("avatars/%s/%s.jpg", userID, uuid.New().String())
	url, err := s.storage.UploadFile(key, buf, "image/jpeg")
	if err != nil {
		return "", err
	}

	err = s.repo.UpdateUserAvatar(ctx, userID, url)
	if err != nil {
		return "", err
	}

	return url, nil
}
