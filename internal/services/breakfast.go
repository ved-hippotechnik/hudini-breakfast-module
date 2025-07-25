package services

import (
	"errors"
	"fmt"
	"time"

	"hudini-breakfast-module/internal/logging"
	"hudini-breakfast-module/internal/models"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type BreakfastService struct {
	db          *gorm.DB
	ohipService *OHIPService
}

func NewBreakfastService(db *gorm.DB, ohipService *OHIPService) *BreakfastService {
	return &BreakfastService{
		db:          db,
		ohipService: ohipService,
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
