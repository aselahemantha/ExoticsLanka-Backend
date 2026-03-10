package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aselahemantha/exoticsLanka/services/user-service/internal/config"
	delivery "github.com/aselahemantha/exoticsLanka/services/user-service/internal/delivery/http"
	"github.com/aselahemantha/exoticsLanka/services/user-service/internal/repository"
	"github.com/aselahemantha/exoticsLanka/services/user-service/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// 1. Load Config
	cfg := config.LoadConfig()

	// 2. Connect to PostgreSQL
	dbPool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer dbPool.Close()

	if err := dbPool.Ping(context.Background()); err != nil {
		log.Fatalf("Unable to ping database: %v\n", err)
	}
	log.Println("Connected to PostgreSQL")

	// Note: You would theoretically run migrations here exactly like auth-service
	// repository.RunMigrations(context.Background(), dbPool, "sql/migrations")

	// 3. Initialize Layers
	userRepo := repository.NewPostgresUserRepository(dbPool)
	verificationRepo := repository.NewPostgresVerificationRepository(dbPool)

	profileUC := usecase.NewProfileUseCase(userRepo, verificationRepo)
	profileHandler := delivery.NewProfileHandler(profileUC)
	authMiddleware := delivery.NewAuthMiddleware(cfg.JWTSecret)

	// 4. Setup Router
	router := gin.Default()
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	profileHandler.RegisterRoutes(router, authMiddleware)

	// 5. Start Server
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		log.Printf("Starting User Service on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

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
