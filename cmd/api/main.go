package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aselahemantha/exoticsLanka/internal/common"
	"github.com/aselahemantha/exoticsLanka/internal/config"

	// Auth Module
	authHttp "github.com/aselahemantha/exoticsLanka/internal/auth/delivery/http"
	authRepo "github.com/aselahemantha/exoticsLanka/internal/auth/repository"
	authUC "github.com/aselahemantha/exoticsLanka/internal/auth/usecase"

	// Listings Module
	listingsHttp "github.com/aselahemantha/exoticsLanka/internal/listings/delivery/http"
	listingsRepo "github.com/aselahemantha/exoticsLanka/internal/listings/repository"
	listingsUC "github.com/aselahemantha/exoticsLanka/internal/listings/usecase"

	// User Module
	userHttp "github.com/aselahemantha/exoticsLanka/internal/user/delivery/http"
	userRepo "github.com/aselahemantha/exoticsLanka/internal/user/repository"
	userUC "github.com/aselahemantha/exoticsLanka/internal/user/usecase"

	// Image Module
	imageHttp "github.com/aselahemantha/exoticsLanka/internal/image/handler"
	imageRepo "github.com/aselahemantha/exoticsLanka/internal/image/repository"
	imageService "github.com/aselahemantha/exoticsLanka/internal/image/service"
	imageStorage "github.com/aselahemantha/exoticsLanka/internal/image/storage"

	// Messaging Module
	messagingHttp "github.com/aselahemantha/exoticsLanka/internal/messaging/handler"
	messagingRepo "github.com/aselahemantha/exoticsLanka/internal/messaging/repository"
	messagingUC "github.com/aselahemantha/exoticsLanka/internal/messaging/service"

	// Notification Module
	notificationHttp "github.com/aselahemantha/exoticsLanka/internal/notification/handler"
	notificationProvider "github.com/aselahemantha/exoticsLanka/internal/notification/provider"
	notificationRepo "github.com/aselahemantha/exoticsLanka/internal/notification/repository"
	notificationUC "github.com/aselahemantha/exoticsLanka/internal/notification/service"

	// Reviews Module
	reviewsHttp "github.com/aselahemantha/exoticsLanka/internal/reviews/handler"
	reviewsRepo "github.com/aselahemantha/exoticsLanka/internal/reviews/repository"
	reviewsSvc "github.com/aselahemantha/exoticsLanka/internal/reviews/service"

	// Favorites Module
	favoritesHttp "github.com/aselahemantha/exoticsLanka/internal/favorites/handler"
	favoritesRepo "github.com/aselahemantha/exoticsLanka/internal/favorites/repository"
	favoritesSvc "github.com/aselahemantha/exoticsLanka/internal/favorites/service"

	// Phase 5 Modules
	analyticsHttp "github.com/aselahemantha/exoticsLanka/internal/analytics/handler"
	analyticsRepo "github.com/aselahemantha/exoticsLanka/internal/analytics/repository"
	analyticsSvc "github.com/aselahemantha/exoticsLanka/internal/analytics/service"

	comparisonHttp "github.com/aselahemantha/exoticsLanka/internal/comparison/handler"
	comparisonRepo "github.com/aselahemantha/exoticsLanka/internal/comparison/repository"
	comparisonSvc "github.com/aselahemantha/exoticsLanka/internal/comparison/service"

	contactHttp "github.com/aselahemantha/exoticsLanka/internal/contact/handler"
	contactRepo "github.com/aselahemantha/exoticsLanka/internal/contact/repository"
	contactSvc "github.com/aselahemantha/exoticsLanka/internal/contact/service"

	// Phase 6 Modules
	reportsHttp "github.com/aselahemantha/exoticsLanka/internal/reports/handler"
	reportsRepo "github.com/aselahemantha/exoticsLanka/internal/reports/repository"
	reportsSvc "github.com/aselahemantha/exoticsLanka/internal/reports/service"

	savedSearchesHttp "github.com/aselahemantha/exoticsLanka/internal/saved_searches/handler"
	savedSearchesRepo "github.com/aselahemantha/exoticsLanka/internal/saved_searches/repository"
	savedSearchesSvc "github.com/aselahemantha/exoticsLanka/internal/saved_searches/service"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Load Config
	cfg := config.LoadConfig()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 2. Connect to PostgreSQL
	log.Printf("Attempting to connect to PostgreSQL...")
	dbPool, err := common.InitDatabase(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Printf("CRITICAL ERROR: Failed to initialize database: %v", err)
		log.Printf("App will continue to start the server for debugging purposes, but database features will fail.")
	} else {
		defer dbPool.Close()
	}

	// 3. Connect to Redis
	log.Printf("Attempting to connect to Redis...")
	rdb, err := common.InitRedis(ctx, cfg.RedisURL)
	if err != nil {
		log.Printf("CRITICAL ERROR: Failed to initialize Redis: %v", err)
	} else {
		defer rdb.Close()
	}

	// 4. Initialize External Clients
	cldClient, err := imageStorage.NewCloudinaryClient(cfg.CloudinaryURL)
	if err != nil {
		log.Printf("Warning: Failed to initialize Cloudinary client: %v", err)
	}

	sendGridProv := notificationProvider.NewSendGridProvider(cfg.SendGridAPIKey, cfg.FromEmail, cfg.FromName)
	twilioProv := notificationProvider.NewTwilioProvider(cfg.TwilioAccountSID, cfg.TwilioAuthToken, cfg.TwilioFromNumber)

	// 5. Run Migrations
	log.Println("Running migrations...")
	migrationPaths := []string{
		"migrations/auth",
		"migrations/listings",
		"migrations/user",
		"migrations/messaging",
		"migrations/notification",
		"migrations/reviews",
		"migrations/favorites",
		"migrations/analytics",
		"migrations/comparison",
		"migrations/contact",
		"migrations/reports",
		"migrations/saved_searches",
	}
	for _, path := range migrationPaths {
		if err := common.RunMigrations(ctx, dbPool, path); err != nil {
			log.Printf("Warning: Failed to run migrations from %s: %v", path, err)
		}
	}

	// 6. Initialize Modules Layer by Layer

	// Auth Module
	authRP := authRepo.NewPostgresUserRepository(dbPool)
	authAuditRP := authRepo.NewPostgresAuditRepository(dbPool)
	authSessionRP := authRepo.NewRedisSessionRepository(rdb)
	authTokenRP := authRepo.NewPostgresTokenRepository(dbPool)

	authUseCase := authUC.NewAuthUseCase(authRP, authSessionRP, authAuditRP, authTokenRP, cfg)
	authHandler := authHttp.NewAuthHandler(authUseCase)
	authMiddleware := authHttp.NewAuthMiddleware(cfg, authSessionRP)

	// Listings Module
	listingRP := listingsRepo.NewPostgresListingRepository(dbPool)
	listingUseCase := listingsUC.NewListingUseCase(listingRP)
	listingHandler := listingsHttp.NewListingHandler(listingUseCase)

	// User Module
	uRP := userRepo.NewPostgresUserRepository(dbPool)
	uVerificationRP := userRepo.NewPostgresVerificationRepository(dbPool)
	userUseCase := userUC.NewProfileUseCase(uRP, uVerificationRP)
	userHandler := userHttp.NewProfileHandler(userUseCase)

	// Image Module
	imRP := imageRepo.NewRepository(dbPool)
	imSvc := imageService.NewService(imRP, cldClient)
	imageHandler := imageHttp.NewHandler(imSvc)

	// Messaging Module
	msgRP := messagingRepo.NewPostgresRepository(dbPool)
	msgUC := messagingUC.NewService(msgRP)
	messagingHandler := messagingHttp.NewHandler(msgUC)

	// Notification Module
	ntRP := notificationRepo.NewRepository(dbPool)
	ntUC := notificationUC.NewService(ntRP, sendGridProv, twilioProv)
	notificationHandler := notificationHttp.NewHandler(ntUC)

	// Reviews Module
	revRP := reviewsRepo.NewPostgresRepository(dbPool)
	revSvc := reviewsSvc.NewService(revRP)
	reviewsHandler := reviewsHttp.NewHandler(revSvc)

	// Favorites Module
	favRP := favoritesRepo.NewPostgresRepository(dbPool)
	favSvc := favoritesSvc.NewService(favRP)
	favoritesHandler := favoritesHttp.NewHandler(favSvc)

	// Phase 5 Modules
	anaRP := analyticsRepo.NewPostgresRepository(dbPool)
	anaSvc := analyticsSvc.NewService(anaRP)
	analyticsHandler := analyticsHttp.NewHandler(anaSvc)

	compRP := comparisonRepo.NewPostgresRepository(dbPool)
	compSvc := comparisonSvc.NewService(compRP)
	comparisonHandler := comparisonHttp.NewHandler(compSvc)

	contRP := contactRepo.NewPostgresRepository(dbPool)
	contSvc := contactSvc.NewService(contRP)
	contactHandler := contactHttp.NewHandler(contSvc)

	// Phase 6 Modules
	repRP := reportsRepo.NewPostgresRepository(dbPool)
	repSvc := reportsSvc.NewService(repRP)
	reportsHandler := reportsHttp.NewHandler(repSvc)

	ssRP := savedSearchesRepo.NewPostgresRepository(dbPool)
	ssSvc := savedSearchesSvc.NewService(ssRP)
	savedSearchesHandler := savedSearchesHttp.NewHandler(ssSvc)

	// 7. Setup Router
	router := gin.Default()

	// Global Health Check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "monolith",
			"modules": []string{
				"auth", "listings", "user", "image", "messaging",
				"notification", "reviews", "favorites", "analytics",
				"comparison", "contact", "reports", "saved_searches",
			},
		})
	})

	// Register Module Routes
	authHandler.RegisterRoutes(router, authMiddleware)
	listingHandler.RegisterRoutes(router, authMiddleware)
	userHandler.RegisterRoutes(router, authMiddleware)
	imageHandler.RegisterRoutes(router, authMiddleware)
	messagingHandler.RegisterRoutes(router, authMiddleware)
	notificationHandler.RegisterRoutes(router, authMiddleware)
	reviewsHandler.RegisterRoutes(router, authMiddleware)
	favoritesHandler.RegisterRoutes(router, authMiddleware)
	analyticsHandler.RegisterRoutes(router, authMiddleware)
	comparisonHandler.RegisterRoutes(router, authMiddleware)
	contactHandler.RegisterRoutes(router, authMiddleware)
	reportsHandler.RegisterRoutes(router, authMiddleware)
	savedSearchesHandler.RegisterRoutes(router, authMiddleware)

	// 8. Start Server
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		log.Printf("Starting Monolith Server on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()

	if err := srv.Shutdown(ctxShutdown); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
	log.Println("Server exiting")
}
