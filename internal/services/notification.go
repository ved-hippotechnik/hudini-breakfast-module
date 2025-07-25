package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"hudini-breakfast-module/internal/cache"
	"hudini-breakfast-module/internal/logging"
	"hudini-breakfast-module/internal/models"

	"gorm.io/gorm"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	NotificationVIPArrival    NotificationType = "vip_arrival"
	NotificationUpsetGuest    NotificationType = "upset_guest"
	NotificationServiceDelay  NotificationType = "service_delay"
	NotificationBreakfastRush NotificationType = "breakfast_rush"
	NotificationLowSupplies   NotificationType = "low_supplies"
	NotificationStaffAlert    NotificationType = "staff_alert"
	NotificationSystemAlert   NotificationType = "system_alert"
)

// NotificationPriority represents the priority level
type NotificationPriority string

const (
	PriorityLow      NotificationPriority = "low"
	PriorityMedium   NotificationPriority = "medium"
	PriorityHigh     NotificationPriority = "high"
	PriorityCritical NotificationPriority = "critical"
)

// NotificationChannel represents the delivery channel
type NotificationChannel string

const (
	ChannelPush      NotificationChannel = "push"
	ChannelEmail     NotificationChannel = "email"
	ChannelSMS       NotificationChannel = "sms"
	ChannelWebSocket NotificationChannel = "websocket"
	ChannelSlack     NotificationChannel = "slack"
)

// NotificationService handles all notification operations
type NotificationService struct {
	db           *gorm.DB
	cache        *cache.RedisCache
	wsHub        interface{} // WebSocket hub interface
	pushProvider PushNotificationProvider
	emailProvider EmailProvider
	smsProvider  SMSProvider
}

// Notification represents a notification in the system
type Notification struct {
	ID             uint                 `json:"id" gorm:"primaryKey"`
	Type           NotificationType     `json:"type" gorm:"not null"`
	Priority       NotificationPriority `json:"priority" gorm:"not null"`
	Title          string              `json:"title" gorm:"not null"`
	Message        string              `json:"message" gorm:"not null"`
	Data           string              `json:"data" gorm:"type:text"` // JSON string
	PropertyID     string              `json:"property_id" gorm:"not null"`
	RecipientID    uint                `json:"recipient_id,omitempty"`
	RecipientRole  string              `json:"recipient_role,omitempty"`
	Channels       string              `json:"channels" gorm:"type:text"` // JSON string
	Sent           bool                 `json:"sent" gorm:"default:false"`
	SentAt         *time.Time          `json:"sent_at,omitempty"`
	Read           bool                 `json:"read" gorm:"default:false"`
	ReadAt         *time.Time          `json:"read_at,omitempty"`
	ExpiresAt      *time.Time          `json:"expires_at,omitempty"`
	CreatedAt      time.Time           `json:"created_at"`
	UpdatedAt      time.Time           `json:"updated_at"`
}

