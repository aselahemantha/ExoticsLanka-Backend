package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/aselahemantha/exoticsLanka/services/listings-service/internal/domain"
	"github.com/aselahemantha/exoticsLanka/services/listings-service/internal/repository"
	"github.com/aselahemantha/exoticsLanka/services/listings-service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	svc service.Service
}

func NewHandler(svc service.Service) *Handler {
	return &Handler{svc: svc}
}

// CreateListing handles creating a new listing
// POST /api/listings
func (h *Handler) CreateListing(c *gin.Context) {
	var req domain.CreateListingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context (assuming auth middleware sets it)
	// For now using a hardcoded or header-based ID for testing if no auth middleware
	userIDStr := c.GetHeader("X-User-ID")
	if userIDStr == "" {
		// fallback for dev/testing
		userIDStr = "00000000-0000-0000-0000-000000000000"
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	listing, err := h.svc.CreateListing(c.Request.Context(), &req, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Listing created successfully",
		"data": gin.H{
			"id":     listing.ID,
			"status": listing.Status,
		},
	})
}

// GetListing handles getting a single listing
// GET /api/listings/:id
func (h *Handler) GetListing(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid listing ID"})
		return
	}

	listing, err := h.svc.GetListing(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    listing,
	})
}

// GetListings handles searching and filtering
// GET /api/listings
func (h *Handler) GetListings(c *gin.Context) {
	// Parse query params
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	minPrice, _ := strconv.ParseFloat(c.Query("minPrice"), 64)
	maxPrice, _ := strconv.ParseFloat(c.Query("maxPrice"), 64)

	filter := repository.ListingFilter{
		Query:    c.Query("search"),
		MinPrice: minPrice,
		MaxPrice: maxPrice,
		SortBy:   c.Query("sortBy"),
		Page:     page,
		Limit:    limit,
		// Add other filters parsing...
	}

	if brands := c.Query("brands"); brands != "" {
		filter.Brands = strings.Split(brands, ",")
	}

	listings, total, err := h.svc.GetListings(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	totalPages := (total + limit - 1) / limit

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"listings": listings,
			"pagination": gin.H{
				"page":       page,
				"limit":      limit,
				"total":      total,
				"totalPages": totalPages,
			},
		},
	})
}

// GetFeatured handles getting featured listings
// GET /api/listings/featured
func (h *Handler) GetFeatured(c *gin.Context) {
	listings, err := h.svc.GetFeaturedListings(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    listings,
	})
}

// GetTrending handles getting trending listings
// GET /api/listings/trending
func (h *Handler) GetTrending(c *gin.Context) {
	listings, err := h.svc.GetTrendingListings(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    listings,
	})
}

// GetBrands handles getting all brands
// GET /api/brands
func (h *Handler) GetBrands(c *gin.Context) {
	brands, err := h.svc.GetBrands(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    brands,
	})
}
