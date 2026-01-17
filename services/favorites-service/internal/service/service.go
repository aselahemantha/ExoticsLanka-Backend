package service

import (
	"context"
	"time"

	"github.com/aselahemantha/exoticsLanka/services/favorites-service/internal/domain"
	"github.com/aselahemantha/exoticsLanka/services/favorites-service/internal/repository"
	"github.com/google/uuid"
)

type Service interface {
	AddFavorite(ctx context.Context, userID, listingID uuid.UUID) (*domain.Favorite, error)
	RemoveFavorite(ctx context.Context, userID, listingID uuid.UUID) error
	GetFavorites(ctx context.Context, userID uuid.UUID, page, limit int) ([]domain.Favorite, *domain.Pagination, error)
	CheckFavorite(ctx context.Context, userID, listingID uuid.UUID) (bool, *time.Time, error)
	GetFavoritesCount(ctx context.Context, userID uuid.UUID) (int64, error)
	ClearAllFavorites(ctx context.Context, userID uuid.UUID) (int64, error)
}

type service struct {
	repo repository.Repository
}

func NewService(repo repository.Repository) Service {
	return &service{repo: repo}
}

func (s *service) ClearAllFavorites(ctx context.Context, userID uuid.UUID) (int64, error) {
	return s.repo.ClearAllFavorites(ctx, userID)
}

func (s *service) AddFavorite(ctx context.Context, userID, listingID uuid.UUID) (*domain.Favorite, error) {
	// Business logic: check if already favorited (optional, as DB constraint captures it)
	// But let's rely on repo constraints for efficiency
	return s.repo.AddFavorite(ctx, userID, listingID)
}

func (s *service) RemoveFavorite(ctx context.Context, userID, listingID uuid.UUID) error {
	return s.repo.RemoveFavorite(ctx, userID, listingID)
}

func (s *service) GetFavorites(ctx context.Context, userID uuid.UUID, page, limit int) ([]domain.Favorite, *domain.Pagination, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	offset := (page - 1) * limit

	favorites, total, err := s.repo.GetFavorites(ctx, userID, limit, offset)
	if err != nil {
		return nil, nil, err
	}

	totalPages := 0
	if limit > 0 {
		totalPages = int((total + int64(limit) - 1) / int64(limit))
	}

	pagination := &domain.Pagination{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}

	return favorites, pagination, nil
}

func (s *service) CheckFavorite(ctx context.Context, userID, listingID uuid.UUID) (bool, *time.Time, error) {
	return s.repo.CheckFavorite(ctx, userID, listingID)
}

func (s *service) GetFavoritesCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	return s.repo.GetFavoritesCount(ctx, userID)
}
