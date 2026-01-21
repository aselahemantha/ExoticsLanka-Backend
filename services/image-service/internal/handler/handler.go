package handler

import (
	"net/http"

	"github.com/aselahemantha/exoticsLanka/services/image-service/internal/domain"
	"github.com/aselahemantha/exoticsLanka/services/image-service/internal/service"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service    *service.Service
	authSecret string
}

func NewHandler(service *service.Service, authSecret string) *Handler {
	return &Handler{
		service:    service,
		authSecret: authSecret,
	}
}

func (h *Handler) UploadListingImage(c *gin.Context) {
	listingID := c.Param("id")
	userID := c.GetString("userID")

	file, header, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Image file required"})
		return
	}
	defer file.Close()

	resp, err := h.service.UploadListingImage(c.Request.Context(), listingID, userID, file, header)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": resp})
}

func (h *Handler) ReorderListingImages(c *gin.Context) {
	listingID := c.Param("id")
	userID := c.GetString("userID")

	var req domain.ReorderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	err := h.service.ReorderImages(c.Request.Context(), listingID, userID, req.ImageIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *Handler) DeleteListingImage(c *gin.Context) {
	_ = c.Param("id") // listingID - not strictly needed if we just delete by imageID, but good for URL structure
	imageID := c.Param("imageId")
	userID := c.GetString("userID")

	err := h.service.DeleteListingImage(c.Request.Context(), imageID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *Handler) UploadUserAvatar(c *gin.Context) {
	userID := c.GetString("userID")

	file, header, err := c.Request.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Avatar file required"})
		return
	}
	defer file.Close()

	url, err := h.service.UploadUserAvatar(c.Request.Context(), userID, file, header)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"url": url}})
}
