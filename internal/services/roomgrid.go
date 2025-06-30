package services

import (
	"fmt"
	"time"

	"hudini-breakfast-module/internal/models"

	"gorm.io/gorm"
)

type RoomGridService struct {
	db          *gorm.DB
	pmsService  *PMSService
	ohipService *OHIPService
}

func NewRoomGridService(db *gorm.DB, pmsService *PMSService, ohipService *OHIPService) *RoomGridService {
	return &RoomGridService{
		db:          db,
		pmsService:  pmsService,
		ohipService: ohipService,
	}
}

// GetRoomGrid returns the breakfast status for all rooms in a property
func (s *RoomGridService) GetRoomGrid(propertyID string, date time.Time) ([]models.RoomBreakfastStatus, error) {
	// Get all rooms for the property
	var rooms []models.Room
	if err := s.db.Where("property_id = ?", propertyID).Find(&rooms).Error; err != nil {
		return nil, fmt.Errorf("failed to get rooms: %w", err)
	}

	var roomStatuses []models.RoomBreakfastStatus
	dateOnly := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())

	for _, room := range rooms {
		status := models.RoomBreakfastStatus{
			PropertyID: propertyID,
			RoomNumber: room.RoomNumber,
			Floor:      room.Floor,
			RoomType:   room.RoomType,
			Status:     room.Status,
		}

		// Get current guest for this room
		var guest models.Guest
		err := s.db.Where("room_number = ? AND property_id = ? AND is_active = ? AND check_in_date <= ? AND check_out_date > ?",
			room.RoomNumber, propertyID, true, dateOnly, dateOnly).
			First(&guest).Error

		if err == nil {
			// Room has an active guest
			status.HasGuest = true
			status.GuestName = fmt.Sprintf("%s %s", guest.FirstName, guest.LastName)
			status.BreakfastPackage = guest.BreakfastPackage
			status.BreakfastCount = guest.BreakfastCount
			status.CheckInDate = &guest.CheckInDate
			status.CheckOutDate = &guest.CheckOutDate

			// Check if breakfast was consumed today
			var consumption models.DailyBreakfastConsumption
			err = s.db.Preload("Staff").
				Where("room_number = ? AND property_id = ? AND consumption_date = ?",
					room.RoomNumber, propertyID, dateOnly).
				First(&consumption).Error

			if err == nil {
				status.ConsumedToday = consumption.Status == "consumed"
				status.ConsumedAt = consumption.ConsumedAt
				if consumption.Staff != nil {
					status.ConsumedBy = fmt.Sprintf("%s %s", consumption.Staff.FirstName, consumption.Staff.LastName)
				}
			}
		}

		roomStatuses = append(roomStatuses, status)
	}

	return roomStatuses, nil
}

// MarkBreakfastConsumed marks breakfast as consumed for a specific room
func (s *RoomGridService) MarkBreakfastConsumed(propertyID, roomNumber string, staffID uint, paymentMethod string, notes string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Get the guest for this room
		var guest models.Guest
		today := time.Now().Truncate(24 * time.Hour)
		
		err := tx.Where("room_number = ? AND property_id = ? AND is_active = ? AND check_in_date <= ? AND check_out_date > ?",
			roomNumber, propertyID, true, today, today).
			First(&guest).Error
		if err != nil {
			return fmt.Errorf("no active guest found for room %s: %w", roomNumber, err)
		}

		if !guest.BreakfastPackage {
			return fmt.Errorf("guest in room %s does not have breakfast package", roomNumber)
		}

		// Check if already consumed today
		var existingConsumption models.DailyBreakfastConsumption
		err = tx.Where("room_number = ? AND property_id = ? AND consumption_date = ?",
			roomNumber, propertyID, today).First(&existingConsumption).Error
		
		if err == nil && existingConsumption.Status == "consumed" {
			return fmt.Errorf("breakfast already consumed today for room %s", roomNumber)
		}

		now := time.Now()
		consumption := models.DailyBreakfastConsumption{
			PropertyID:      propertyID,
			RoomNumber:      roomNumber,
			GuestID:         guest.ID,
			ConsumptionDate: today,
			ConsumedAt:      &now,
			ConsumedBy:      &staffID,
			Status:          "consumed",
			Notes:           notes,
			PaymentMethod:   paymentMethod,
			Amount:          25.00, // Default breakfast charge
		}

		// Handle OHIP coverage if applicable
		if guest.OHIPNumber != "" && paymentMethod == "ohip" {
			consumption.OHIPCovered = true
			// Submit OHIP claim (simplified)
			// In production, this would be more complex
		}

		if err := tx.Create(&consumption).Error; err != nil {
			return fmt.Errorf("failed to create consumption record: %w", err)
		}

		// Post charge to PMS if payment method is room_charge
		if paymentMethod == "room_charge" {
			// Create charge request using the PMS service method
			err := s.pmsService.ChargeBreakfast(guest.PMSGuestID, guest.ReservationID, roomNumber, consumption.Amount, propertyID)
			if err != nil {
				// Log error but don't fail the consumption tracking
				fmt.Printf("Failed to post charge to PMS: %v\n", err)
			} else {
				consumption.PMSPosted = true
				consumption.PMSTransactionID = "PMS_SUCCESS" // Simplified for now
				tx.Save(&consumption)
			}
		}

		return nil
	})
}

// GetRoomDetails returns detailed information for a specific room
func (s *RoomGridService) GetRoomDetails(propertyID, roomNumber string, date time.Time) (*models.RoomBreakfastStatus, error) {
	rooms, err := s.GetRoomGrid(propertyID, date)
	if err != nil {
		return nil, err
	}

	for _, room := range rooms {
		if room.RoomNumber == roomNumber {
			return &room, nil
		}
	}

	return nil, fmt.Errorf("room %s not found", roomNumber)
}

