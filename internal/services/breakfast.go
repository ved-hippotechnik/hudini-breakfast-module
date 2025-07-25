package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"hudini-breakfast-module/internal/cache"
	"hudini-breakfast-module/internal/logging"
	"hudini-breakfast-module/internal/models"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type BreakfastService struct {
	db          *gorm.DB
	ohipService *OHIPService
	vipCache    *cache.VIPCache
}

func NewBreakfastService(db *gorm.DB, ohipService *OHIPService) *BreakfastService {
	return &BreakfastService{
		db:          db,
		ohipService: ohipService,
	}
}

// NewBreakfastServiceWithCache creates a new breakfast service with caching
func NewBreakfastServiceWithCache(db *gorm.DB, ohipService *OHIPService, vipCache *cache.VIPCache) *BreakfastService {
	return &BreakfastService{
		db:          db,
		ohipService: ohipService,
		vipCache:    vipCache,
	}
}

// Room Grid Management
func (s *BreakfastService) GetRoomBreakfastStatus(propertyID string) ([]models.RoomBreakfastStatus, error) {
	logging.WithFields(logrus.Fields{
		"service":     "BreakfastService",
		"method":      "GetRoomBreakfastStatus",
		"property_id": propertyID,
	}).Debug("Fetching room breakfast status")

	var roomStatuses []models.RoomBreakfastStatus

	query := `
		SELECT DISTINCT
			r.property_id,
			r.room_number,
			r.floor,
			r.room_type,
			r.status,
			CASE WHEN g.id IS NOT NULL THEN true ELSE false END as has_guest,
			COALESCE(g.first_name || ' ' || g.last_name, '') as guest_name,
			COALESCE(g.breakfast_package, false) as breakfast_package,
			COALESCE(g.breakfast_count, 0) as breakfast_count,
			CASE WHEN dbc.id IS NOT NULL THEN true ELSE false END as consumed_today,
			dbc.consumed_at,
			COALESCE(s.first_name || ' ' || s.last_name, '') as consumed_by,
			g.check_in_date,
			g.check_out_date,
			COALESCE(g.is_vip, false) as is_vip,
			COALESCE(g.is_upset, false) as is_upset,
			COALESCE(g.pms_special_requests, '') as special_requests
		FROM rooms r
		LEFT JOIN (
			SELECT DISTINCT room_number, property_id, first_name, last_name, 
				breakfast_package, breakfast_count, check_in_date, check_out_date, id,
				is_vip, is_upset, pms_special_requests
			FROM guests 
			WHERE is_active = true
				AND DATE(check_in_date) <= DATE('now') 
				AND DATE(check_out_date) >= DATE('now')
		) g ON r.room_number = g.room_number AND r.property_id = g.property_id
		LEFT JOIN (
			SELECT DISTINCT room_number, property_id, consumed_at, consumed_by, id
			FROM daily_breakfast_consumptions 
			WHERE DATE(consumption_date) = DATE('now') AND status = 'consumed'
		) dbc ON r.room_number = dbc.room_number AND r.property_id = dbc.property_id
		LEFT JOIN staffs s ON dbc.consumed_by = s.id
		WHERE r.property_id = ?
		ORDER BY r.room_number
	`

	err := s.db.Raw(query, propertyID).Scan(&roomStatuses).Error
	if err != nil {
		logging.WithFields(logrus.Fields{
			"service":     "BreakfastService",
			"method":      "GetRoomBreakfastStatus",
			"property_id": propertyID,
			"error":       err.Error(),
		}).Error("Failed to fetch room breakfast status from database")
		return nil, fmt.Errorf("failed to fetch room breakfast status: %w", err)
	}

	logging.WithFields(logrus.Fields{
		"service":     "BreakfastService",
		"method":      "GetRoomBreakfastStatus",
		"property_id": propertyID,
		"room_count":  len(roomStatuses),
	}).Debug("Successfully fetched room breakfast status")

	return roomStatuses, nil
}

