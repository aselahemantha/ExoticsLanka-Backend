package handler

import (
	"net/http"
	"strconv"

	"github.com/aselahemantha/exoticsLanka/services/messaging-service/internal/domain"
	"github.com/aselahemantha/exoticsLanka/services/messaging-service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	service service.Service
}

func NewHandler(service service.Service) *Handler {
	return &Handler{service: service}
}

// GET /api/conversations
func (h *Handler) GetUserConversations(c *gin.Context) {
	userID, err := GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	archived := c.Query("archived") == "true"

	conversations, pagination, err := h.service.GetUserConversations(c.Request.Context(), userID, page, limit, archived)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	// Also get total unread for header/badge separately?
	// Spec says: "GET /api/messages/unread-count" is separate.
	// But list response implies "totalUnread" in example?
	// Current impl returns pagination + list.
	// If required, we can inject `totalUnread` here too.
	// Let's stick to base for now.

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"conversations": conversations,
			"pagination":    pagination,
		},
	})
}

// POST /api/conversations
func (h *Handler) CreateConversation(c *gin.Context) {
	userID, err := GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req domain.CreateConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.service.CreateConversation(c.Request.Context(), req, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	msg := "Conversation created"
	if !resp.IsNew {
		msg = "Message sent to existing conversation"
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": msg,
		"data":    resp,
	})
}

// GET /api/conversations/:id
func (h *Handler) GetConversationByID(c *gin.Context) {
	userID, err := GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid conversation ID"})
		return
	}

	conv, messages, err := h.service.GetConversationByID(c.Request.Context(), id, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	// Helper to merge messages into response
	// We might want to restructure response to match spec exactly
	// Spec: { data: { ...convFields, messages: [] } }
	// We have conv struct and messages slice.
	// We can't easily add field to `domain.Conversation` without dirtying it.
	// We'll construct a map.

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id": conv.ID,
			"listing": gin.H{
				"id":    conv.ListingID,
				"title": conv.ListingTitle,
				"image": conv.ListingImage,
				"price": conv.ListingPrice,
			},
			"participant":   conv.Participant,
			"messages":      messages,
			"createdAt":     conv.CreatedAt,
			"lastMessageAt": conv.LastMessageAt,
		},
	})
}

// POST /api/conversations/:id/messages
func (h *Handler) SendMessage(c *gin.Context) {
	userID, err := GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid conversation ID"})
		return
	}

	var req domain.SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	msg, err := h.service.SendMessage(c.Request.Context(), id, userID, req.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    msg,
	})
}

// PUT /api/conversations/:id/read
func (h *Handler) MarkRead(c *gin.Context) {
	userID, err := GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid conversation ID"})
		return
	}

	err = h.service.MarkConversationRead(c.Request.Context(), id, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Messages marked as read",
	})
}

// GET /api/messages/unread-count
func (h *Handler) GetUnreadCount(c *gin.Context) {
	userID, err := GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	stats, err := h.service.GetUnreadCount(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}
