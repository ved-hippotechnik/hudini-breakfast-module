package main

import (
	"log"
	"os"

	"hudini-breakfast-module/internal/api"
	"hudini-breakfast-module/internal/config"
	"hudini-breakfast-module/internal/database"
	"hudini-breakfast-module/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := database.Initialize(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize OHIP service
	ohipService := services.NewOHIPService(cfg.OHIP)

	// Initialize breakfast service
	breakfastService := services.NewBreakfastService(db, ohipService)

	// Setup router
	router := gin.Default()
	
	// Setup API routes
	api.SetupRoutes(router, breakfastService, db, cfg.JWTSecret)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
