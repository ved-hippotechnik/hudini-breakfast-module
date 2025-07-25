package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"hudini-breakfast-module/internal/audit"
	"hudini-breakfast-module/internal/models"

	"gorm.io/gorm"
)

// AuditService handles audit logging for system events
type AuditService struct {
	db *gorm.DB
}

// NewAuditService creates a new audit service
func NewAuditService(db *gorm.DB) *AuditService {
	return &AuditService{
		db: db,
	}
}


// LogEntry creates a new audit log entry
func (s *AuditService) LogEntry(ctx context.Context, entry *models.AuditLog) error {
	// Ensure created_at is set
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now().UTC()
	}

	// Create the audit log entry
	if err := s.db.WithContext(ctx).Create(entry).Error; err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	return nil
}

// LogSuccess logs a successful action
func (s *AuditService) LogSuccess(ctx context.Context, userID *uint, action audit.AuditAction, resource audit.AuditResource, resourceID string, ipAddress, userAgent string, oldValues, newValues interface{}) error {
	entry := &models.AuditLog{
		UserID:     userID,
		Action:     string(action),
		Resource:   string(resource),
		ResourceID: resourceID,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		Status:     "success",
	}

	// Convert old values to JSON if provided
	if oldValues != nil {
		oldJSON, err := json.Marshal(oldValues)
		if err != nil {
			return fmt.Errorf("failed to marshal old values: %w", err)
		}
		entry.OldValues = string(oldJSON)
	}

	// Convert new values to JSON if provided
	if newValues != nil {
		newJSON, err := json.Marshal(newValues)
		if err != nil {
			return fmt.Errorf("failed to marshal new values: %w", err)
		}
		entry.NewValues = string(newJSON)
	}

	return s.LogEntry(ctx, entry)
}

// LogFailure logs a failed action
func (s *AuditService) LogFailure(ctx context.Context, userID *uint, action audit.AuditAction, resource audit.AuditResource, resourceID string, ipAddress, userAgent string, err error) error {
	entry := &models.AuditLog{
		UserID:     userID,
		Action:     string(action),
		Resource:   string(resource),
		ResourceID: resourceID,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		Status:     "failed",
		Error:      err.Error(),
	}

	return s.LogEntry(ctx, entry)
}

// GetAuditLogs retrieves audit logs with filters
func (s *AuditService) GetAuditLogs(ctx context.Context, filters AuditFilters) ([]models.AuditLog, int64, error) {
	query := s.db.WithContext(ctx).Model(&models.AuditLog{})

	// Apply filters
	if filters.UserID != nil {
		query = query.Where("user_id = ?", *filters.UserID)
	}
	if filters.Action != "" {
		query = query.Where("action = ?", filters.Action)
	}
	if filters.Resource != "" {
		query = query.Where("resource = ?", filters.Resource)
	}
	if filters.ResourceID != "" {
		query = query.Where("resource_id = ?", filters.ResourceID)
	}
	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}
	if filters.IPAddress != "" {
		query = query.Where("ip_address = ?", filters.IPAddress)
	}
	if !filters.StartDate.IsZero() {
		query = query.Where("created_at >= ?", filters.StartDate)
	}
	if !filters.EndDate.IsZero() {
		query = query.Where("created_at <= ?", filters.EndDate)
	}

	// Count total records
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count audit logs: %w", err)
	}

	// Apply sorting
	sortOrder := "DESC"
	if filters.SortOrder == "asc" {
		sortOrder = "ASC"
	}
	query = query.Order(fmt.Sprintf("created_at %s", sortOrder))

	// Apply pagination
	if filters.Limit > 0 {
		query = query.Limit(filters.Limit)
	}
	if filters.Offset > 0 {
		query = query.Offset(filters.Offset)
	}

	// Preload user information
	query = query.Preload("User")

	// Execute query
	var logs []models.AuditLog
	if err := query.Find(&logs).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to retrieve audit logs: %w", err)
	}

	return logs, total, nil
}

// AuditFilters represents filters for querying audit logs
type AuditFilters struct {
	UserID     *uint
	Action     string
	Resource   string
	ResourceID string
	Status     string
	IPAddress  string
	StartDate  time.Time
	EndDate    time.Time
	Limit      int
	Offset     int
	SortOrder  string // asc or desc
}

// GetUserActivity retrieves audit logs for a specific user
func (s *AuditService) GetUserActivity(ctx context.Context, userID uint, limit int) ([]models.AuditLog, error) {
	var logs []models.AuditLog
	
	err := s.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error
		
	if err != nil {
		return nil, fmt.Errorf("failed to get user activity: %w", err)
	}
	
	return logs, nil
}

// GetResourceHistory retrieves audit logs for a specific resource
func (s *AuditService) GetResourceHistory(ctx context.Context, resource audit.AuditResource, resourceID string) ([]models.AuditLog, error) {
	var logs []models.AuditLog
	
	err := s.db.WithContext(ctx).
		Where("resource = ? AND resource_id = ?", string(resource), resourceID).
		Order("created_at DESC").
		Preload("User").
		Find(&logs).Error
		
	if err != nil {
		return nil, fmt.Errorf("failed to get resource history: %w", err)
	}
	
	return logs, nil
}

// CleanupOldLogs removes audit logs older than the specified duration
func (s *AuditService) CleanupOldLogs(ctx context.Context, olderThan time.Duration) (int64, error) {
	cutoffDate := time.Now().Add(-olderThan)
	
	result := s.db.WithContext(ctx).
		Where("created_at < ?", cutoffDate).
		Delete(&models.AuditLog{})
		
	if result.Error != nil {
		return 0, fmt.Errorf("failed to cleanup old logs: %w", result.Error)
	}
	
	return result.RowsAffected, nil
}