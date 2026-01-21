package service

import (
	"context"
	"fmt"

	"github.com/aselahemantha/exoticsLanka/services/saved-searches-service/internal/domain"
	"github.com/aselahemantha/exoticsLanka/services/saved-searches-service/internal/repository"
	"github.com/google/uuid"
)

type Service interface {
	CreateSavedSearch(ctx context.Context, req domain.CreateSavedSearchRequest, userID uuid.UUID) (*domain.SavedSearch, error)
	GetSavedSearch(ctx context.Context, id, userID uuid.UUID) (*domain.SavedSearch, error)
	GetUserSavedSearches(ctx context.Context, userID uuid.UUID) ([]domain.SavedSearch, map[string]int, error) // Returns saved searches + total new matches
	UpdateSavedSearch(ctx context.Context, id, userID uuid.UUID, req domain.UpdateSavedSearchRequest) (*domain.SavedSearch, error)
	DeleteSavedSearch(ctx context.Context, id, userID uuid.UUID) error

	CheckForNewMatches(ctx context.Context, id, userID uuid.UUID) (*domain.CheckMatchesResponse, error)
	RunSearch(ctx context.Context, id, userID uuid.UUID, page, limit int) (*domain.RunSearchResponse, error)

	UpdateAlertSettings(ctx context.Context, id, userID uuid.UUID, req domain.UpdateAlertsRequest) error
	GetTotalNewMatches(ctx context.Context, userID uuid.UUID) (*domain.NewMatchesResponse, error)
}

type service struct {
	repo repository.Repository
}

func NewService(repo repository.Repository) Service {
	return &service{repo: repo}
}

func (s *service) CreateSavedSearch(ctx context.Context, req domain.CreateSavedSearchRequest, userID uuid.UUID) (*domain.SavedSearch, error) {
	// 1. Calculate current matches
	count, err := s.repo.CountMatchingListings(ctx, req.Filters)
	if err != nil {
		return nil, err
	}

	alertEnabled := true
	if req.AlertEnabled != nil {
		alertEnabled = *req.AlertEnabled
	}

	freq := "daily"
	if req.AlertFrequency != "" {
		freq = req.AlertFrequency
	}

	ss := &domain.SavedSearch{
		UserID:         userID,
		Name:           req.Name,
		Filters:        req.Filters,
		AlertEnabled:   alertEnabled,
		AlertFrequency: freq,
		TotalMatches:   count,
	}

	return s.repo.CreateSavedSearch(ctx, ss)
}

func (s *service) GetSavedSearch(ctx context.Context, id, userID uuid.UUID) (*domain.SavedSearch, error) {
	ss, err := s.repo.GetSavedSearchByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	if ss == nil {
		return nil, fmt.Errorf("saved search not found")
	}
	return ss, nil
}

func (s *service) GetUserSavedSearches(ctx context.Context, userID uuid.UUID) ([]domain.SavedSearch, map[string]int, error) {
	searches, err := s.repo.GetUserSavedSearches(ctx, userID)
	if err != nil {
		return nil, nil, err
	}

	// Return basic list, stats embedded in struct
	return searches, nil, nil
}

func (s *service) UpdateSavedSearch(ctx context.Context, id, userID uuid.UUID, req domain.UpdateSavedSearchRequest) (*domain.SavedSearch, error) {
	// Check coverage
	ss, err := s.GetSavedSearch(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		ss.Name = req.Name
	}

	// If filters change, need to recalculate total matches?
	// Or logic "Checked since last time" becomes tricky if criteria changed.
	// We reset LastChecked to NOW to avoid false "new matches" from old criteria history?
	// Or we recalculate TotalMatches immediately.
	updatedFilters := false
	if req.Filters.Brands != nil { // Check basic diff? Assume if provided it's update
		ss.Filters = req.Filters
		updatedFilters = true
	}

	if updatedFilters {
		count, err := s.repo.CountMatchingListings(ctx, ss.Filters)
		if err == nil {
			ss.TotalMatches = count
		}
	}

	err = s.repo.UpdateSavedSearch(ctx, ss)
	if err != nil {
		return nil, err
	}

	return ss, nil
}

func (s *service) DeleteSavedSearch(ctx context.Context, id, userID uuid.UUID) error {
	return s.repo.DeleteSavedSearch(ctx, id, userID)
}

func (s *service) CheckForNewMatches(ctx context.Context, id, userID uuid.UUID) (*domain.CheckMatchesResponse, error) {
	ss, err := s.GetSavedSearch(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	// Get new listings since LastChecked
	newListings, err := s.repo.GetNewListingsSince(ctx, ss.Filters, ss.LastChecked)
	if err != nil {
		return nil, err
	}

	// Recalculate total
	total, err := s.repo.CountMatchingListings(ctx, ss.Filters)
	if err != nil {
		return nil, err
	}

	// Update stats
	// Note: Only update LastChecked if we actually "viewed" or acknowledged?
	// The requirement `POST /check` implies we are checking now.
	// So we update LastChecked to NOW(), and NewMatchesCount to however many we found since OLD LastChecked used above.
	// Wait, if we set LastChecked to NOW, then next time we check from NOW.
	// So "New Listings" are effectively cleared from "New" status after this call?
	// Implementation: "Check" returns them. Usually client calls this, shows badge.
	// If we update LastChecked, we imply user has been notified/seen them.
	// Let's assume this endpoint is "Check and Clear".
	// Or maybe just "Poll"?
	// Documentation says: "Check for new matches and mark as checked."

	err = s.repo.UpdateMatchStats(ctx, id, len(newListings), total)
	if err != nil {
		return nil, err
	}

	return &domain.CheckMatchesResponse{
		NewMatches:   len(newListings),
		TotalMatches: total,
		NewListings:  newListings,
	}, nil
}

func (s *service) RunSearch(ctx context.Context, id, userID uuid.UUID, page, limit int) (*domain.RunSearchResponse, error) {
	ss, err := s.GetSavedSearch(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	listings, total, err := s.repo.RunSearch(ctx, ss.Filters, page, limit)
	if err != nil {
		return nil, err
	}

	pagination := domain.Pagination{
		Page:  page,
		Limit: limit,
		Total: total,
	}
	if limit > 0 {
		pagination.TotalPages = int((total + int64(limit) - 1) / int64(limit))
	}

	return &domain.RunSearchResponse{
		Listings:   listings,
		Pagination: pagination,
	}, nil
}

func (s *service) UpdateAlertSettings(ctx context.Context, id, userID uuid.UUID, req domain.UpdateAlertsRequest) error {
	// Verify ownership
	_, err := s.GetSavedSearch(ctx, id, userID)
	if err != nil {
		return err
	}

	return s.repo.UpdateAlertSettings(ctx, id, userID, req.AlertEnabled, req.AlertFrequency)
}

func (s *service) GetTotalNewMatches(ctx context.Context, userID uuid.UUID) (*domain.NewMatchesResponse, error) {
	total, stats, err := s.repo.GetTotalNewMatches(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &domain.NewMatchesResponse{
		TotalNewMatches: total,
		BySearch:        stats,
	}, nil
}
