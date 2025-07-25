package api

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"hudini-breakfast-module/internal/logging"
	"hudini-breakfast-module/internal/models"
	"hudini-breakfast-module/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type BreakfastHandler struct {
	breakfastService *services.BreakfastService
}

func NewBreakfastHandler(breakfastService *services.BreakfastService) *BreakfastHandler {
	return &BreakfastHandler{
		breakfastService: breakfastService,
	}
}

// GET /api/rooms/breakfast-status
func (h *BreakfastHandler) GetRoomBreakfastStatus(c *gin.Context) {
	propertyID := c.Query("property_id")
	if propertyID == "" {
		logging.WithFields(logrus.Fields{
			"handler": "GetRoomBreakfastStatus",
			"error":   "missing property_id",
		}).Warn("Bad request")
		ValidationErrorResponse(c, "property_id is required")
		return
	}

	logging.WithFields(logrus.Fields{
		"handler":     "GetRoomBreakfastStatus",
		"property_id": propertyID,
	}).Info("Fetching room breakfast status")

	roomStatuses, err := h.breakfastService.GetRoomBreakfastStatus(propertyID)
	if err != nil {
		logging.WithFields(logrus.Fields{
			"handler":     "GetRoomBreakfastStatus",
			"property_id": propertyID,
			"error":       err.Error(),
		}).Error("Failed to fetch room breakfast status")
		InternalErrorResponse(c, fmt.Errorf("failed to fetch room breakfast status: %w", err))
		return
	}
	
	logging.WithFields(logrus.Fields{
		"handler":     "GetRoomBreakfastStatus",
		"property_id": propertyID,
		"room_count":  len(roomStatuses),
	}).Info("Successfully fetched room breakfast status")

	SuccessResponse(c, gin.H{"rooms": roomStatuses})
}

// POST /api/rooms/:room_number/consume
func (h *BreakfastHandler) MarkBreakfastConsumed(c *gin.Context) {
	roomNumber := c.Param("room_number")
	propertyID := c.Query("property_id")
	
	if propertyID == "" {
		logging.WithFields(logrus.Fields{
			"handler":     "MarkBreakfastConsumed",
			"room_number": roomNumber,
			"error":       "missing property_id",
		}).Warn("Bad request")
		ValidationErrorResponse(c, "property_id is required")
		return
	}

	// Get staff ID from JWT token (this would come from auth middleware)
	staffIDInterface, exists := c.Get("staff_id")
	if !exists {
		logging.WithFields(logrus.Fields{
			"handler":     "MarkBreakfastConsumed",
			"room_number": roomNumber,
			"property_id": propertyID,
			"error":       "missing staff authentication",
		}).Warn("Unauthorized request")
		UnauthorizedResponse(c)
		return
	}
	
	staffID, ok := staffIDInterface.(uint)
	if !ok {
		logging.WithFields(logrus.Fields{
			"handler":     "MarkBreakfastConsumed",
			"room_number": roomNumber,
			"property_id": propertyID,
			"error":       "invalid staff ID type",
		}).Error("Invalid staff ID")
		InternalErrorResponse(c, fmt.Errorf("invalid staff ID"))
		return
	}

	logging.WithFields(logrus.Fields{
		"handler":     "MarkBreakfastConsumed",
		"room_number": roomNumber,
		"property_id": propertyID,
		"staff_id":    staffID,
	}).Info("Marking breakfast as consumed")

	err := h.breakfastService.MarkBreakfastConsumed(propertyID, roomNumber, staffID)
	if err != nil {
		logging.WithFields(logrus.Fields{
			"handler":     "MarkBreakfastConsumed",
			"room_number": roomNumber,
			"property_id": propertyID,
			"staff_id":    staffID,
			"error":       err.Error(),
		}).Error("Failed to mark breakfast as consumed")
		ErrorResponse(c, http.StatusBadRequest, "CONSUMPTION_ERROR", err.Error())
		return
	}
	
	logging.WithFields(logrus.Fields{
		"handler":     "MarkBreakfastConsumed",
		"room_number": roomNumber,
		"property_id": propertyID,
		"staff_id":    staffID,
	}).Info("Successfully marked breakfast as consumed")

	SuccessResponseWithMessage(c, "Breakfast consumption marked successfully", nil)
}

// GET /api/consumption/history
func (h *BreakfastHandler) GetConsumptionHistory(c *gin.Context) {
	propertyID := c.Query("property_id")
	if propertyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "property_id is required"})
		return
	}

	// Parse date range
	startDateStr := c.DefaultQuery("start_date", time.Now().AddDate(0, 0, -7).Format("2006-01-02"))
	endDateStr := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))
	
	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format (YYYY-MM-DD)"})
		return
	}
	
	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format (YYYY-MM-DD)"})
		return
	}

	consumptions, err := h.breakfastService.GetConsumptionHistory(propertyID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"consumptions": consumptions})
}

