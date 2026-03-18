package service

import (
	"bytes"
	"context"
	"fmt"
	"image/jpeg"
	"mime/multipart"

	"github.com/aselahemantha/exoticsLanka/services/image-service/internal/domain"
	"github.com/aselahemantha/exoticsLanka/services/image-service/internal/repository"
	"github.com/aselahemantha/exoticsLanka/services/image-service/internal/storage"
	"github.com/disintegration/imaging"
)

type Service struct {
	repo    *repository.Repository
	storage *storage.CloudinaryClient
}

func NewService(repo *repository.Repository, storage *storage.CloudinaryClient) *Service {
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

	optimisedImg := imaging.Fit(img, 1920, 1080, imaging.Lanczos)

	// Encode to JPEG
	buf := new(bytes.Buffer)
	err = jpeg.Encode(buf, optimisedImg, &jpeg.Options{Quality: 80})
	if err != nil {
		return nil, fmt.Errorf("failed to encode image: %w", err)
	}

	// Upload to Cloudinary
	folder := fmt.Sprintf("listings/%s", listingID)
	url, publicID, err := s.storage.UploadFile(ctx, buf, folder)
	if err != nil {
		return nil, err
	}

	// Determine usage
	nextOrder, err := s.repo.GetNextDisplayOrder(ctx, listingID)
	if err != nil {
		return nil, err
	}
	isPrimary := nextOrder == 1

	// Save to DB
	id, err := s.repo.CreateListingImage(ctx, listingID, url, publicID, isPrimary, nextOrder)
	if err != nil {
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

	// Remove from DB and get URL and Public ID
	_, publicID, err := s.repo.DeleteListingImage(ctx, imageID)
	if err != nil {
		return err
	}

	if publicID != "" {
		_ = s.storage.DeleteFile(ctx, publicID) // Best effort delete
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

	folder := fmt.Sprintf("avatars/%s", userID)
	url, publicID, err := s.storage.UploadFile(ctx, buf, folder)
	if err != nil {
		return "", err
	}

	err = s.repo.UpdateUserAvatar(ctx, userID, url, publicID)
	if err != nil {
		return "", err
	}

	return url, nil
}
