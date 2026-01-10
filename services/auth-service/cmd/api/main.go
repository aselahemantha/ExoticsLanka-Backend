package main

import (
	"context"
	"log"
	netHttp "net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/exoticsLanka/auth-service/internal/config"
	"github.com/exoticsLanka/auth-service/internal/delivery/http"
	"github.com/exoticsLanka/auth-service/internal/repository"
	"github.com/exoticsLanka/auth-service/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
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

	// 2.1 Run Migrations
	// Dockerfile places migrations in ./sql/migrations
	if err := repository.RunMigrations(context.Background(), dbPool, "sql/migrations"); err != nil {
		log.Printf("Warning: Failed to run migrations: %v", err)
		// We don't Fatalf here because sometimes local dev paths differ,
		// but in Prod it should ideally work or we assume it's fine if tables exist.
		// However, for strict startup, we might want to log.Fatal.
		// Let's log.Fatal to be safe, assuming the path is correct in Docker.
	}

	// 3. Connect to Redis
	redisOpts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Fatalf("Invalid Redis URL: %v\n", err)
	}
	rdb := redis.NewClient(redisOpts)
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Unable to connect to Redis: %v\n", err)
	}
	log.Println("Connected to Redis")

	// 4. Initialize Layers
	userRepo := repository.NewPostgresUserRepository(dbPool)
	auditRepo := repository.NewPostgresAuditRepository(dbPool)
	sessionRepo := repository.NewRedisSessionRepository(rdb)

	authUC := usecase.NewAuthUseCase(userRepo, sessionRepo, auditRepo, cfg)
	authHandler := http.NewAuthHandler(authUC)
	authMiddleware := http.NewAuthMiddleware(cfg, sessionRepo)

	// 5. Setup Router
	router := gin.Default()
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	authHandler.RegisterRoutes(router, authMiddleware)

	// 6. Start Server
	srv := &netHttp.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		log.Printf("Starting Auth Service on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != netHttp.ErrServerClosed {
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