// GET /api/reports/daily
func (h *BreakfastHandler) GetDailyReport(c *gin.Context) {
	propertyID := c.Query("property_id")
	if propertyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "property_id is required"})
		return
	}

	dateStr := c.DefaultQuery("date", time.Now().Format("2006-01-02"))
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format (YYYY-MM-DD)"})
		return
	}

	report, err := h.breakfastService.GetDailyReport(propertyID, date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"report": report})
}

// GET /api/analytics
func (h *BreakfastHandler) GetAnalytics(c *gin.Context) {
	propertyID := c.Query("property_id")
	if propertyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "property_id is required"})
		return
	}

	period := c.DefaultQuery("period", "week")
	if period != "today" && period != "week" && period != "month" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "period must be one of: today, week, month"})
		return
	}

	analytics, err := h.breakfastService.GetAnalyticsData(propertyID, period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"analytics": analytics})
}

// Guest Management Endpoints
type GuestHandler struct {
	guestService *services.GuestService
}

func NewGuestHandler(guestService *services.GuestService) *GuestHandler {
	return &GuestHandler{
		guestService: guestService,
	}
}

// POST /api/guests
func (h *GuestHandler) CreateGuest(c *gin.Context) {
	var guest models.Guest
	if err := c.ShouldBindJSON(&guest); err != nil {
		logging.WithFields(logrus.Fields{
			"handler": "CreateGuest",
			"error":   err.Error(),
		}).Warn("Invalid JSON in request")
		ValidationErrorResponse(c, err.Error())
		return
	}

	logging.WithFields(logrus.Fields{
		"handler":        "CreateGuest",
		"property_id":    guest.PropertyID,
		"room_number":    guest.RoomNumber,
		"pms_guest_id":   guest.PMSGuestID,
	}).Info("Creating guest")

	err := h.guestService.CreateGuest(&guest)
	if err != nil {
		logging.WithFields(logrus.Fields{
			"handler":        "CreateGuest",
			"property_id":    guest.PropertyID,
			"pms_guest_id":   guest.PMSGuestID,
			"error":          err.Error(),
		}).Error("Failed to create guest")
		ErrorResponse(c, http.StatusBadRequest, "CREATE_GUEST_ERROR", err.Error())
		return
	}

	logging.WithFields(logrus.Fields{
		"handler":        "CreateGuest",
		"guest_id":       guest.ID,
		"property_id":    guest.PropertyID,
		"pms_guest_id":   guest.PMSGuestID,
	}).Info("Successfully created guest")

	CreatedResponse(c, guest)
}

// GET /api/guests
func (h *GuestHandler) GetGuests(c *gin.Context) {
	propertyID := c.Query("property_id")
	// propertyID validation is handled by middleware

	logging.WithFields(logrus.Fields{
		"handler":     "GetGuests",
		"property_id": propertyID,
	}).Info("Fetching guests")

	guests, err := h.guestService.GetGuests(propertyID)
	if err != nil {
		logging.WithFields(logrus.Fields{
			"handler":     "GetGuests",
			"property_id": propertyID,
			"error":       err.Error(),
		}).Error("Failed to fetch guests")
		InternalErrorResponse(c, fmt.Errorf("failed to fetch guests: %w", err))
		return
	}

	logging.WithFields(logrus.Fields{
		"handler":     "GetGuests",
		"property_id": propertyID,
		"count":       len(guests),
	}).Info("Successfully fetched guests")

	SuccessResponse(c, gin.H{"guests": guests})
}

// PUT /api/guests/:id
func (h *GuestHandler) UpdateGuest(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		logging.WithFields(logrus.Fields{
			"handler": "UpdateGuest",
			"id":      c.Param("id"),
			"error":   err.Error(),
		}).Warn("Invalid guest ID")
		ValidationErrorResponse(c, "Invalid guest ID")
		return
	}
	
	var guest models.Guest
	if err := c.ShouldBindJSON(&guest); err != nil {
		logging.WithFields(logrus.Fields{
			"handler":  "UpdateGuest",
			"guest_id": id,
			"error":    err.Error(),
		}).Warn("Invalid JSON in request")
		ValidationErrorResponse(c, err.Error())
		return
	}

	logging.WithFields(logrus.Fields{
		"handler":  "UpdateGuest",
		"guest_id": id,
	}).Info("Updating guest")

	err = h.guestService.UpdateGuest(uint(id), &guest)
	if err != nil {
		logging.WithFields(logrus.Fields{
			"handler":  "UpdateGuest",
			"guest_id": id,
			"error":    err.Error(),
		}).Error("Failed to update guest")
		
		if err.Error() == "guest not found" {
			NotFoundResponse(c, "Guest")
		} else {
			ErrorResponse(c, http.StatusBadRequest, "UPDATE_GUEST_ERROR", err.Error())
		}
		return
	}

	logging.WithFields(logrus.Fields{
		"handler":  "UpdateGuest",
		"guest_id": id,
	}).Info("Successfully updated guest")

	SuccessResponseWithMessage(c, "Guest updated successfully", gin.H{"guest_id": id})
}

// Health check endpoint
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"service":   "breakfast-module",
	})
}
