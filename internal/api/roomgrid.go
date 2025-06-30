package api

import (
	"net/http"
	"strconv"
	"time"

	"hudini-breakfast-module/internal/models"
	"hudini-breakfast-module/internal/services"

	"github.com/gin-gonic/gin"
)

type RoomGridHandler struct {
	roomGridService *services.RoomGridService
}

func NewRoomGridHandler(roomGridService *services.RoomGridService) *RoomGridHandler {
	return &RoomGridHandler{
		roomGridService: roomGridService,
	}
}

// GET /api/room-grid/:propertyId
func (h *RoomGridHandler) GetRoomGrid(c *gin.Context) {
	propertyID := c.Param("propertyId")
	if propertyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Property ID is required"})
		return
	}

	// Parse date parameter (default to today)
	dateStr := c.DefaultQuery("date", time.Now().Format("2006-01-02"))
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format. Use YYYY-MM-DD"})
		return
	}

	roomStatuses, err := h.roomGridService.GetRoomGrid(propertyID, date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"property_id": propertyID,
		"date":        date.Format("2006-01-02"),
		"rooms":       roomStatuses,
	})
}

// GET /api/room-grid/:propertyId/room/:roomNumber
func (h *RoomGridHandler) GetRoomDetails(c *gin.Context) {
	propertyID := c.Param("propertyId")
	roomNumber := c.Param("roomNumber")
	
	if propertyID == "" || roomNumber == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Property ID and Room Number are required"})
		return
	}

	dateStr := c.DefaultQuery("date", time.Now().Format("2006-01-02"))
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format. Use YYYY-MM-DD"})
		return
	}

	roomDetails, err := h.roomGridService.GetRoomDetails(propertyID, roomNumber, date)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"room": roomDetails})
}

// POST /api/room-grid/:propertyId/room/:roomNumber/consume
func (h *RoomGridHandler) MarkBreakfastConsumed(c *gin.Context) {
	propertyID := c.Param("propertyId")
	roomNumber := c.Param("roomNumber")
	
	if propertyID == "" || roomNumber == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Property ID and Room Number are required"})
		return
	}

	// Get staff ID from JWT token
	staffID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Staff authentication required"})
		return
	}

	var req struct {
		PaymentMethod string `json:"payment_method" binding:"required"` // room_charge, ohip, comp, cash
		Notes         string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate payment method
	validMethods := map[string]bool{
		"room_charge": true,
		"ohip":        true,
		"comp":        true,
		"cash":        true,
	}
	if !validMethods[req.PaymentMethod] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payment method"})
		return
	}

	err := h.roomGridService.MarkBreakfastConsumed(
		propertyID, 
		roomNumber, 
		staffID.(uint), 
		req.PaymentMethod, 
		req.Notes,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Breakfast consumption recorded successfully",
		"property_id": propertyID,
		"room_number": roomNumber,
		"consumed_at": time.Now(),
	})
}

// POST /api/room-grid/:propertyId/sync
func (h *RoomGridHandler) SyncFromPMS(c *gin.Context) {
	propertyID := c.Param("propertyId")
	if propertyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Property ID is required"})
		return
	}

	// Check if user has admin/manager role
	userRole, exists := c.Get("user_role")
	if !exists || (userRole != "admin" && userRole != "manager") {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin or manager access required"})
		return
	}

	err := h.roomGridService.SyncGuestsFromPMS(propertyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Guest data synchronized successfully",
		"property_id": propertyID,
		"synced_at":   time.Now(),
	})
}

// GET /api/room-grid/:propertyId/history
func (h *RoomGridHandler) GetConsumptionHistory(c *gin.Context) {
	propertyID := c.Param("propertyId")
	if propertyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Property ID is required"})
		return
	}

	// Parse date range
	startDateStr := c.DefaultQuery("start_date", time.Now().AddDate(0, 0, -7).Format("2006-01-02"))
	endDateStr := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format. Use YYYY-MM-DD"})
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format. Use YYYY-MM-DD"})
		return
	}

	// Add pagination
	page := 1
	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	limit := 50
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	history, err := h.roomGridService.GetConsumptionHistory(propertyID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Apply pagination
	start := (page - 1) * limit
	end := start + limit
	
	if start >= len(history) {
		history = []models.DailyBreakfastConsumption{}
	} else {
		if end > len(history) {
			end = len(history)
		}
		history = history[start:end]
	}

	c.JSON(http.StatusOK, gin.H{
		"property_id":  propertyID,
		"start_date":   startDate.Format("2006-01-02"),
		"end_date":     endDate.Format("2006-01-02"),
		"page":         page,
		"limit":        limit,
		"history":      history,
	})
}

// GET /api/room-grid/:propertyId/report
func (h *RoomGridHandler) GetDailyReport(c *gin.Context) {
	propertyID := c.Param("propertyId")
	if propertyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Property ID is required"})
		return
	}

	dateStr := c.DefaultQuery("date", time.Now().Format("2006-01-02"))
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format. Use YYYY-MM-DD"})
		return
	}

	report, err := h.roomGridService.GenerateDailyReport(propertyID, date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"report": report})
}
