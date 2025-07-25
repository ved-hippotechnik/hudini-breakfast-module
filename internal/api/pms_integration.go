package api

import (
	"context"
	"net/http"
	"time"

	"hudini-breakfast-module/internal/logging"
	"hudini-breakfast-module/internal/services"

	"github.com/gin-gonic/gin"
)

// PMSIntegrationHandler handles PMS integration API endpoints
type PMSIntegrationHandler struct {
	pmsService *services.PMSIntegrationService
}

// NewPMSIntegrationHandler creates a new PMS integration handler
func NewPMSIntegrationHandler(pmsService *services.PMSIntegrationService) *PMSIntegrationHandler {
	return &PMSIntegrationHandler{
		pmsService: pmsService,
	}
}

// RegisterRoutes registers PMS integration routes
func (h *PMSIntegrationHandler) RegisterRoutes(r *gin.RouterGroup) {
	pms := r.Group("/pms")
	{
		// Health check
		pms.GET("/health", h.HealthCheck)
		
		// Provider management
		pms.GET("/providers", h.GetProviders)
		pms.POST("/providers/:name/switch", h.SwitchProvider)
		pms.POST("/providers/refresh-tokens", h.RefreshTokens)
		
		// Guest management
		pms.GET("/guests/room/:roomNumber", h.GetGuestByRoom)
		pms.GET("/guests/reservation/:reservationId", h.GetGuestByReservation)
		pms.GET("/guests/property/:propertyId", h.GetGuestsByProperty)
		
		// Room management
		pms.GET("/rooms/:roomNumber", h.GetRoomStatus)
		pms.GET("/rooms/property/:propertyId", h.GetRoomsByProperty)
		
		// Charges
		pms.POST("/charges/breakfast", h.PostBreakfastCharge)
		
		// Sync
		pms.POST("/sync/property/:propertyId", h.SyncPropertyData)
	}
}

// HealthCheck checks the health of all PMS providers
func (h *PMSIntegrationHandler) HealthCheck(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	results := h.pmsService.HealthCheck(ctx)
	
	response := gin.H{
		"status":    "ok",
		"timestamp": time.Now().Format(time.RFC3339),
		"providers": results,
	}
	
	// Check if any provider failed
	hasFailure := false
	for _, err := range results {
		if err != nil {
			hasFailure = true
			break
		}
	}
	
	if hasFailure {
		response["status"] = "degraded"
		c.JSON(http.StatusPartialContent, response)
	} else {
		c.JSON(http.StatusOK, response)
	}
}

// GetProviders returns all registered PMS providers
func (h *PMSIntegrationHandler) GetProviders(c *gin.Context) {
	providers := h.pmsService.GetProviderNames()
	
	c.JSON(http.StatusOK, gin.H{
		"providers": providers,
		"count":     len(providers),
	})
}

// SwitchProvider switches to a different PMS provider
func (h *PMSIntegrationHandler) SwitchProvider(c *gin.Context) {
	providerName := c.Param("name")
	
	if err := h.pmsService.SwitchProvider(providerName); err != nil {
		logging.Error("Failed to switch provider:", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to switch provider",
			"details": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Provider switched successfully",
		"provider": providerName,
	})
}

// RefreshTokens refreshes authentication tokens for all providers
func (h *PMSIntegrationHandler) RefreshTokens(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	
	if err := h.pmsService.RefreshProviderTokens(ctx); err != nil {
		logging.Error("Failed to refresh tokens:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to refresh tokens",
			"details": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Tokens refreshed successfully",
	})
}

// GetGuestByRoom retrieves guest information by room number
func (h *PMSIntegrationHandler) GetGuestByRoom(c *gin.Context) {
	roomNumber := c.Param("roomNumber")
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	guest, err := h.pmsService.GetGuestProfile(ctx, roomNumber)
	if err != nil {
		logging.Error("Failed to get guest by room:", err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Guest not found",
			"details": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"guest": guest,
	})
}

// GetGuestByReservation retrieves guest information by reservation ID
func (h *PMSIntegrationHandler) GetGuestByReservation(c *gin.Context) {
	reservationID := c.Param("reservationId")
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	guest, err := h.pmsService.GetGuestByReservation(ctx, reservationID)
	if err != nil {
		logging.Error("Failed to get guest by reservation:", err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Guest not found",
			"details": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"guest": guest,
	})
}

