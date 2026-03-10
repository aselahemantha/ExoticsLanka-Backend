package http

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// AuthMiddleware decodes JWT tokens and manages protected requests
type AuthMiddleware struct {
	jwtSecret string
}

func NewAuthMiddleware(secret string) *AuthMiddleware {
	return &AuthMiddleware{jwtSecret: secret}
}

func (m *AuthMiddleware) OptionalHandle() gin.HandlerFunc {
	return func(c *gin.Context) {
		m.processToken(c, false)
		c.Next()
	}
}

func (m *AuthMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		m.processToken(c, true)
		if !c.IsAborted() {
			c.Next()
		}
	}
}

func (m *AuthMiddleware) processToken(c *gin.Context, required bool) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		if required {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
		}
		return
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		if required {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
			c.Abort()
		}
		return
	}

	tokenStr := parts[1]
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(m.jwtSecret), nil
	})

	if err != nil || !token.Valid {
		if required {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
		}
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		if required {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
		}
		return
	}

	subject, ok := claims["sub"].(string)
	if !ok {
		if required {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid subject in token"})
			c.Abort()
		}
		return
	}

	userID, err := uuid.Parse(subject)
	if err != nil {
		if required {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID format in token"})
			c.Abort()
		}
		return
	}

	// Set context variables
	c.Set("userID", userID)
}
