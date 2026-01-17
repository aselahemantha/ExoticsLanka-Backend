package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aselahemantha/exoticsLanka/services/listings-service/internal/config"
	"github.com/aselahemantha/exoticsLanka/services/listings-service/internal/handler"
	"github.com/aselahemantha/exoticsLanka/services/listings-service/internal/jobs"
	"github.com/aselahemantha/exoticsLanka/services/listings-service/internal/repository"
	"github.com/aselahemantha/exoticsLanka/services/listings-service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// 1. Load Config
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Printf("Warning: Failed to load config: %v", err)
	}

	// 2. Connect to Database
	dbPool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer dbPool.Close()

	if err := dbPool.Ping(context.Background()); err != nil {
		log.Fatalf("Unable to ping database: %v\n", err)
	}
	log.Println("Connected to PostgreSQL")

	// 3. Run Migrations
	if err := repository.RunMigrations(context.Background(), dbPool, "migrations"); err != nil {
		log.Printf("Warning: Failed to run migrations: %v", err)
	}

	// 4. Initialize Layers
	repo := repository.NewPostgresRepository(dbPool)
	svc := service.NewService(repo)
	h := handler.NewHandler(svc)

	// 5. Start Background Jobs
	jobScheduler := jobs.NewJobScheduler(dbPool)
	jobScheduler.Start()

	// 6. Setup Router
	router := gin.Default()

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	api := router.Group("/api")
	api.Use(handler.AuthMiddleware(cfg))
	{
		// Listings
		api.POST("/listings", h.CreateListing)
		api.GET("/listings", h.GetListings)
		api.GET("/listings/:id", h.GetListing)
		api.GET("/listings/featured", h.GetFeatured)
		api.GET("/listings/trending", h.GetTrending)

		// Brands
		api.GET("/brands", h.GetBrands)
	}

	// 7. Start Server
	port := cfg.Port
	if port == "" {
		port = "8081" // Default port different from auth service (8080 usually)
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	go func() {
		log.Printf("Starting Listings Service on port %s", port)
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
