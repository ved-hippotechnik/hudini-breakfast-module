package api

import (
	"os"
	"strings"

	// "hudini-breakfast-module/internal/middleware"
	"hudini-breakfast-module/internal/services"
	"hudini-breakfast-module/internal/validation"
	"hudini-breakfast-module/internal/websocket"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoutes(router *gin.Engine, breakfastService *services.BreakfastService, guestService *services.GuestService, auditService *services.AuditService, db *gorm.DB, jwtSecret string, wsHub *websocket.Hub) {
	// CORS middleware with security improvements
	config := cors.DefaultConfig()

	// Get allowed origins from environment
	allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
	if allowedOrigins != "" {
		config.AllowOrigins = strings.Split(allowedOrigins, ",")
	} else {
		// Default for development
		config.AllowOrigins = []string{"http://localhost:3000", "http://localhost:3001", "http://localhost:8080"}
	}

	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowCredentials = true
	router.Use(cors.New(config))
	
	// Add audit middleware
	// router.Use(middleware.AuditMiddleware(auditService))

	// Initialize handlers
	authHandler := NewAuthHandler(db, jwtSecret)
	breakfastHandler := NewBreakfastHandler(breakfastService)
	guestHandler := NewGuestHandler(guestService)
	auditHandler := NewAuditHandler(auditService)
	executiveHandler := NewExecutiveHandler(breakfastService, guestService)

	// Public routes
	api := router.Group("/api")
	{
		// Authentication routes
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		// Public demo endpoints (no auth required)
		demo := api.Group("/demo")
		{
			demo.GET("/rooms/breakfast-status", 
				validation.ValidatePropertyID(),
				breakfastHandler.GetRoomBreakfastStatus)
			demo.GET("/analytics/advanced", 
				validation.ValidatePropertyID(),
				breakfastHandler.GetAdvancedAnalytics)
			demo.GET("/analytics/realtime", 
				validation.ValidatePropertyID(),
				breakfastHandler.GetRealtimeMetrics)
			
			// Executive demo routes
			demo.GET("/executive/kpis", executiveHandler.GetExecutiveKPIs)
			demo.GET("/executive/vip-trends", executiveHandler.GetVIPTrends)
			demo.GET("/executive/service-performance", executiveHandler.GetServicePerformance)
			demo.GET("/executive/revenue-analysis", executiveHandler.GetRevenueAnalysis)
			demo.GET("/executive/guest-preferences", executiveHandler.GetGuestPreferences)
			demo.GET("/executive/upset-guests", executiveHandler.GetUpsetVIPGuests)
			demo.GET("/executive/alerts", executiveHandler.GetExecutiveAlerts)
		}
	}

	// Protected routes (require authentication)
	protected := api.Group("")
	protected.Use(authHandler.AuthMiddleware())
	{
		// User profile
		protected.GET("/auth/me", authHandler.GetProfile)

		// Breakfast Room Management - Main functionality
		protected.GET("/rooms/breakfast-status", 
			validation.ValidatePropertyID(),
			breakfastHandler.GetRoomBreakfastStatus)
		protected.GET("/consumption/history", 
			validation.ValidatePropertyID(),
			validation.ValidateDateFormat("start_date"),
			validation.ValidateDateFormat("end_date"),
			breakfastHandler.GetConsumptionHistory)
		protected.GET("/reports/daily", 
			validation.ValidatePropertyID(),
			validation.ValidateDateFormat("date"),
			breakfastHandler.GetDailyReport)

		// Analytics and Business Intelligence endpoints
		protected.GET("/analytics", 
			validation.ValidatePropertyID(),
			breakfastHandler.GetAnalytics)
		protected.GET("/analytics/advanced", 
			validation.ValidatePropertyID(),
			breakfastHandler.GetAdvancedAnalytics)
		protected.GET("/analytics/realtime", 
			validation.ValidatePropertyID(),
			breakfastHandler.GetRealtimeMetrics)
		protected.GET("/analytics/predictive", 
			validation.ValidatePropertyID(),
			breakfastHandler.GetPredictiveInsights)
		protected.GET("/analytics/business-intelligence", 
			validation.ValidatePropertyID(),
			breakfastHandler.GetBusinessIntelligence)

		// Guest Management
		protected.GET("/guests", 
			validation.ValidatePropertyID(),
			guestHandler.GetGuests)
		protected.POST("/guests", 
			validation.RequestSizeLimit(1024*1024), // 1MB limit
			guestHandler.CreateGuest)
		protected.PUT("/guests/:id", 
			validation.RequestSizeLimit(1024*1024), // 1MB limit
			guestHandler.UpdateGuest)

		// Staff actions (require staff role)
		staff := protected.Group("/")
		staff.Use(authHandler.RequireRole("staff", "manager", "admin"))
		{
			staff.POST("/rooms/:room_number/consume", 
				validation.ValidatePropertyID(),
				validation.ValidateRoomNumber(),
				breakfastHandler.MarkBreakfastConsumed)
		}
		
		// Admin-only routes
		admin := protected.Group("/")
		admin.Use(authHandler.RequireRole("admin"))
		{
			// Audit log endpoints
			admin.GET("/audit/logs", auditHandler.GetAuditLogs)
			admin.GET("/audit/users/:user_id/activity", auditHandler.GetUserActivity)
			admin.GET("/audit/resources/:resource/:resource_id/history", auditHandler.GetResourceHistory)
			admin.GET("/audit/summary", auditHandler.GetAuditSummary)
		}
		
		// Executive routes (require manager or admin role)
		executive := protected.Group("/executive")
		executive.Use(authHandler.RequireRole("manager", "admin"))
		{
			executive.GET("/kpis", executiveHandler.GetExecutiveKPIs)
			executive.GET("/vip-trends", executiveHandler.GetVIPTrends)
			executive.GET("/service-performance", executiveHandler.GetServicePerformance)
			executive.GET("/revenue-analysis", executiveHandler.GetRevenueAnalysis)
			executive.GET("/guest-preferences", executiveHandler.GetGuestPreferences)
			executive.GET("/upset-guests", executiveHandler.GetUpsetVIPGuests)
			executive.GET("/alerts", executiveHandler.GetExecutiveAlerts)
		}
	}

	// WebSocket endpoint
	router.GET("/ws", func(c *gin.Context) {
		wsHub.ServeWS(c.Writer, c.Request)
	})

	// Health check
	router.GET("/health", HealthCheck)
}
