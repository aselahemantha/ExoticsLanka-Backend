package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/aselahemantha/exoticsLanka/internal/common"
	"github.com/aselahemantha/exoticsLanka/internal/favorites/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	service service.Service
}

func NewHandler(service service.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(r *gin.Engine, auth common.AuthMiddleware) {
	api := r.Group("/api/favorites")
	api.Use(auth.Handle())
	{
		api.GET("", h.GetFavorites)
		api.DELETE("", h.ClearAllFavorites)
		api.GET("/count", h.GetFavoritesCount)
		api.POST("/:listingId", h.AddFavorite)
		api.DELETE("/:listingId", h.RemoveFavorite)
		api.GET("/check/:listingId", h.CheckFavorite)
	}
}

func (h *Handler) getUserID(c *gin.Context) (uuid.UUID, error) {
	val, exists := c.Get("userID")
	if !exists {
		return uuid.Nil, fmt.Errorf("user ID not found in context")
	}
	id, ok := val.(uuid.UUID)
	if !ok {
		return uuid.Nil, fmt.Errorf("invalid user ID type in context")
	}
	return id, nil
}

func (h *Handler) AddFavorite(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	listingIDStr := c.Param("listingId")
	listingID, err := uuid.Parse(listingIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid listing ID"})
		return
	}

	fav, err := h.service.AddFavorite(c.Request.Context(), userID, listingID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Added to favorites",
		"data":    fav,
	})
}

func (h *Handler) RemoveFavorite(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	listingIDStr := c.Param("listingId")
	listingID, err := uuid.Parse(listingIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid listing ID"})
		return
	}

	err = h.service.RemoveFavorite(c.Request.Context(), userID, listingID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Removed from favorites",
	})
}

func (h *Handler) GetFavorites(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	favorites, pagination, err := h.service.GetFavorites(c.Request.Context(), userID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"favorites":  favorites,
			"pagination": pagination,
		},
	})
}

func (h *Handler) CheckFavorite(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	listingIDStr := c.Param("listingId")
	listingID, err := uuid.Parse(listingIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid listing ID"})
		return
	}

	isFavorited, favoritedAt, err := h.service.CheckFavorite(c.Request.Context(), userID, listingID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	data := gin.H{"isFavorited": isFavorited}
	if isFavorited && favoritedAt != nil {
		data["favoritedAt"] = favoritedAt
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
	})
}

func (h *Handler) GetFavoritesCount(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	count, err := h.service.GetFavoritesCount(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"count": count,
		},
	})
}

func (h *Handler) ClearAllFavorites(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	count, err := h.service.ClearAllFavorites(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "All favorites cleared",
		"data": gin.H{
			"removedCount": count,
		},
	})
}