func (s *BreakfastService) MarkBreakfastConsumed(propertyID, roomNumber string, staffID uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Check if guest exists and has breakfast package
		var guest models.Guest
		err := tx.Where("property_id = ? AND room_number = ? AND is_active = true", propertyID, roomNumber).First(&guest).Error
		if err != nil {
			return errors.New("no active guest found in this room")
		}

		if !guest.BreakfastPackage {
			return errors.New("guest does not have breakfast package")
		}

		// Check if already consumed today
		var existing models.DailyBreakfastConsumption
		today := time.Now().Format("2006-01-02")
		err = tx.Where("property_id = ? AND room_number = ? AND DATE(consumption_date) = ?",
			propertyID, roomNumber, today).First(&existing).Error

		if err == nil && existing.Status == "consumed" {
			return errors.New("breakfast already consumed today")
		}

		// Create or update consumption record
		now := time.Now()
		consumption := models.DailyBreakfastConsumption{
			PropertyID:      propertyID,
			RoomNumber:      roomNumber,
			GuestID:         guest.ID,
			ConsumptionDate: now,
			ConsumedAt:      &now,
			ConsumedBy:      &staffID,
			Status:          "consumed",
			PaymentMethod:   "room_charge",
			Amount:          25.00, // Default breakfast price
		}

		if err == gorm.ErrRecordNotFound {
			// Create new record
			return tx.Create(&consumption).Error
		} else {
			// Update existing record
			return tx.Model(&existing).Updates(consumption).Error
		}
	})
}

// Consumption History Management
func (s *BreakfastService) GetConsumptionHistory(propertyID string, startDate, endDate time.Time) ([]models.DailyBreakfastConsumption, error) {
	var consumptions []models.DailyBreakfastConsumption
	err := s.db.Preload("Guest").
		Preload("Staff").
		Preload("Room").
		Where("property_id = ? AND consumption_date BETWEEN ? AND ?", propertyID, startDate, endDate).
		Order("consumption_date DESC, consumed_at DESC").
		Find(&consumptions).Error
	return consumptions, err
}

func (s *BreakfastService) GetDailyReport(propertyID string, date time.Time) (*DailyBreakfastReport, error) {
	var report DailyBreakfastReport
	dateStr := date.Format("2006-01-02")

	// Get total rooms with breakfast packages
	var totalWithBreakfast int64
	err := s.db.Model(&models.Guest{}).
		Where("property_id = ? AND breakfast_package = true AND is_active = true", propertyID).
		Where("DATE(check_in_date) <= ? AND DATE(check_out_date) >= ?", dateStr, dateStr).
		Count(&totalWithBreakfast).Error
	if err != nil {
		return nil, err
	}

	// Get consumed count
	var consumed int64
	err = s.db.Model(&models.DailyBreakfastConsumption{}).
		Where("property_id = ? AND DATE(consumption_date) = ? AND status = 'consumed'", propertyID, dateStr).
		Count(&consumed).Error
	if err != nil {
		return nil, err
	}

	// Get OHIP covered count
	var ohipCovered int64
	err = s.db.Model(&models.DailyBreakfastConsumption{}).
		Where("property_id = ? AND DATE(consumption_date) = ? AND ohip_covered = true", propertyID, dateStr).
		Count(&ohipCovered).Error
	if err != nil {
		return nil, err
	}

	// Get PMS posted count
	var pmsPosted int64
	err = s.db.Model(&models.DailyBreakfastConsumption{}).
		Where("property_id = ? AND DATE(consumption_date) = ? AND pms_posted = true", propertyID, dateStr).
		Count(&pmsPosted).Error
	if err != nil {
		return nil, err
	}

	report.Date = date
	report.TotalRoomsWithBreakfast = int(totalWithBreakfast)
	report.TotalConsumed = int(consumed)
	report.TotalNotConsumed = int(totalWithBreakfast - consumed)
	report.ConsumptionRate = 0
	if totalWithBreakfast > 0 {
		report.ConsumptionRate = float64(consumed) / float64(totalWithBreakfast) * 100
	}
	report.OHIPCoveredCount = int(ohipCovered)
	report.PMSChargesPosted = int(pmsPosted)

	return &report, nil
}

