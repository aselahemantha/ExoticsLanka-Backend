package http

import (
	"net/http"

	"github.com/aselahemantha/exoticsLanka/internal/common"
	"github.com/aselahemantha/exoticsLanka/internal/user/domain"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ProfileHandler struct {
	useCase domain.ProfileUseCase
}

func NewProfileHandler(useCase domain.ProfileUseCase) *ProfileHandler {
	return &ProfileHandler{
		useCase: useCase,
	}
}

func (h *ProfileHandler) RegisterRoutes(router *gin.Engine, authMiddleware common.AuthMiddleware) {
	users := router.Group("/api/users")
	users.Use(authMiddleware.Handle())
	{
		users.GET("/me", h.GetProfile)
		users.PATCH("/me", h.UpdateProfile)
		users.DELETE("/me", h.DeleteAccount)
		users.POST("/me/verification", h.RequestVerification)
		// users.POST("/me/avatar", h.UpdateAvatar) // Optional / Future
	}
}

func (h *ProfileHandler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	resp, err := h.useCase.GetProfile(c.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *ProfileHandler) UpdateProfile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req domain.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.useCase.UpdateProfile(c.Request.Context(), userID.(uuid.UUID), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *ProfileHandler) DeleteAccount(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	if err := h.useCase.DeleteAccount(c.Request.Context(), userID.(uuid.UUID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Account deleted successfully"})
}

func (h *ProfileHandler) RequestVerification(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req domain.VerificationRequestPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.useCase.RequestVerification(c.Request.Context(), userID.(uuid.UUID), &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Verification request submitted successfully"})
}
