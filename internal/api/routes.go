package api

import (
	"hudini-breakfast-module/internal/services"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoutes(router *gin.Engine, breakfastService *services.BreakfastService, db *gorm.DB, jwtSecret string) {
	// CORS middleware
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	router.Use(cors.New(config))

	// Initialize handlers
	authHandler := NewAuthHandler(db, jwtSecret)
	breakfastHandler := NewBreakfastHandler(breakfastService)
	guestHandler := NewGuestHandler(breakfastService)

	// Public routes
	api := router.Group("/api")
	{
		// Authentication routes
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}
	}

	// Protected routes (require authentication)
	protected := api.Group("")
	protected.Use(authHandler.AuthMiddleware())
	{
		// User profile
		protected.GET("/auth/me", authHandler.GetProfile)

		// Breakfast Room Management - Main functionality
		protected.GET("/rooms/breakfast-status", breakfastHandler.GetRoomBreakfastStatus)
		protected.GET("/consumption/history", breakfastHandler.GetConsumptionHistory)
		protected.GET("/reports/daily", breakfastHandler.GetDailyReport)
		protected.GET("/analytics", breakfastHandler.GetAnalytics)
		
		// Guest Management
		protected.GET("/guests", guestHandler.GetGuests)
		protected.POST("/guests", guestHandler.CreateGuest)
		protected.PUT("/guests/:id", guestHandler.UpdateGuest)
		
		// Staff actions (require staff role)
		staff := protected.Group("/")
		staff.Use(authHandler.RequireRole("staff", "manager", "admin"))
		{
			staff.POST("/rooms/:room_number/consume", breakfastHandler.MarkBreakfastConsumed)
		}
	}

	// Health check
	router.GET("/health", HealthCheck)
}