func (s *BreakfastService) GetAnalyticsData(propertyID string, period string) (*BreakfastAnalytics, error) {
	var analytics BreakfastAnalytics

	// Calculate date range based on period
	now := time.Now()
	var startDate time.Time

	switch period {
	case "today":
		startDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	case "week":
		startDate = now.AddDate(0, 0, -7)
	case "month":
		startDate = now.AddDate(0, -1, 0)
	default:
		startDate = now.AddDate(0, 0, -7)
	}

	// Get daily breakdown
	query := `
		SELECT 
			DATE(consumption_date) as date,
			COUNT(CASE WHEN status = 'consumed' THEN 1 END) as consumed,
			COUNT(*) as total,
			ROUND(COUNT(CASE WHEN status = 'consumed' THEN 1 END) * 100.0 / COUNT(*), 2) as rate
		FROM daily_breakfast_consumptions 
		WHERE property_id = ? AND consumption_date >= ?
		GROUP BY DATE(consumption_date)
		ORDER BY date
	`

	var dailyTrend []DailyTrend
	err := s.db.Raw(query, propertyID, startDate).Scan(&dailyTrend).Error
	if err != nil {
		return nil, err
	}

	analytics.DailyTrend = dailyTrend
	analytics.Period = period

	return &analytics, nil
}

// Helper types for reports and analytics
type DailyBreakfastReport struct {
	Date                    time.Time `json:"date"`
	TotalRoomsWithBreakfast int       `json:"total_rooms_with_breakfast"`
	TotalConsumed           int       `json:"total_consumed"`
	TotalNotConsumed        int       `json:"total_not_consumed"`
	ConsumptionRate         float64   `json:"consumption_rate"`
	OHIPCoveredCount        int       `json:"ohip_covered_count"`
	PMSChargesPosted        int       `json:"pms_charges_posted"`
}

type BreakfastAnalytics struct {
	Period     string       `json:"period"`
	DailyTrend []DailyTrend `json:"daily_trend"`
}

type DailyTrend struct {
	Date     string  `json:"date"`
	Consumed int     `json:"consumed"`
	Total    int     `json:"total"`
	Rate     float64 `json:"rate"`
}

// GetVIPGuests retrieves VIP guests with caching
func (s *BreakfastService) GetVIPGuests(ctx context.Context, propertyID string) ([]models.Guest, error) {
	// Try cache first if available
	if s.vipCache != nil {
		cached, err := s.vipCache.GetVIPGuests(ctx, propertyID)
		if err == nil && cached != nil {
			logging.WithField("property_id", propertyID).Debug("VIP guests retrieved from cache")
			return cached, nil
		}
	}

	// Query database
	var guests []models.Guest
	err := s.db.Where("property_id = ? AND is_vip = ? AND is_active = ?", propertyID, true, true).
		Preload("Room").
		Find(&guests).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to fetch VIP guests: %w", err)
	}

	// Cache the results
	if s.vipCache != nil {
		if err := s.vipCache.SetVIPGuests(ctx, propertyID, guests); err != nil {
			logging.WithError(err).Warn("Failed to cache VIP guests")
		}
	}

	return guests, nil
}

