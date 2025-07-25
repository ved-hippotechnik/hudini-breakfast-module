package services

import (
	"fmt"
	"time"

	"hudini-breakfast-module/internal/logging"
	"hudini-breakfast-module/internal/models"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type GuestService struct {
	db *gorm.DB
}

func NewGuestService(db *gorm.DB) *GuestService {
	return &GuestService{
		db: db,
	}
}

// GetGuests retrieves all active guests for a property
func (s *GuestService) GetGuests(propertyID string) ([]models.Guest, error) {
	logging.WithFields(logrus.Fields{
		"service":     "GuestService",
		"method":      "GetGuests",
		"property_id": propertyID,
	}).Debug("Fetching guests")

	var guests []models.Guest
	err := s.db.Where("property_id = ? AND is_active = true", propertyID).
		Order("room_number ASC").
		Find(&guests).Error

	if err != nil {
		logging.WithFields(logrus.Fields{
			"service":     "GuestService",
			"method":      "GetGuests",
			"property_id": propertyID,
			"error":       err.Error(),
		}).Error("Failed to fetch guests")
		return nil, fmt.Errorf("failed to fetch guests: %w", err)
	}

	logging.WithFields(logrus.Fields{
		"service":     "GuestService",
		"method":      "GetGuests",
		"property_id": propertyID,
		"count":       len(guests),
	}).Debug("Successfully fetched guests")

	return guests, nil
}

// CreateGuest creates a new guest
func (s *GuestService) CreateGuest(guest *models.Guest) error {
	logging.WithFields(logrus.Fields{
		"service":        "GuestService",
		"method":         "CreateGuest",
		"property_id":    guest.PropertyID,
		"room_number":    guest.RoomNumber,
		"pms_guest_id":   guest.PMSGuestID,
		"reservation_id": guest.ReservationID,
	}).Debug("Creating guest")

	// Validate required fields
	if guest.PropertyID == "" {
		return fmt.Errorf("property_id is required")
	}
	if guest.RoomNumber == "" {
		return fmt.Errorf("room_number is required")
	}
	if guest.PMSGuestID == "" {
		return fmt.Errorf("pms_guest_id is required")
	}
	if guest.ReservationID == "" {
		return fmt.Errorf("reservation_id is required")
	}
	if guest.FirstName == "" {
		return fmt.Errorf("first_name is required")
	}
	if guest.LastName == "" {
		return fmt.Errorf("last_name is required")
	}

	// Set default values
	guest.IsActive = true
	guest.CreatedAt = time.Now()
	guest.UpdatedAt = time.Now()

	// Check if guest already exists
	var existingGuest models.Guest
	err := s.db.Where("pms_guest_id = ? AND property_id = ?", guest.PMSGuestID, guest.PropertyID).
		First(&existingGuest).Error

	if err == nil {
		logging.WithFields(logrus.Fields{
			"service":        "GuestService",
			"method":         "CreateGuest",
			"property_id":    guest.PropertyID,
			"pms_guest_id":   guest.PMSGuestID,
			"existing_id":    existingGuest.ID,
		}).Warn("Guest already exists")
		return fmt.Errorf("guest with PMS ID %s already exists", guest.PMSGuestID)
	}

	if err != gorm.ErrRecordNotFound {
		logging.WithFields(logrus.Fields{
			"service":      "GuestService",
			"method":       "CreateGuest",
			"property_id":  guest.PropertyID,
			"pms_guest_id": guest.PMSGuestID,
			"error":        err.Error(),
		}).Error("Failed to check existing guest")
		return fmt.Errorf("failed to check existing guest: %w", err)
	}

	// Create the guest
	err = s.db.Create(guest).Error
	if err != nil {
		logging.WithFields(logrus.Fields{
			"service":      "GuestService",
			"method":       "CreateGuest",
			"property_id":  guest.PropertyID,
			"pms_guest_id": guest.PMSGuestID,
			"error":        err.Error(),
		}).Error("Failed to create guest")
		return fmt.Errorf("failed to create guest: %w", err)
	}

	logging.WithFields(logrus.Fields{
		"service":      "GuestService",
		"method":       "CreateGuest",
		"property_id":  guest.PropertyID,
		"guest_id":     guest.ID,
		"pms_guest_id": guest.PMSGuestID,
	}).Info("Successfully created guest")

	return nil
}

// UpdateGuest updates an existing guest
func (s *GuestService) UpdateGuest(guestID uint, updates *models.Guest) error {
	logging.WithFields(logrus.Fields{
		"service":   "GuestService",
		"method":    "UpdateGuest",
		"guest_id":  guestID,
	}).Debug("Updating guest")

	// Find existing guest
	var existingGuest models.Guest
	err := s.db.First(&existingGuest, guestID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			logging.WithFields(logrus.Fields{
				"service":  "GuestService",
				"method":   "UpdateGuest",
				"guest_id": guestID,
			}).Warn("Guest not found")
			return fmt.Errorf("guest not found")
		}
		logging.WithFields(logrus.Fields{
			"service":  "GuestService",
			"method":   "UpdateGuest",
			"guest_id": guestID,
			"error":    err.Error(),
		}).Error("Failed to find guest")
		return fmt.Errorf("failed to find guest: %w", err)
	}

	// Update fields
	updates.UpdatedAt = time.Now()
	updates.ID = guestID // Ensure ID doesn't change

	err = s.db.Model(&existingGuest).Updates(updates).Error
	if err != nil {
		logging.WithFields(logrus.Fields{
			"service":  "GuestService",
			"method":   "UpdateGuest",
			"guest_id": guestID,
			"error":    err.Error(),
		}).Error("Failed to update guest")
		return fmt.Errorf("failed to update guest: %w", err)
	}

	logging.WithFields(logrus.Fields{
		"service":  "GuestService",
		"method":   "UpdateGuest",
		"guest_id": guestID,
	}).Info("Successfully updated guest")

	return nil
}

