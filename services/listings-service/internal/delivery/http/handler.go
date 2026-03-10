package http

import (
	"net/http"
	"strconv"

	"github.com/aselahemantha/exoticsLanka/services/listings-service/internal/domain"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ListingHandler struct {
	uc domain.ListingUseCase
}

func NewListingHandler(uc domain.ListingUseCase) *ListingHandler {
	return &ListingHandler{uc: uc}
}

func (h *ListingHandler) RegisterRoutes(r *gin.Engine, auth *AuthMiddleware) {
	api := r.Group("/api")

	// Public Discovery
	api.GET("/listings", auth.OptionalHandle(), h.GetListings)
	api.GET("/listings/:id", auth.OptionalHandle(), h.GetListing)
	api.GET("/listings/featured", auth.OptionalHandle(), h.GetFeatured)
	api.GET("/listings/trending", auth.OptionalHandle(), h.GetTrending)
	api.GET("/brands", h.GetBrands)
	api.GET("/models", h.GetModels)
	api.POST("/listings/:id/view", auth.OptionalHandle(), h.IncrementView)

	// Protected Endpoints
	protected := api.Group("")
	protected.Use(auth.Handle())
	{
		// CRUD
		protected.POST("/listings", h.CreateListing)
		protected.PATCH("/listings/:id", h.UpdateListing)
		protected.DELETE("/listings/:id", h.DeleteListing)

		// Self Listings
		protected.GET("/users/me/listings", h.GetMyListings)

		// Favorites
		protected.POST("/listings/:id/favorite", h.AddFavorite)
		protected.DELETE("/listings/:id/favorite", h.RemoveFavorite)
		protected.GET("/users/me/favorites", h.GetMyFavorites)

		// Reports
		protected.POST("/listings/:id/reports", h.ReportListing)

		// Note: We don't implement full image proxying/upload processing in this spec, just keeping placeholders for real life
		// protected.POST("/listings/:id/images", h.UploadImages)
	}
}

// ---------------------- Public ----------------------

func (h *ListingHandler) GetListings(c *gin.Context) {
	// Parse Filters
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pMin, _ := strconv.ParseFloat(c.Query("priceMin"), 64)
	pMax, _ := strconv.ParseFloat(c.Query("priceMax"), 64)
	yMin, _ := strconv.Atoi(c.Query("yearMin"))
	yMax, _ := strconv.Atoi(c.Query("yearMax"))
	mMin, _ := strconv.Atoi(c.Query("mileageMin"))
	mMax, _ := strconv.Atoi(c.Query("mileageMax"))

	filter := domain.ListingFilter{
		Query:        c.Query("q"),
		Brand:        c.Query("make"),
		Model:        c.Query("model"),
		Location:     c.Query("location"),
		FuelType:     c.Query("fuelType"),
		Transmission: c.Query("transmission"),
		Condition:    c.Query("condition"),
		Sort:         c.Query("sort"),
		PriceMin:     pMin, PriceMax: pMax,
		YearMin: yMin, YearMax: yMax,
		MileageMin: mMin, MileageMax: mMax,
		Limit: limit,
		Page:  page,
	}

	if c.Query("isFeatured") == "true" {
		b := true
		filter.IsFeatured = &b
	}
	if c.Query("isTrending") == "true" {
		b := true
		filter.IsTrending = &b
	}

	res, err := h.uc.GetListings(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *ListingHandler) GetListing(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid listing ID"})
		return
	}
	res, err := h.uc.GetListing(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *ListingHandler) GetFeatured(c *gin.Context) {
	res, err := h.uc.GetFeaturedListings(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *ListingHandler) GetTrending(c *gin.Context) {
	res, err := h.uc.GetTrendingListings(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *ListingHandler) GetBrands(c *gin.Context) {
	res, err := h.uc.GetBrands(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *ListingHandler) GetModels(c *gin.Context) {
	res, err := h.uc.GetModelsByBrand(c.Request.Context(), c.Query("brand"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *ListingHandler) IncrementView(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid listing ID"})
		return
	}
	_ = h.uc.IncrementView(c.Request.Context(), id)
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ---------------------- Protected ----------------------

func (h *ListingHandler) CreateListing(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)
	var req domain.CreateListingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	listing, err := h.uc.CreateListing(c.Request.Context(), userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, listing)
}

func (h *ListingHandler) UpdateListing(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid listing ID"})
		return
	}

	var req domain.UpdateListingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	listing, err := h.uc.UpdateListing(c.Request.Context(), userID, id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, listing)
}

func (h *ListingHandler) DeleteListing(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid listing ID"})
		return
	}
	if err := h.uc.DeleteListing(c.Request.Context(), userID, id); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *ListingHandler) GetMyListings(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

	filter := domain.ListingFilter{SellerID: &userID, Limit: limit, Page: page}

	res, err := h.uc.GetListings(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *ListingHandler) AddFavorite(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid listing ID"})
		return
	}

	if err := h.uc.ToggleFavorite(c.Request.Context(), userID, id, true); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *ListingHandler) RemoveFavorite(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid listing ID"})
		return
	}

	if err := h.uc.ToggleFavorite(c.Request.Context(), userID, id, false); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *ListingHandler) GetMyFavorites(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	res, err := h.uc.GetUserFavorites(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *ListingHandler) ReportListing(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid listing ID"})
		return
	}

	var req domain.ReportListingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.uc.ReportListing(c.Request.Context(), userID, id, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true})
}
