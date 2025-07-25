package main

import (
	"os"
	"time"

	"hudini-breakfast-module/internal/api"
	"hudini-breakfast-module/internal/cache"
	"hudini-breakfast-module/internal/config"
	"hudini-breakfast-module/internal/database"
	"hudini-breakfast-module/internal/logging"
	"hudini-breakfast-module/internal/services"
	"hudini-breakfast-module/internal/websocket"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		// Don't log error here, we'll handle it after logger initialization
	}

	// Load configuration
	cfg := config.Load()

	// Initialize logger
	logging.InitLogger(logging.LoggingConfig{
		Level:    cfg.Logging.Level,
		Format:   cfg.Logging.Format,
		Output:   cfg.Logging.Output,
		FilePath: cfg.Logging.FilePath,
	})

	logging.Info("Starting Hudini Breakfast Module server...")

	// Initialize database
	db, err := database.InitializeWithConfig(cfg.DatabaseURL, database.DatabaseConfig{
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
	})
	if err != nil {
		logging.Fatalf("Failed to initialize database: %v", err)
	}
	logging.Info("Database initialized successfully")

	// Initialize Redis cache if configured
	var redisCache *cache.RedisCache
	var vipCache *cache.VIPCache
	
	if cfg.RedisURL != "" {
		// Parse Redis URL to get address, password, and DB
		addr := "localhost:6379" // Default
		password := ""
		dbNum := 0
		
		// In production, parse the Redis URL properly
		if cfg.RedisURL != "" {
			// For now, use simple defaults
			addr = "localhost:6379"
		}
		
		redisCache, err = cache.NewRedisCache(addr, password, dbNum, logging.GetLogger())
		if err != nil {
			logging.WithError(err).Warn("Failed to initialize Redis cache, continuing without caching")
		} else {
			logging.Info("Redis cache initialized successfully")
			
			// Initialize VIP cache with 5-minute TTL
			vipCache = cache.NewVIPCache(redisCache, 5*time.Minute)
			logging.Info("VIP cache initialized")
		}
	}

	// Initialize WebSocket hub
	wsHub := websocket.NewHub()
	go wsHub.Run()
	logging.Info("WebSocket hub initialized")

	// Initialize OHIP service
	ohipService := services.NewOHIPService(cfg.OHIP)
	logging.Info("OHIP service initialized")

	// Initialize breakfast service with cache support
	var breakfastService *services.BreakfastService
	if vipCache != nil {
		breakfastService = services.NewBreakfastServiceWithCache(db, ohipService, vipCache)
		logging.Info("Breakfast service initialized with caching")
	} else {
		breakfastService = services.NewBreakfastService(db, ohipService)
		logging.Info("Breakfast service initialized")
	}

	// Initialize guest service
	guestService := services.NewGuestService(db)
	logging.Info("Guest service initialized")
	
	// Initialize audit service
	auditService := services.NewAuditService(db)
	logging.Info("Audit service initialized")
	
	// Initialize notification service
	notificationService := services.NewNotificationService(db, redisCache)
	
	// Set up notification providers (using mock for now)
	pushProvider := services.NewMockPushProvider()
	emailProvider := services.NewMockEmailProvider()
	smsProvider := services.NewMockSMSProvider()
	notificationService.SetProviders(pushProvider, emailProvider, smsProvider)
	notificationService.SetWebSocketHub(wsHub)
	logging.Info("Notification service initialized")

	// Setup router
	router := gin.Default()

	// Setup API routes
	api.SetupRoutes(router, breakfastService, guestService, auditService, notificationService, db, cfg.JWTSecret, wsHub)
	logging.Info("API routes configured")

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logging.WithField("port", port).Info("Server starting...")
	if err := router.Run(":" + port); err != nil {
		logging.Fatalf("Failed to start server: %v", err)
	}
}
