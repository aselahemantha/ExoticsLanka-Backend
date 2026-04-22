package http

import (
	"net/http"
	"strings"

	"github.com/aselahemantha/exoticsLanka/internal/config"
	"github.com/aselahemantha/exoticsLanka/internal/auth/domain"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type AuthMiddleware struct {
	cfg         *config.Config
	sessionRepo domain.SessionRepository
}

func NewAuthMiddleware(cfg *config.Config, sessionRepo domain.SessionRepository) *AuthMiddleware {
	return &AuthMiddleware{
		cfg:         cfg,
		sessionRepo: sessionRepo,
	}
}

func (m *AuthMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		m.processToken(c, true)
	}
}

func (m *AuthMiddleware) OptionalHandle() gin.HandlerFunc {
	return func(c *gin.Context) {
		m.processToken(c, false)
	}
}

func (m *AuthMiddleware) processToken(c *gin.Context, required bool) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		if required {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
		}
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		if required {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
		}
		return
	}

	// 1. Verify JWT Signature and Claims
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(m.cfg.JWTSecret), nil
	})

	if err != nil || !token.Valid {
		if required {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		}
		return
	}

	// 2. Check if session is active in Redis (Session Revocation Check)
	session, err := m.sessionRepo.GetByToken(c.Request.Context(), tokenString)
	if err != nil {
		if required {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Session validation failed"})
		}
		return
	}
	if session == nil {
		if required {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Session expired or revoked"})
		}
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		if required {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
		}
		return
	}

	userIDStr, ok := claims["sub"].(string)
	if !ok {
		if required {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token user id"})
		}
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		if required {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token user id format"})
		}
		return
	}

	c.Set("userID", userID)
	c.Set("role", claims["role"])

	c.Next()
}

// Optional: Role-based authorization middleware
func (m *AuthMiddleware) RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		roleStr, ok := userRole.(string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid role"})
			return
		}

		for _, role := range roles {
			if role == roleStr {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
	}
}
