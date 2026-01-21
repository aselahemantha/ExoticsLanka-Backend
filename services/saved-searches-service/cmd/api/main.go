package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aselahemantha/exoticsLanka/services/saved-searches-service/internal/config"
	"github.com/aselahemantha/exoticsLanka/services/saved-searches-service/internal/handler"
	"github.com/aselahemantha/exoticsLanka/services/saved-searches-service/internal/repository"
	"github.com/aselahemantha/exoticsLanka/services/saved-searches-service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// 1. Load Configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. Database Connection
	dbPool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer dbPool.Close()

	if err := dbPool.Ping(context.Background()); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	fmt.Println("Connected to PostgreSQL")

	// 3. Run Migrations
	if err := repository.RunMigrations(dbPool, "migrations"); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// 4. Initialize Dependency Injection
	repo := repository.NewPostgresRepository(dbPool)
	svc := service.NewService(repo)
	h := handler.NewHandler(svc)

	// 5. Setup Router
	router := gin.Default()

	// CORS
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	api := router.Group("/api")

	api.Use(handler.AuthMiddleware(cfg))
	{
		api.GET("/searches", h.GetUserSearches)
		api.POST("/searches", h.CreateSavedSearch)

		api.GET("/searches/:id", h.GetDetailedSearch)
		api.PUT("/searches/:id", h.UpdateSavedSearch)
		api.DELETE("/searches/:id", h.DeleteSavedSearch)

		api.POST("/searches/:id/check", h.CheckMatches)
		api.POST("/searches/:id/run", h.RunSearch)
		api.PUT("/searches/:id/alerts", h.UpdateAlerts)

		api.GET("/searches/new-matches", h.GetNewMatchesOverview)
	}

	// 6. Start Server
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		fmt.Printf("Saved Searches Service starting on port %s\n", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// 7. Graceful Shutdown
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
