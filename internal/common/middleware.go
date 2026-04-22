package common

import "github.com/gin-gonic/gin"

// AuthMiddleware defines the interface for authentication middleware
type AuthMiddleware interface {
	Handle() gin.HandlerFunc
	OptionalHandle() gin.HandlerFunc
}