// GetGuestsByProperty retrieves all guests for a property
func (h *PMSIntegrationHandler) GetGuestsByProperty(c *gin.Context) {
	propertyID := c.Param("propertyId")
	
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	
	guests, err := h.pmsService.GetAllGuests(ctx, propertyID)
	if err != nil {
		logging.Error("Failed to get guests by property:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve guests",
			"details": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"guests": guests,
		"count":  len(guests),
	})
}

// GetRoomStatus retrieves room status information
func (h *PMSIntegrationHandler) GetRoomStatus(c *gin.Context) {
	roomNumber := c.Param("roomNumber")
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	room, err := h.pmsService.GetRoomStatus(ctx, roomNumber)
	if err != nil {
		logging.Error("Failed to get room status:", err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Room not found",
			"details": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"room": room,
	})
}

// GetRoomsByProperty retrieves all rooms for a property
func (h *PMSIntegrationHandler) GetRoomsByProperty(c *gin.Context) {
	propertyID := c.Param("propertyId")
	
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	
	rooms, err := h.pmsService.GetAllRooms(ctx, propertyID)
	if err != nil {
		logging.Error("Failed to get rooms by property:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve rooms",
			"details": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"rooms": rooms,
		"count": len(rooms),
	})
}

// PostBreakfastCharge posts a breakfast charge to PMS
func (h *PMSIntegrationHandler) PostBreakfastCharge(c *gin.Context) {
	var request struct {
		GuestID    string  `json:"guest_id" binding:"required"`
		RoomNumber string  `json:"room_number" binding:"required"`
		Amount     float64 `json:"amount" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"details": err.Error(),
		})
		return
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := h.pmsService.PostBreakfastCharge(ctx, request.GuestID, request.RoomNumber, request.Amount); err != nil {
		logging.Error("Failed to post breakfast charge:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to post charge",
			"details": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Breakfast charge posted successfully",
		"guest_id": request.GuestID,
		"room_number": request.RoomNumber,
		"amount": request.Amount,
	})
}

// SyncPropertyData synchronizes property data with PMS
func (h *PMSIntegrationHandler) SyncPropertyData(c *gin.Context) {
	propertyID := c.Param("propertyId")
	
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second) // 5 minutes for sync
	defer cancel()
	
	if err := h.pmsService.SyncRoomData(ctx, propertyID); err != nil {
		logging.Error("Failed to sync property data:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to sync property data",
			"details": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Property data synchronized successfully",
		"property_id": propertyID,
	})
}

// PMSStatusResponse represents the PMS status response
type PMSStatusResponse struct {
	Status        string                 `json:"status"`
	Timestamp     string                 `json:"timestamp"`
	Providers     map[string]interface{} `json:"providers"`
	DefaultProvider string               `json:"default_provider"`
}

// GetPMSStatus returns the overall PMS integration status
func (h *PMSIntegrationHandler) GetPMSStatus(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	healthResults := h.pmsService.HealthCheck(ctx)
	providers := h.pmsService.GetProviderNames()
	
	providerStatus := make(map[string]interface{})
	overallStatus := "healthy"
	
	for _, providerName := range providers {
		if err, exists := healthResults[providerName]; exists {
			if err != nil {
				providerStatus[providerName] = map[string]interface{}{
					"status": "unhealthy",
					"error":  err.Error(),
				}
				overallStatus = "degraded"
			} else {
				providerStatus[providerName] = map[string]interface{}{
					"status": "healthy",
				}
			}
		}
	}
	
	response := PMSStatusResponse{
		Status:        overallStatus,
		Timestamp:     time.Now().Format(time.RFC3339),
		Providers:     providerStatus,
		DefaultProvider: "oracle_ohip", // TODO: Get from config
	}
	
	statusCode := http.StatusOK
	if overallStatus == "degraded" {
		statusCode = http.StatusPartialContent
	}
	
	c.JSON(statusCode, response)
}

// TestPMSConnection tests the connection to a specific PMS provider
func (h *PMSIntegrationHandler) TestPMSConnection(c *gin.Context) {
	providerName := c.Param("name")
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	provider, err := h.pmsService.GetProviderWithName(providerName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Provider not found",
			"details": err.Error(),
		})
		return
	}
	
	if err := provider.HealthCheck(ctx); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"provider": providerName,
			"status":   "unhealthy",
			"error":    err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"provider": providerName,
		"status":   "healthy",
		"message":  "Connection test successful",
	})
}