// GetGuestByID retrieves a guest by ID
func (s *GuestService) GetGuestByID(guestID uint) (*models.Guest, error) {
	logging.WithFields(logrus.Fields{
		"service":  "GuestService",
		"method":   "GetGuestByID",
		"guest_id": guestID,
	}).Debug("Fetching guest by ID")

	var guest models.Guest
	err := s.db.First(&guest, guestID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			logging.WithFields(logrus.Fields{
				"service":  "GuestService",
				"method":   "GetGuestByID",
				"guest_id": guestID,
			}).Warn("Guest not found")
			return nil, fmt.Errorf("guest not found")
		}
		logging.WithFields(logrus.Fields{
			"service":  "GuestService",
			"method":   "GetGuestByID",
			"guest_id": guestID,
			"error":    err.Error(),
		}).Error("Failed to fetch guest")
		return nil, fmt.Errorf("failed to fetch guest: %w", err)
	}

	logging.WithFields(logrus.Fields{
		"service":  "GuestService",
		"method":   "GetGuestByID",
		"guest_id": guestID,
	}).Debug("Successfully fetched guest")

	return &guest, nil
}

// GetGuestByPMSID retrieves a guest by PMS guest ID
func (s *GuestService) GetGuestByPMSID(propertyID, pmsGuestID string) (*models.Guest, error) {
	logging.WithFields(logrus.Fields{
		"service":        "GuestService",
		"method":         "GetGuestByPMSID",
		"property_id":    propertyID,
		"pms_guest_id":   pmsGuestID,
	}).Debug("Fetching guest by PMS ID")

	var guest models.Guest
	err := s.db.Where("property_id = ? AND pms_guest_id = ?", propertyID, pmsGuestID).
		First(&guest).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			logging.WithFields(logrus.Fields{
				"service":        "GuestService",
				"method":         "GetGuestByPMSID",
				"property_id":    propertyID,
				"pms_guest_id":   pmsGuestID,
			}).Warn("Guest not found")
			return nil, fmt.Errorf("guest not found")
		}
		logging.WithFields(logrus.Fields{
			"service":        "GuestService",
			"method":         "GetGuestByPMSID",
			"property_id":    propertyID,
			"pms_guest_id":   pmsGuestID,
			"error":          err.Error(),
		}).Error("Failed to fetch guest")
		return nil, fmt.Errorf("failed to fetch guest: %w", err)
	}

	logging.WithFields(logrus.Fields{
		"service":        "GuestService",
		"method":         "GetGuestByPMSID",
		"property_id":    propertyID,
		"pms_guest_id":   pmsGuestID,
		"guest_id":       guest.ID,
	}).Debug("Successfully fetched guest")

	return &guest, nil
}

// DeactivateGuest marks a guest as inactive
func (s *GuestService) DeactivateGuest(guestID uint) error {
	logging.WithFields(logrus.Fields{
		"service":  "GuestService",
		"method":   "DeactivateGuest",
		"guest_id": guestID,
	}).Debug("Deactivating guest")

	err := s.db.Model(&models.Guest{}).
		Where("id = ?", guestID).
		Updates(map[string]interface{}{
			"is_active":  false,
			"updated_at": time.Now(),
		}).Error

	if err != nil {
		logging.WithFields(logrus.Fields{
			"service":  "GuestService",
			"method":   "DeactivateGuest",
			"guest_id": guestID,
			"error":    err.Error(),
		}).Error("Failed to deactivate guest")
		return fmt.Errorf("failed to deactivate guest: %w", err)
	}

	logging.WithFields(logrus.Fields{
		"service":  "GuestService",
		"method":   "DeactivateGuest",
		"guest_id": guestID,
	}).Info("Successfully deactivated guest")

	return nil
}

// GetActiveGuestsByRoom retrieves active guests for a specific room
func (s *GuestService) GetActiveGuestsByRoom(propertyID, roomNumber string) ([]models.Guest, error) {
	logging.WithFields(logrus.Fields{
		"service":     "GuestService",
		"method":      "GetActiveGuestsByRoom",
		"property_id": propertyID,
		"room_number": roomNumber,
	}).Debug("Fetching active guests by room")

	var guests []models.Guest
	err := s.db.Where("property_id = ? AND room_number = ? AND is_active = true", propertyID, roomNumber).
		Find(&guests).Error

	if err != nil {
		logging.WithFields(logrus.Fields{
			"service":     "GuestService",
			"method":      "GetActiveGuestsByRoom",
			"property_id": propertyID,
			"room_number": roomNumber,
			"error":       err.Error(),
		}).Error("Failed to fetch guests by room")
		return nil, fmt.Errorf("failed to fetch guests by room: %w", err)
	}

	logging.WithFields(logrus.Fields{
		"service":     "GuestService",
		"method":      "GetActiveGuestsByRoom",
		"property_id": propertyID,
		"room_number": roomNumber,
		"count":       len(guests),
	}).Debug("Successfully fetched guests by room")

	return guests, nil
}