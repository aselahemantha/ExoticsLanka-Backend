package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aselahemantha/exoticsLanka/services/notification-service/internal/config"
	"github.com/aselahemantha/exoticsLanka/services/notification-service/internal/handler"
	"github.com/aselahemantha/exoticsLanka/services/notification-service/internal/provider"
	"github.com/aselahemantha/exoticsLanka/services/notification-service/internal/repository"
	"github.com/aselahemantha/exoticsLanka/services/notification-service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg := config.LoadConfig()

	// Database Connection
	dbPool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer dbPool.Close()

	// Run Migrations
	repository.RunMigrations(dbPool)

	// Providers
	emailProvider := provider.NewSendGridProvider(cfg.SendGridAPIKey, cfg.EmailFrom, cfg.EmailFromName)
	smsProvider := provider.NewTwilioProvider(cfg.TwilioAccountSID, cfg.TwilioAuthToken, cfg.TwilioPhoneNumber)

	// Dependencies
	repo := repository.NewRepository(dbPool)
	svc := service.NewService(repo, emailProvider, smsProvider)
	h := handler.NewHandler(svc)
	auth := handler.NewAuthMiddleware(cfg.JWTSecret)

	// Router
	r := gin.Default()

	// CORS (Basic)
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	api := r.Group("/api/notifications")
	api.Use(auth.RequireAuth())
	{
		api.GET("/preferences", h.GetPreferences)
		api.PUT("/preferences", h.UpdatePreferences)

		// Internal Trigger Endpoint
		api.POST("/send", h.SendNotification)
	}

	// Server
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	log.Printf("Notification Service running on port %s", cfg.Port)

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
	log.Println("Server exiting")
}
