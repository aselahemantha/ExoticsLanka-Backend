package handler

import (
	"net/http"

	"github.com/aselahemantha/exoticsLanka/internal/common"
	"github.com/aselahemantha/exoticsLanka/internal/image/domain"
	"github.com/aselahemantha/exoticsLanka/internal/image/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	service *service.Service
}

func NewHandler(service *service.Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) RegisterRoutes(r *gin.Engine, auth common.AuthMiddleware) {
	api := r.Group("/api")
	{
		// Listings images
		listings := api.Group("/listings/:id/images")
		listings.Use(auth.Handle())
		{
			listings.POST("", h.UploadListingImage)
			listings.PUT("/reorder", h.ReorderListingImages)
			listings.PUT("/:imageId/primary", h.SetCoverImage)
			listings.DELETE("/:imageId", h.DeleteListingImage)
		}

		// User avatar
		users := api.Group("/users")
		users.Use(auth.Handle())
		{
			users.PUT("/me/avatar", h.UploadUserAvatar)
		}
	}
}

func (h *Handler) UploadListingImage(c *gin.Context) {
	listingID := c.Param("id")
	userID, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "unauthorized"})
		return
	}

	file, header, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Image file required"})
		return
	}
	defer file.Close()

	resp, err := h.service.UploadListingImage(c.Request.Context(), listingID, userID.(uuid.UUID).String(), file, header)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": resp})
}

func (h *Handler) ReorderListingImages(c *gin.Context) {
	listingID := c.Param("id")
	userID, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "unauthorized"})
		return
	}

	var req domain.ReorderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	err := h.service.ReorderImages(c.Request.Context(), listingID, userID.(uuid.UUID).String(), req.ImageIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *Handler) DeleteListingImage(c *gin.Context) {
	_ = c.Param("id") // listingID - not strictly needed
	imageID := c.Param("imageId")
	userID, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "unauthorized"})
		return
	}

	err := h.service.DeleteListingImage(c.Request.Context(), imageID, userID.(uuid.UUID).String())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *Handler) SetCoverImage(c *gin.Context) {
	listingID := c.Param("id")
	imageID := c.Param("imageId")
	userID, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "unauthorized"})
		return
	}

	err := h.service.SetCoverImage(c.Request.Context(), listingID, userID.(uuid.UUID).String(), imageID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *Handler) UploadUserAvatar(c *gin.Context) {
	userID, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "unauthorized"})
		return
	}

	file, header, err := c.Request.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Avatar file required"})
		return
	}
	defer file.Close()

	url, err := h.service.UploadUserAvatar(c.Request.Context(), userID.(uuid.UUID).String(), file, header)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"url": url}})
}
