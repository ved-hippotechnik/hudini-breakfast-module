package api

import (
	"net/http"
	"strconv"
	"time"

	"hudini-breakfast-module/internal/logging"
	"hudini-breakfast-module/internal/services"

	"github.com/gin-gonic/gin"
)

type NotificationHandler struct {
	notificationService *services.NotificationService
}

func NewNotificationHandler(notificationService *services.NotificationService) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
	}
}

// RegisterDevice registers a device for push notifications
func (h *NotificationHandler) RegisterDevice(c *gin.Context) {
	var req RegisterDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Get user ID from context (set by auth middleware)
	userID := c.GetUint("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Register device
	device := &services.UserDevice{
		UserID:       userID,
		DeviceID:     req.DeviceID,
		DeviceType:   req.DeviceType,
		DeviceName:   req.DeviceName,
		PushToken:    req.PushToken,
		PushEnabled:  true,
		LastActiveAt: time.Now(),
	}

	if err := h.notificationService.RegisterDevice(device); err != nil {
		logging.WithError(err).Error("Failed to register device")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register device"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"device":  device,
	})
}

// UpdateDeviceToken updates the push token for a device
func (h *NotificationHandler) UpdateDeviceToken(c *gin.Context) {
	deviceID := c.Param("device_id")
	
	var req UpdateTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	userID := c.GetUint("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	if err := h.notificationService.UpdateDeviceToken(userID, deviceID, req.PushToken); err != nil {
		logging.WithError(err).Error("Failed to update device token")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update device token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// GetNotifications retrieves notifications for the authenticated user
func (h *NotificationHandler) GetNotifications(c *gin.Context) {
	userID := c.GetUint("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Parse query parameters
	unreadOnly := c.Query("unread_only") == "true"
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	notifications, total, err := h.notificationService.GetUserNotifications(userID, unreadOnly, limit, offset)
	if err != nil {
		logging.WithError(err).Error("Failed to get notifications")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get notifications"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"notifications": notifications,
		"total":        total,
		"limit":        limit,
		"offset":       offset,
	})
}

// MarkNotificationRead marks a notification as read
func (h *NotificationHandler) MarkNotificationRead(c *gin.Context) {
	notificationID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification ID"})
		return
	}

	userID := c.GetUint("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	if err := h.notificationService.MarkNotificationRead(uint(notificationID), userID); err != nil {
		logging.WithError(err).Error("Failed to mark notification as read")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark notification as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// MarkAllNotificationsRead marks all notifications as read
func (h *NotificationHandler) MarkAllNotificationsRead(c *gin.Context) {
	userID := c.GetUint("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	if err := h.notificationService.MarkAllNotificationsRead(userID); err != nil {
		logging.WithError(err).Error("Failed to mark all notifications as read")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark notifications as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// GetNotificationPreferences gets user notification preferences
func (h *NotificationHandler) GetNotificationPreferences(c *gin.Context) {
	userID := c.GetUint("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	preferences, err := h.notificationService.GetNotificationPreferences(userID)
	if err != nil {
		logging.WithError(err).Error("Failed to get notification preferences")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get preferences"})
		return
	}

	c.JSON(http.StatusOK, preferences)
}

// UpdateNotificationPreferences updates user notification preferences
func (h *NotificationHandler) UpdateNotificationPreferences(c *gin.Context) {
	userID := c.GetUint("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var pref services.NotificationPreference
	if err := c.ShouldBindJSON(&pref); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if err := h.notificationService.UpdateNotificationPreferences(userID, &pref); err != nil {
		logging.WithError(err).Error("Failed to update notification preferences")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update preferences"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// SendTestNotification sends a test notification (admin only)
func (h *NotificationHandler) SendTestNotification(c *gin.Context) {
	var req SendTestNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	userID := c.GetUint("userID")
	propertyID := c.Query("property_id")
	if propertyID == "" {
		propertyID = "HOTEL001"
	}

	// Convert channels
	channels := make([]services.NotificationChannel, len(req.Channels))
	for i, ch := range req.Channels {
		channels[i] = services.NotificationChannel(ch)
	}

	// Create test notification
	notifReq := &services.CreateNotificationRequest{
		Type:        services.NotificationSystemAlert,
		Priority:    services.NotificationPriority(req.Priority),
		Title:       req.Title,
		Message:     req.Message,
		PropertyID:  propertyID,
		RecipientID: userID,
		Channels:    channels,
	}

	notification, err := h.notificationService.CreateNotification(c.Request.Context(), notifReq)
	if err != nil {
		logging.WithError(err).Error("Failed to send test notification")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send notification"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"notification": notification,
	})
}

// GetNotificationStats gets notification statistics (admin only)
func (h *NotificationHandler) GetNotificationStats(c *gin.Context) {
	propertyID := c.Query("property_id")
	if propertyID == "" {
		propertyID = "HOTEL001"
	}

	periodStr := c.DefaultQuery("period", "24h")
	period, err := time.ParseDuration(periodStr)
	if err != nil {
		period = 24 * time.Hour
	}

	stats, err := h.notificationService.GetNotificationStats(propertyID, period)
	if err != nil {
		logging.WithError(err).Error("Failed to get notification stats")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get statistics"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// Request structures
type RegisterDeviceRequest struct {
	DeviceID   string `json:"device_id" binding:"required"`
	DeviceType string `json:"device_type" binding:"required,oneof=ios android web"`
	DeviceName string `json:"device_name"`
	PushToken  string `json:"push_token" binding:"required"`
}

type UpdateTokenRequest struct {
	PushToken string `json:"push_token" binding:"required"`
}

type SendTestNotificationRequest struct {
	Title    string   `json:"title" binding:"required"`
	Message  string   `json:"message" binding:"required"`
	Priority string   `json:"priority" binding:"required,oneof=low medium high critical"`
	Channels []string `json:"channels" binding:"required,min=1"`
}