// SyncGuestsFromPMS synchronizes guest data from PMS
func (s *RoomGridService) SyncGuestsFromPMS(propertyID string) error {
	// Search for all checked-in guests from PMS
	criteria := map[string]string{
		"property_id": propertyID,
		"status":      "checked_in",
	}

	pmsGuests, err := s.pmsService.SearchGuests(criteria)
	if err != nil {
		return fmt.Errorf("failed to fetch guests from PMS: %w", err)
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		for _, pmsGuest := range pmsGuests {
			// Check if guest already exists
			var existingGuest models.Guest
			err := tx.Where("pms_guest_id = ?", pmsGuest.GuestID).First(&existingGuest).Error
			
			if err == gorm.ErrRecordNotFound {
				// Create new guest
				guest := models.Guest{
					PMSGuestID:       pmsGuest.GuestID,
					ReservationID:    pmsGuest.ReservationID,
					RoomNumber:       pmsGuest.RoomNumber,
					FirstName:        pmsGuest.FirstName,
					LastName:         pmsGuest.LastName,
					Email:            pmsGuest.Email,
					Phone:            pmsGuest.Phone,
					CheckInDate:      pmsGuest.CheckInDate,
					CheckOutDate:     pmsGuest.CheckOutDate,
					BreakfastPackage: pmsGuest.BreakfastPackage,
					BreakfastCount:   2, // Default: 2 breakfasts per day for double occupancy
					PropertyID:       pmsGuest.PropertyID,
					IsActive:         true,
				}

				if err := tx.Create(&guest).Error; err != nil {
					return fmt.Errorf("failed to create guest %s: %w", pmsGuest.GuestID, err)
				}
			} else if err == nil {
				// Update existing guest
				existingGuest.RoomNumber = pmsGuest.RoomNumber
				existingGuest.CheckInDate = pmsGuest.CheckInDate
				existingGuest.CheckOutDate = pmsGuest.CheckOutDate
				existingGuest.BreakfastPackage = pmsGuest.BreakfastPackage
				existingGuest.IsActive = pmsGuest.Status == "checked_in"

				if err := tx.Save(&existingGuest).Error; err != nil {
					return fmt.Errorf("failed to update guest %s: %w", pmsGuest.GuestID, err)
				}
			}
		}

		return nil
	})
}

// GetConsumptionHistory returns breakfast consumption history for a date range
func (s *RoomGridService) GetConsumptionHistory(propertyID string, startDate, endDate time.Time) ([]models.DailyBreakfastConsumption, error) {
	var consumptions []models.DailyBreakfastConsumption
	
	err := s.db.Preload("Guest").
		Preload("Staff").
		Preload("Room").
		Where("property_id = ? AND consumption_date BETWEEN ? AND ?", 
			propertyID, startDate, endDate).
		Order("consumption_date DESC, room_number ASC").
		Find(&consumptions).Error

	return consumptions, err
}

// GenerateDailyReport generates a summary report for breakfast consumption
func (s *RoomGridService) GenerateDailyReport(propertyID string, date time.Time) (map[string]interface{}, error) {
	dateOnly := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	
	var totalRooms int64
	var occupiedRooms int64
	var roomsWithBreakfast int64
	var consumedBreakfasts int64
	var ohipCovered int64
	var totalRevenue float64

	// Count total rooms
	s.db.Model(&models.Room{}).Where("property_id = ?", propertyID).Count(&totalRooms)

	// Count occupied rooms with active guests
	s.db.Model(&models.Guest{}).
		Where("property_id = ? AND is_active = ? AND check_in_date <= ? AND check_out_date > ?",
			propertyID, true, dateOnly, dateOnly).
		Count(&occupiedRooms)

	// Count rooms with breakfast packages
	s.db.Model(&models.Guest{}).
		Where("property_id = ? AND is_active = ? AND breakfast_package = ? AND check_in_date <= ? AND check_out_date > ?",
			propertyID, true, true, dateOnly, dateOnly).
		Count(&roomsWithBreakfast)

	// Count consumed breakfasts today
	s.db.Model(&models.DailyBreakfastConsumption{}).
		Where("property_id = ? AND consumption_date = ? AND status = ?",
			propertyID, dateOnly, "consumed").
		Count(&consumedBreakfasts)

	// Count OHIP covered breakfasts
	s.db.Model(&models.DailyBreakfastConsumption{}).
		Where("property_id = ? AND consumption_date = ? AND status = ? AND ohip_covered = ?",
			propertyID, dateOnly, "consumed", true).
		Count(&ohipCovered)

	// Calculate total revenue
	s.db.Model(&models.DailyBreakfastConsumption{}).
		Where("property_id = ? AND consumption_date = ? AND status = ?",
			propertyID, dateOnly, "consumed").
		Select("COALESCE(SUM(amount), 0)").
		Row().Scan(&totalRevenue)

	report := map[string]interface{}{
		"date":                 dateOnly.Format("2006-01-02"),
		"property_id":          propertyID,
		"total_rooms":          totalRooms,
		"occupied_rooms":       occupiedRooms,
		"rooms_with_breakfast": roomsWithBreakfast,
		"consumed_breakfasts":  consumedBreakfasts,
		"ohip_covered":         ohipCovered,
		"total_revenue":        totalRevenue,
		"consumption_rate":     float64(consumedBreakfasts) / float64(roomsWithBreakfast) * 100,
		"occupancy_rate":       float64(occupiedRooms) / float64(totalRooms) * 100,
	}

	return report, nil
}
