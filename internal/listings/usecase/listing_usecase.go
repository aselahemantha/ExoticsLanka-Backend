package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/aselahemantha/exoticsLanka/internal/listings/domain"
	"github.com/google/uuid"
)

type listingUseCase struct {
	repo domain.ListingRepository
}

func NewListingUseCase(repo domain.ListingRepository) domain.ListingUseCase {
	return &listingUseCase{repo: repo}
}

func (u *listingUseCase) CreateListing(ctx context.Context, userID uuid.UUID, req *domain.CreateListingRequest) (*domain.CarListing, error) {
	featuresJSON, _ := json.Marshal(req.Features)
	if string(featuresJSON) == "null" {
		featuresJSON = []byte("[]")
	}

	status := domain.StatusActive
	if req.Status != nil {
		status = *req.Status
	}

	l := &domain.CarListing{
		ID:           uuid.New(),
		SellerID:     userID,
		Title:        req.Title,
		Make:         req.Make,
		Model:        req.Model,
		Year:         req.Year,
		Price:        req.Price,
		Mileage:      req.Mileage,
		Condition:    req.Condition,
		BodyType:     req.BodyType,
		FuelType:     req.FuelType,
		Transmission: req.Transmission,
		Color:        req.Color,
		Doors:        req.Doors,
		Seats:        req.Seats,
		EngineSize:   req.EngineSize,
		Drivetrain:   req.Drivetrain,
		Features:     featuresJSON,
		Description:  req.Description,
		Location:     req.Location,
		Status:       status,
		IsFeatured:   false,
		IsTrending:   false,
		Views:        0,
		Favorites:    0,
		HealthScore:  calculateHealthScore(req, len(req.Images)),
		Verified:     false,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := u.repo.Create(ctx, l, req.Images); err != nil {
		return nil, err
	}

	return u.repo.GetByID(ctx, l.ID)
}

func (u *listingUseCase) GetListing(ctx context.Context, id uuid.UUID) (*domain.CarListing, error) {
	l, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if l == nil {
		return nil, errors.New("listing not found")
	}

	// Fire and forget view increment
	go func() {
		_ = u.repo.IncrementViews(context.Background(), id)
	}()

	return l, nil
}

func (u *listingUseCase) GetListings(ctx context.Context, filter domain.ListingFilter) (*domain.PaginatedListingsResponse, error) {
	listings, total, err := u.repo.GetListings(ctx, filter)
	if err != nil {
		return nil, err
	}

	limit := filter.Limit
	if limit == 0 {
		limit = 20
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}

	totalPages := (total + limit - 1) / limit

	return &domain.PaginatedListingsResponse{
		Listings:   listings,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

func (u *listingUseCase) UpdateListing(ctx context.Context, userID, listingID uuid.UUID, req *domain.UpdateListingRequest) (*domain.CarListing, error) {
	l, err := u.repo.GetByID(ctx, listingID)
	if err != nil {
		return nil, err
	}
	if l == nil {
		return nil, errors.New("listing not found")
	}

	if l.SellerID != userID {
		return nil, errors.New("unauthorized to update this listing")
	}

	if req.Title != nil {
		l.Title = *req.Title
	}
	if req.Make != nil {
		l.Make = *req.Make
	}
	if req.Model != nil {
		l.Model = *req.Model
	}
	if req.Year != nil {
		l.Year = *req.Year
	}
	if req.Price != nil {
		l.Price = *req.Price
	}
	if req.Mileage != nil {
		l.Mileage = *req.Mileage
	}
	if req.Condition != nil {
		l.Condition = *req.Condition
	}
	if req.BodyType != nil {
		l.BodyType = req.BodyType
	}
	if req.FuelType != nil {
		l.FuelType = req.FuelType
	}
	if req.Transmission != nil {
		l.Transmission = req.Transmission
	}
	if req.Color != nil {
		l.Color = req.Color
	}
	if req.Doors != nil {
		l.Doors = req.Doors
	}
	if req.Seats != nil {
		l.Seats = req.Seats
	}
	if req.EngineSize != nil {
		l.EngineSize = req.EngineSize
	}
	if req.Drivetrain != nil {
		l.Drivetrain = req.Drivetrain
	}
	if req.Description != nil {
		l.Description = req.Description
	}
	if req.Location != nil {
		l.Location = req.Location
	}
	if req.Status != nil {
		l.Status = *req.Status
	}

	if req.Features != nil {
		featuresJSON, _ := json.Marshal(req.Features)
		l.Features = featuresJSON
	}

	if err := u.repo.Update(ctx, l); err != nil {
		return nil, err
	}

	return u.repo.GetByID(ctx, l.ID)
}

func (u *listingUseCase) DeleteListing(ctx context.Context, userID, listingID uuid.UUID) error {
	l, err := u.repo.GetByID(ctx, listingID)
	if err != nil {
		return err
	}
	if l == nil {
		return errors.New("listing not found")
	}

	if l.SellerID != userID {
		return errors.New("unauthorized to delete this listing")
	}

	return u.repo.Delete(ctx, listingID)
}

// ----------------------------------------------------
// Discovery
// ----------------------------------------------------

func (u *listingUseCase) GetFeaturedListings(ctx context.Context) ([]*domain.CarListing, error) {
	return u.repo.GetFeatured(ctx, 10)
}

func (u *listingUseCase) GetTrendingListings(ctx context.Context) ([]*domain.CarListing, error) {
	return u.repo.GetTrending(ctx, 10)
}

func (u *listingUseCase) GetBrands(ctx context.Context) ([]string, error) {
	return u.repo.GetBrands(ctx)
}

func (u *listingUseCase) GetModelsByBrand(ctx context.Context, brand string) ([]string, error) {
	return u.repo.GetModels(ctx, brand)
}

// ----------------------------------------------------
// Interactions
// ----------------------------------------------------

func (u *listingUseCase) IncrementView(ctx context.Context, listingID uuid.UUID) error {
	return u.repo.IncrementViews(ctx, listingID)
}

func (u *listingUseCase) ToggleFavorite(ctx context.Context, userID, listingID uuid.UUID, isFavorite bool) error {
	if isFavorite {
		return u.repo.AddFavorite(ctx, userID, listingID)
	}
	return u.repo.RemoveFavorite(ctx, userID, listingID)
}

func (u *listingUseCase) GetUserFavorites(ctx context.Context, userID uuid.UUID) ([]*domain.CarListing, error) {
	return u.repo.GetUserFavorites(ctx, userID)
}

func (u *listingUseCase) ReportListing(ctx context.Context, userID, listingID uuid.UUID, req *domain.ReportListingRequest) error {
	l, err := u.repo.GetByID(ctx, listingID)
	if err != nil {
		return err
	}
	if l == nil {
		return errors.New("listing not found")
	}

	report := &domain.ListingReport{
		ID:         uuid.New(),
		ListingID:  listingID,
		ReporterID: userID,
		Reason:     req.Reason,
		Details:    req.Details,
		Status:     domain.ReportPending,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	return u.repo.CreateReport(ctx, report)
}

// ----------------------------------------------------
// Helpers
// ----------------------------------------------------

func calculateHealthScore(req *domain.CreateListingRequest, imageCount int) int {
	score := 50
	score += min(imageCount*5, 20)

	descLen := 0
	if req.Description != nil {
		descLen = len(*req.Description)
	}

	if descLen > 200 {
		score += 10
	} else if descLen > 50 {
		score += 5
	}

	score += min(len(req.Features), 10)

	optionalFields := 0
	if req.Transmission != nil {
		optionalFields++
	}
	if req.FuelType != nil {
		optionalFields++
	}
	if req.Color != nil {
		optionalFields++
	}
	if req.BodyType != nil {
		optionalFields++
	}
	if req.Doors != nil {
		optionalFields++
	}
	if req.Seats != nil {
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