// GetUpsetGuests retrieves upset guests with caching
func (s *BreakfastService) GetUpsetGuests(ctx context.Context, propertyID string) ([]models.Guest, error) {
	// Try cache first if available
	if s.vipCache != nil {
		cached, err := s.vipCache.GetUpsetGuests(ctx, propertyID)
		if err == nil && cached != nil {
			logging.WithField("property_id", propertyID).Debug("Upset guests retrieved from cache")
			return cached, nil
		}
	}

	// Query database
	var guests []models.Guest
	err := s.db.Where("property_id = ? AND is_upset = ? AND is_active = ?", propertyID, true, true).
		Preload("Room").
		Find(&guests).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to fetch upset guests: %w", err)
	}

	// Cache the results
	if s.vipCache != nil {
		if err := s.vipCache.SetUpsetGuests(ctx, propertyID, guests); err != nil {
			logging.WithError(err).Warn("Failed to cache upset guests")
		}
	}

	return guests, nil
}

// GetVIPMetrics retrieves VIP metrics with caching
func (s *BreakfastService) GetVIPMetrics(ctx context.Context, propertyID string) (*cache.VIPMetrics, error) {
	// Try cache first if available
	if s.vipCache != nil {
		cached, err := s.vipCache.GetVIPMetrics(ctx, propertyID)
		if err == nil && cached != nil {
			logging.WithField("property_id", propertyID).Debug("VIP metrics retrieved from cache")
			return cached, nil
		}
	}

	// Calculate metrics from database
	var metrics cache.VIPMetrics
	
	// Count VIPs and upset guests
	var totalVIPs, totalUpset int64
	s.db.Model(&models.Guest{}).Where("property_id = ? AND is_vip = ? AND is_active = ?", propertyID, true, true).Count(&totalVIPs)
	s.db.Model(&models.Guest{}).Where("property_id = ? AND is_upset = ? AND is_active = ?", propertyID, true, true).Count(&totalUpset)
	metrics.TotalVIPs = int(totalVIPs)
	metrics.TotalUpset = int(totalUpset)
	
	// Calculate VIP breakfast consumption rate
	var vipConsumed, vipTotal int64
	s.db.Model(&models.DailyBreakfastConsumption{}).
		Joins("JOIN guests ON guests.id = daily_breakfast_consumptions.guest_id").
		Where("guests.property_id = ? AND guests.is_vip = ? AND daily_breakfast_consumptions.consumption_date = ?", 
			propertyID, true, time.Now().Format("2006-01-02")).
		Where("daily_breakfast_consumptions.status = ?", "consumed").
		Count(&vipConsumed)
	
	s.db.Model(&models.DailyBreakfastConsumption{}).
		Joins("JOIN guests ON guests.id = daily_breakfast_consumptions.guest_id").
		Where("guests.property_id = ? AND guests.is_vip = ? AND daily_breakfast_consumptions.consumption_date = ?", 
			propertyID, true, time.Now().Format("2006-01-02")).
		Count(&vipTotal)
	
	if vipTotal > 0 {
		metrics.VIPBreakfastRate = float64(vipConsumed) / float64(vipTotal) * 100
	}
	
	// Calculate average stay duration for VIP guests
	var avgStay float64
	s.db.Model(&models.Guest{}).
		Select("AVG(JULIANDAY(check_out_date) - JULIANDAY(check_in_date))").
		Where("property_id = ? AND is_vip = ? AND is_active = ?", propertyID, true, true).
		Scan(&avgStay)
	metrics.AverageStayDuration = avgStay
	
	// Get top preferences (simplified for now)
	metrics.TopPreferences = []string{"Window Seating", "Continental Breakfast", "Early Service"}
	metrics.LastUpdated = time.Now()
	
	// Cache the results
	if s.vipCache != nil {
		if err := s.vipCache.SetVIPMetrics(ctx, propertyID, &metrics); err != nil {
			logging.WithError(err).Warn("Failed to cache VIP metrics")
		}
	}

	return &metrics, nil
}

// InvalidateGuestCache invalidates cache when guest data changes
func (s *BreakfastService) InvalidateGuestCache(ctx context.Context, guestID uint, propertyID string) {
	if s.vipCache != nil {
		if err := s.vipCache.InvalidateGuestCache(ctx, guestID, propertyID); err != nil {
			logging.WithError(err).Warn("Failed to invalidate guest cache")
		}
	}
}