// NotificationPreference represents user notification preferences
type NotificationPreference struct {
	UserID                uint      `json:"user_id" gorm:"primaryKey"`
	NotificationType      string    `json:"notification_type" gorm:"primaryKey"`
	EnablePush           bool      `json:"enable_push" gorm:"default:true"`
	EnableEmail          bool      `json:"enable_email" gorm:"default:true"`
	EnableSMS            bool      `json:"enable_sms" gorm:"default:false"`
	QuietHoursStart      string    `json:"quiet_hours_start" gorm:"default:'22:00'"`
	QuietHoursEnd        string    `json:"quiet_hours_end" gorm:"default:'07:00'"`
	MinimumPriority      string    `json:"minimum_priority" gorm:"default:'medium'"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// PushNotificationProvider interface for push notification services
type PushNotificationProvider interface {
	Send(ctx context.Context, token string, notification *PushNotification) error
	SendBatch(ctx context.Context, tokens []string, notification *PushNotification) error
}

// EmailProvider interface for email services
type EmailProvider interface {
	Send(ctx context.Context, to string, subject string, body string) error
}

// SMSProvider interface for SMS services
type SMSProvider interface {
	Send(ctx context.Context, to string, message string) error
}

// PushNotification represents a push notification payload
type PushNotification struct {
	Title    string                 `json:"title"`
	Body     string                 `json:"body"`
	Sound    string                 `json:"sound,omitempty"`
	Badge    int                    `json:"badge,omitempty"`
	Data     map[string]interface{} `json:"data,omitempty"`
	Priority string                 `json:"priority,omitempty"`
}

// NewNotificationService creates a new notification service
func NewNotificationService(db *gorm.DB, cache *cache.RedisCache) *NotificationService {
	return &NotificationService{
		db:    db,
		cache: cache,
	}
}

// SetProviders sets the notification providers
func (s *NotificationService) SetProviders(push PushNotificationProvider, email EmailProvider, sms SMSProvider) {
	s.pushProvider = push
	s.emailProvider = email
	s.smsProvider = sms
}

// SetWebSocketHub sets the WebSocket hub for real-time notifications
func (s *NotificationService) SetWebSocketHub(hub interface{}) {
	s.wsHub = hub
}

// CreateNotification creates and sends a notification
func (s *NotificationService) CreateNotification(ctx context.Context, req *CreateNotificationRequest) (*Notification, error) {
	// Marshal data and channels to JSON
	dataJSON, _ := json.Marshal(req.Data)
	channelsJSON, _ := json.Marshal(req.Channels)
	
	notification := &Notification{
		Type:          req.Type,
		Priority:      req.Priority,
		Title:         req.Title,
		Message:       req.Message,
		Data:          string(dataJSON),
		PropertyID:    req.PropertyID,
		RecipientID:   req.RecipientID,
		RecipientRole: req.RecipientRole,
		Channels:      string(channelsJSON),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Set expiration if not provided
	if req.ExpiresIn > 0 {
		expiresAt := time.Now().Add(req.ExpiresIn)
		notification.ExpiresAt = &expiresAt
	}

	// Save to database
	if err := s.db.Create(notification).Error; err != nil {
		logging.WithError(err).Error("Failed to create notification")
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}

	// Send notification through requested channels
	if err := s.sendNotification(ctx, notification); err != nil {
		logging.WithError(err).Warn("Failed to send notification")
		// Don't return error - notification is created, just not sent
	}

	return notification, nil
}

// sendNotification sends the notification through all requested channels
func (s *NotificationService) sendNotification(ctx context.Context, notification *Notification) error {
	var lastErr error
	sent := false

	// Unmarshal channels from JSON
	var channels []NotificationChannel
	if err := json.Unmarshal([]byte(notification.Channels), &channels); err != nil {
		return fmt.Errorf("failed to unmarshal channels: %w", err)
	}

	for _, channel := range channels {
		switch channel {
		case ChannelPush:
			if err := s.sendPushNotification(ctx, notification); err != nil {
				lastErr = err
				logging.WithError(err).WithField("channel", "push").Warn("Failed to send push notification")
			} else {
				sent = true
			}

		case ChannelEmail:
			if err := s.sendEmailNotification(ctx, notification); err != nil {
				lastErr = err
				logging.WithError(err).WithField("channel", "email").Warn("Failed to send email notification")
			} else {
				sent = true
			}

		case ChannelSMS:
			if err := s.sendSMSNotification(ctx, notification); err != nil {
				lastErr = err
				logging.WithError(err).WithField("channel", "sms").Warn("Failed to send SMS notification")
			} else {
				sent = true
			}

		case ChannelWebSocket:
			if err := s.sendWebSocketNotification(ctx, notification); err != nil {
				lastErr = err
				logging.WithError(err).WithField("channel", "websocket").Warn("Failed to send WebSocket notification")
			} else {
				sent = true
			}
		}
	}

	// Update notification status
	if sent {
		now := time.Now()
		notification.Sent = true
		notification.SentAt = &now
		s.db.Model(notification).Updates(map[string]interface{}{
			"sent":    true,
			"sent_at": now,
		})
	}

	if !sent && lastErr != nil {
		return lastErr
	}

	return nil
}

// sendPushNotification sends a push notification
func (s *NotificationService) sendPushNotification(ctx context.Context, notification *Notification) error {
	if s.pushProvider == nil {
		return fmt.Errorf("push notification provider not configured")
	}

	// Get user device tokens
	var tokens []string
	if notification.RecipientID > 0 {
		// Get specific user's tokens
		var devices []models.UserDevice
		if err := s.db.Where("user_id = ? AND push_enabled = ?", notification.RecipientID, true).Find(&devices).Error; err != nil {
			return fmt.Errorf("failed to get user devices: %w", err)
		}
		for _, device := range devices {
			tokens = append(tokens, device.PushToken)
		}
	} else if notification.RecipientRole != "" {
		// Get all users with role
		var devices []models.UserDevice
		if err := s.db.
			Joins("JOIN staffs ON staffs.id = user_devices.user_id").
			Where("staffs.role = ? AND user_devices.push_enabled = ?", notification.RecipientRole, true).
			Find(&devices).Error; err != nil {
			return fmt.Errorf("failed to get role devices: %w", err)
		}
		for _, device := range devices {
			tokens = append(tokens, device.PushToken)
		}
	}

	if len(tokens) == 0 {
		return fmt.Errorf("no push tokens found for notification")
	}

	// Unmarshal data for push notification
	var data map[string]interface{}
	if notification.Data != "" {
		json.Unmarshal([]byte(notification.Data), &data)
	}

	// Create push notification payload
	push := &PushNotification{
		Title:    notification.Title,
		Body:     notification.Message,
		Data:     data,
		Priority: string(notification.Priority),
		Sound:    "default",
	}

	// Send to all tokens
	return s.pushProvider.SendBatch(ctx, tokens, push)
}

// sendEmailNotification sends an email notification
func (s *NotificationService) sendEmailNotification(ctx context.Context, notification *Notification) error {
	if s.emailProvider == nil {
		return fmt.Errorf("email provider not configured")
	}

	// Get recipient emails
	var emails []string
	if notification.RecipientID > 0 {
		var staff models.Staff
		if err := s.db.First(&staff, notification.RecipientID).Error; err != nil {
			return fmt.Errorf("failed to get staff: %w", err)
		}
		emails = append(emails, staff.Email)
	} else if notification.RecipientRole != "" {
		var staffList []models.Staff
		if err := s.db.Where("role = ?", notification.RecipientRole).Find(&staffList).Error; err != nil {
			return fmt.Errorf("failed to get staff by role: %w", err)
		}
		for _, staff := range staffList {
			emails = append(emails, staff.Email)
		}
	}

	// Send to all emails
	for _, email := range emails {
		if err := s.emailProvider.Send(ctx, email, notification.Title, notification.Message); err != nil {
			logging.WithError(err).WithField("email", email).Warn("Failed to send email")
		}
	}

	return nil
}

// sendSMSNotification sends an SMS notification
func (s *NotificationService) sendSMSNotification(ctx context.Context, notification *Notification) error {
	if s.smsProvider == nil {
		return fmt.Errorf("SMS provider not configured")
	}

	// Only send SMS for high priority notifications
	if notification.Priority != PriorityHigh && notification.Priority != PriorityCritical {
		return nil
	}

	// Note: SMS functionality requires phone field in Staff model
	// For now, we'll just log the SMS that would be sent

	// Send SMS
	message := fmt.Sprintf("%s: %s", notification.Title, notification.Message)
	// Note: SMS functionality requires phone field in Staff model
	logging.Info("SMS notification would be sent: " + message)

	return nil
}

// sendWebSocketNotification sends a real-time WebSocket notification
func (s *NotificationService) sendWebSocketNotification(ctx context.Context, notification *Notification) error {
	if s.wsHub == nil {
		return fmt.Errorf("WebSocket hub not configured")
	}

	// Publish to Redis for distributed WebSocket
	if s.cache != nil {
		channel := fmt.Sprintf("notifications:%s", notification.PropertyID)
		if err := s.cache.Publish(ctx, channel, notification); err != nil {
			logging.WithError(err).Warn("Failed to publish notification to Redis")
		}
	}

	return nil
}

// NotifyVIPArrival sends notifications for VIP guest arrival
func (s *NotificationService) NotifyVIPArrival(ctx context.Context, guest *models.Guest) error {
	data := map[string]interface{}{
		"guest_id":    guest.ID,
		"guest_name":  fmt.Sprintf("%s %s", guest.FirstName, guest.LastName),
		"room_number": guest.RoomNumber,
		"check_in":    guest.CheckInDate,
	}

	req := &CreateNotificationRequest{
		Type:          NotificationVIPArrival,
		Priority:      PriorityHigh,
		Title:         "VIP Guest Arrival",
		Message:       fmt.Sprintf("VIP guest %s %s has arrived in room %s", guest.FirstName, guest.LastName, guest.RoomNumber),
		Data:          data,
		PropertyID:    guest.PropertyID,
		RecipientRole: "manager",
		Channels:      []NotificationChannel{ChannelPush, ChannelWebSocket},
	}

	_, err := s.CreateNotification(ctx, req)
	return err
}

// NotifyUpsetGuest sends notifications for upset guest situations
func (s *NotificationService) NotifyUpsetGuest(ctx context.Context, guest *models.Guest, issue string) error {
	data := map[string]interface{}{
		"guest_id":    guest.ID,
		"guest_name":  fmt.Sprintf("%s %s", guest.FirstName, guest.LastName),
		"room_number": guest.RoomNumber,
		"issue":       issue,
		"is_vip":      guest.IsVIP,
	}

	priority := PriorityHigh
	if guest.IsVIP {
		priority = PriorityCritical
	}

	req := &CreateNotificationRequest{
		Type:          NotificationUpsetGuest,
		Priority:      priority,
		Title:         "Guest Requires Attention",
		Message:       fmt.Sprintf("Guest %s %s in room %s is upset: %s", guest.FirstName, guest.LastName, guest.RoomNumber, issue),
		Data:          data,
		PropertyID:    guest.PropertyID,
		RecipientRole: "manager",
		Channels:      []NotificationChannel{ChannelPush, ChannelWebSocket, ChannelSMS},
		ExpiresIn:     30 * time.Minute,
	}

	_, err := s.CreateNotification(ctx, req)
	return err
}

// NotifyServiceDelay sends notifications for service delays
func (s *NotificationService) NotifyServiceDelay(ctx context.Context, propertyID string, avgServiceTime float64) error {
	data := map[string]interface{}{
		"avg_service_time": avgServiceTime,
		"threshold":        15.0,
	}

	req := &CreateNotificationRequest{
		Type:          NotificationServiceDelay,
		Priority:      PriorityMedium,
		Title:         "Service Delay Alert",
		Message:       fmt.Sprintf("Average service time is %.1f minutes, exceeding 15-minute threshold", avgServiceTime),
		Data:          data,
		PropertyID:    propertyID,
		RecipientRole: "staff",
		Channels:      []NotificationChannel{ChannelPush, ChannelWebSocket},
	}

	_, err := s.CreateNotification(ctx, req)
	return err
}

// GetUnreadNotifications gets unread notifications for a user
func (s *NotificationService) GetUnreadNotifications(userID uint) ([]*Notification, error) {
	var notifications []*Notification
	err := s.db.
		Where("recipient_id = ? AND read = ? AND (expires_at IS NULL OR expires_at > ?)", 
			userID, false, time.Now()).
		Order("created_at DESC").
		Find(&notifications).Error
	return notifications, err
}

// MarkNotificationRead marks a notification as read
func (s *NotificationService) MarkNotificationRead(notificationID uint, userID uint) error {
	now := time.Now()
	return s.db.Model(&Notification{}).
		Where("id = ? AND recipient_id = ?", notificationID, userID).
		Updates(map[string]interface{}{
			"read":    true,
			"read_at": now,
		}).Error
}

// GetNotificationPreferences gets user notification preferences
func (s *NotificationService) GetNotificationPreferences(userID uint) (*NotificationPreference, error) {
	var pref NotificationPreference
	err := s.db.Where("user_id = ?", userID).First(&pref).Error
	if err == gorm.ErrRecordNotFound {
		// Return default preferences
		return &NotificationPreference{
			UserID:          userID,
			EnablePush:      true,
			EnableEmail:     true,
			EnableSMS:       false,
			QuietHoursStart: "22:00",
			QuietHoursEnd:   "07:00",
			MinimumPriority: string(PriorityMedium),
		}, nil
	}
	return &pref, err
}

// UpdateNotificationPreferences updates user notification preferences
func (s *NotificationService) UpdateNotificationPreferences(userID uint, pref *NotificationPreference) error {
	pref.UserID = userID
	return s.db.Save(pref).Error
}

// CreateNotificationRequest represents a request to create a notification
type CreateNotificationRequest struct {
	Type          NotificationType      `json:"type"`
	Priority      NotificationPriority  `json:"priority"`
	Title         string               `json:"title"`
	Message       string               `json:"message"`
	Data          map[string]interface{} `json:"data,omitempty"`
	PropertyID    string               `json:"property_id"`
	RecipientID   uint                 `json:"recipient_id,omitempty"`
	RecipientRole string               `json:"recipient_role,omitempty"`
	Channels      []NotificationChannel `json:"channels"`
	ExpiresIn     time.Duration        `json:"expires_in,omitempty"`
}

// NotificationStats represents notification statistics
type NotificationStats struct {
	TotalSent      int64 `json:"total_sent"`
	TotalRead      int64 `json:"total_read"`
	TotalUnread    int64 `json:"total_unread"`
	ByType         map[string]int64 `json:"by_type"`
	ByPriority     map[string]int64 `json:"by_priority"`
	ByChannel      map[string]int64 `json:"by_channel"`
	AverageReadTime time.Duration `json:"average_read_time"`
}

// GetNotificationStats returns notification statistics
func (s *NotificationService) GetNotificationStats(propertyID string, period time.Duration) (*NotificationStats, error) {
	since := time.Now().Add(-period)
	stats := &NotificationStats{
		ByType:     make(map[string]int64),
		ByPriority: make(map[string]int64),
		ByChannel:  make(map[string]int64),
	}

	// Get basic counts
	s.db.Model(&Notification{}).Where("property_id = ? AND created_at > ?", propertyID, since).Count(&stats.TotalSent)
	s.db.Model(&Notification{}).Where("property_id = ? AND created_at > ? AND read = ?", propertyID, since, true).Count(&stats.TotalRead)
	stats.TotalUnread = stats.TotalSent - stats.TotalRead

	// Get counts by type
	var typeCounts []struct {
		Type  string
		Count int64
	}
	s.db.Model(&Notification{}).
		Select("type, COUNT(*) as count").
		Where("property_id = ? AND created_at > ?", propertyID, since).
		Group("type").
		Scan(&typeCounts)
	
	for _, tc := range typeCounts {
		stats.ByType[tc.Type] = tc.Count
	}

	// Calculate average read time
	var avgReadSeconds float64
	s.db.Model(&Notification{}).
		Select("AVG(JULIANDAY(read_at) - JULIANDAY(created_at)) * 86400").
		Where("property_id = ? AND created_at > ? AND read = ?", propertyID, since, true).
		Scan(&avgReadSeconds)
	
	stats.AverageReadTime = time.Duration(avgReadSeconds) * time.Second

	return stats, nil
}

// RegisterDevice registers a new device for push notifications
func (s *NotificationService) RegisterDevice(device *UserDevice) error {
	// Update existing device or create new one
	existing := &UserDevice{}
	err := s.db.Where("device_id = ?", device.DeviceID).First(existing).Error
	
	if err == gorm.ErrRecordNotFound {
		// Create new device
		return s.db.Create(device).Error
	} else if err != nil {
		return err
	}
	
	// Update existing device
	existing.UserID = device.UserID
	existing.DeviceType = device.DeviceType
	existing.DeviceName = device.DeviceName
	existing.PushToken = device.PushToken
	existing.PushEnabled = device.PushEnabled
	existing.LastActiveAt = device.LastActiveAt
	
	return s.db.Save(existing).Error
}

// UpdateDeviceToken updates the push token for a device
func (s *NotificationService) UpdateDeviceToken(userID uint, deviceID, pushToken string) error {
	return s.db.Model(&UserDevice{}).
		Where("user_id = ? AND device_id = ?", userID, deviceID).
		Updates(map[string]interface{}{
			"push_token":     pushToken,
			"last_active_at": time.Now(),
		}).Error
}

// GetUserNotifications gets notifications for a user with pagination
func (s *NotificationService) GetUserNotifications(userID uint, unreadOnly bool, limit, offset int) ([]*Notification, int64, error) {
	var notifications []*Notification
	var total int64
	
	query := s.db.Model(&Notification{}).Where("recipient_id = ?", userID)
	
	if unreadOnly {
		query = query.Where("read = ?", false)
	}
	
	// Get total count
	query.Count(&total)
	
	// Get paginated results
	err := query.
		Where("expires_at IS NULL OR expires_at > ?", time.Now()).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&notifications).Error
		
	return notifications, total, err
}

// MarkAllNotificationsRead marks all notifications as read for a user
func (s *NotificationService) MarkAllNotificationsRead(userID uint) error {
	now := time.Now()
	return s.db.Model(&Notification{}).
		Where("recipient_id = ? AND read = ?", userID, false).
		Updates(map[string]interface{}{
			"read":    true,
			"read_at": now,
		}).Error
}

// UserDevice type alias for models
type UserDevice = models.UserDevice