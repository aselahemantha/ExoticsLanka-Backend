package service

import (
	"context"
	"fmt"
	"math"

	"github.com/aselahemantha/exoticsLanka/services/comparison-service/internal/domain"
	"github.com/aselahemantha/exoticsLanka/services/comparison-service/internal/repository"
	"github.com/google/uuid"
)

type Service interface {
	AddToComparison(ctx context.Context, userID, listingID uuid.UUID) (int, error)
	RemoveFromComparison(ctx context.Context, userID, listingID uuid.UUID) error
	ClearComparison(ctx context.Context, userID uuid.UUID) error
	GetComparison(ctx context.Context, userID uuid.UUID) (*domain.ComparisonData, error)
	GetComparisonList(ctx context.Context, userID uuid.UUID) (*domain.ComparisonListResponse, error)
	CheckStatus(ctx context.Context, userID, listingID uuid.UUID) (bool, error)
}

type service struct {
	repo repository.Repository
}

func NewService(repo repository.Repository) Service {
	return &service{repo: repo}
}

func (s *service) AddToComparison(ctx context.Context, userID, listingID uuid.UUID) (int, error) {
	// 1. Check Limit
	count, err := s.repo.GetComparisonCount(ctx, userID)
	if err != nil {
		return 0, err
	}
	if count >= 4 {
		return count, fmt.Errorf("limit exceeded: you can only compare up to 4 vehicles")
	}

	// 2. Validate Listing Exists/Active
	exists, err := s.repo.CheckListingExists(ctx, listingID)
	if err != nil {
		return 0, err
	}
	if !exists {
		return count, fmt.Errorf("listing not found or inactive")
	}

	// 3. Add
	if err := s.repo.AddToComparison(ctx, userID, listingID); err != nil {
		return count, err
	}

	return s.repo.GetComparisonCount(ctx, userID)
}

func (s *service) RemoveFromComparison(ctx context.Context, userID, listingID uuid.UUID) error {
	return s.repo.RemoveFromComparison(ctx, userID, listingID)
}

func (s *service) ClearComparison(ctx context.Context, userID uuid.UUID) error {
	return s.repo.ClearComparison(ctx, userID)
}

func (s *service) CheckStatus(ctx context.Context, userID, listingID uuid.UUID) (bool, error) {
	return s.repo.IsListingInComparison(ctx, userID, listingID)
}

func (s *service) GetComparisonList(ctx context.Context, userID uuid.UUID) (*domain.ComparisonListResponse, error) {
	vehicles, err := s.repo.GetComparisonItems(ctx, userID)
	if err != nil {
		return nil, err
	}

	items := make([]domain.MinimalListing, len(vehicles))
	for i, v := range vehicles {
		items[i] = domain.MinimalListing{
			ID: v.ID, Title: v.Title, Make: v.Specs["make"].(string), Model: v.Specs["model"].(string),
			Year: v.Specs["year"].(int), Price: v.Specs["price"].(float64), Mileage: v.Specs["mileage"].(int),
			Image: v.Image,
		}
	}

	return &domain.ComparisonListResponse{
		Items:    items,
		Count:    len(items),
		MaxItems: 4,
	}, nil
}

func (s *service) GetComparison(ctx context.Context, userID uuid.UUID) (*domain.ComparisonData, error) {
	vehicles, err := s.repo.GetComparisonItems(ctx, userID)
	if err != nil {
		return nil, err
	}

	if len(vehicles) == 0 {
		return &domain.ComparisonData{Vehicles: []domain.VehicleComparison{}, Comparison: domain.ComparisonAnalysis{}}, nil
	}

	// Analyze Data
	var minPrice, maxPrice, minYear, maxYear, minMiles, maxMiles float64
	minPrice, minYear, minMiles = math.MaxFloat64, math.MaxFloat64, math.MaxFloat64

	allFeatures := make([][]string, len(vehicles))

	for i, v := range vehicles {
		// Update ranges
		p := v.Specs["price"].(float64)
		y := float64(v.Specs["year"].(int))
		m := float64(v.Specs["mileage"].(int))

		if p < minPrice {
			minPrice = p
		}
		if p > maxPrice {
			maxPrice = p
		}
		if y < minYear {
			minYear = y
		}
		if y > maxYear {
			maxYear = y
		}
		if m < minMiles {
			minMiles = m
		}
		if m > maxMiles {
			maxMiles = m
		}

		allFeatures[i] = v.Features
	}

	// Common Features
	common := []string{}
	// Intersection logic
	if len(allFeatures) > 0 {
		base := allFeatures[0]
		for _, f := range base {
			isCommon := true
			for j := 1; j < len(allFeatures); j++ {
				found := false
				for _, other := range allFeatures[j] {
					if other == f {
						found = true
						break
					}
				}
				if !found {
					isCommon = false
					break
				}
			}
			if isCommon {
				common = append(common, f)
			}
		}
	}

	// Unique Features
	unique := make(map[string][]string)
	for _, v := range vehicles {
		myUnique := []string{}
		for _, f := range v.Features {
			isCommon := false
			for _, c := range common {
				if c == f {
					isCommon = true
					break
				}
			}
			if !isCommon {
				myUnique = append(myUnique, f)
			}
		}
		unique[v.ID.String()] = myUnique
	}

	analysis := domain.ComparisonAnalysis{
		PriceRange:     domain.RangeData{Lowest: minPrice, Highest: maxPrice, Difference: maxPrice - minPrice},
		YearRange:      domain.RangeData{Lowest: minYear, Highest: maxYear, Difference: maxYear - minYear},
		MileageRange:   domain.RangeData{Lowest: minMiles, Highest: maxMiles, Difference: maxMiles - minMiles},
		CommonFeatures: common,
		UniqueFeatures: unique,
	}

	return &domain.ComparisonData{
		Vehicles:   vehicles,
		Comparison: analysis,
	}, nil
}
