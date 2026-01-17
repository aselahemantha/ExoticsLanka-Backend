package service

import (
	"context"
	"errors"
	"time"

	"github.com/aselahemantha/exoticsLanka/services/listings-service/internal/domain"
	"github.com/aselahemantha/exoticsLanka/services/listings-service/internal/repository"
	"github.com/google/uuid"
)

// Service defines the interface for business logic
type Service interface {
	CreateListing(ctx context.Context, req *domain.CreateListingRequest, userID uuid.UUID) (*domain.CarListing, error)
	GetListing(ctx context.Context, id uuid.UUID) (*domain.CarListing, error)
	GetListings(ctx context.Context, filter repository.ListingFilter) ([]*domain.CarListing, int, error)

	// Methods for dashboard/home
	GetFeaturedListings(ctx context.Context) ([]*domain.CarListing, error)
	GetTrendingListings(ctx context.Context) ([]*domain.CarListing, error)
	GetBrands(ctx context.Context) ([]*domain.CarBrand, error)
}

type service struct {
	repo repository.Repository
}

func NewService(repo repository.Repository) Service {
	return &service{repo: repo}
}

func (s *service) CreateListing(ctx context.Context, req *domain.CreateListingRequest, userID uuid.UUID) (*domain.CarListing, error) {
	// Map request to domain model
	listing := &domain.CarListing{
		ID:           uuid.New(),
		UserID:       userID,
		Title:        req.Title,
		Make:         req.Make,
		Model:        req.Model,
		Year:         req.Year,
		Price:        req.Price,
		Mileage:      req.Mileage,
		Condition:    req.Condition,
		Transmission: req.Transmission,
		FuelType:     req.FuelType,
		BodyType:     req.BodyType,
		Color:        req.Color,
		Doors:        req.Doors,
		Seats:        req.Seats,
		EngineSize:   req.EngineSize,
		Drivetrain:   req.Drivetrain,
		Description:  req.Description,
		Location:     req.Location,
		ContactPhone: req.ContactPhone,
		ContactEmail: req.ContactEmail,
		Status:       domain.StatusPending,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Calculate initial health score
	listing.HealthScore = calculateHealthScore(listing, len(req.Features), 0) // 0 images initially

	// Features
	for _, fName := range req.Features {
		listing.Features = append(listing.Features, domain.ListingFeature{
			FeatureName: fName,
		})
	}

	if err := s.repo.CreateListing(ctx, listing); err != nil {
		return nil, err
	}

	return listing, nil
}

func (s *service) GetListing(ctx context.Context, id uuid.UUID) (*domain.CarListing, error) {
	listing, err := s.repo.GetListingByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if listing == nil {
		return nil, errors.New("listing not found")
	}

	// Increment view count asynchronously
	// In production, push to a queue or use a separate goroutine
	go func() {
		// Create a background context
		_ = s.repo.IncrementViews(context.Background(), id)
	}()

	return listing, nil
}

func (s *service) GetListings(ctx context.Context, filter repository.ListingFilter) ([]*domain.CarListing, int, error) {
	return s.repo.GetListings(ctx, filter)
}

func (s *service) GetFeaturedListings(ctx context.Context) ([]*domain.CarListing, error) {
	return s.repo.GetFeaturedListings(ctx, 10)
}

func (s *service) GetTrendingListings(ctx context.Context) ([]*domain.CarListing, error) {
	return s.repo.GetTrendingListings(ctx, 10)
}

func (s *service) GetBrands(ctx context.Context) ([]*domain.CarBrand, error) {
	return s.repo.GetBrands(ctx)
}

// Logic helpers

func calculateHealthScore(l *domain.CarListing, featureCount, imageCount int) int {
	score := 50

	// Images (+20 max)
	score += min(imageCount*4, 20)

	// Description length (+10 max)
	descLen := 0
	if l.Description != nil {
		descLen = len(*l.Description)
	}
	if descLen > 200 {
		score += 10
	} else if descLen > 100 {
		score += 5
	}

	// Features (+10 max)
	score += min(featureCount, 10)

	// Completeness (+10)
	optionalFields := 0
	if l.Transmission != nil {
		optionalFields++
	}
	if l.FuelType != nil {
		optionalFields++
	}
	if l.Color != nil {
		optionalFields++
	}
	if l.BodyType != nil {
		optionalFields++
	}
	if l.Doors != nil {
		optionalFields++
	}
	if l.Seats != nil {
		optionalFields++
	}

	score += int((float64(optionalFields) / 6.0) * 10)

	if score > 100 {
		return 100
	}
	return score
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
