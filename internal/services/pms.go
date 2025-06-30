package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"hudini-breakfast-module/internal/config"
	"hudini-breakfast-module/internal/models"
)

type PMSService struct {
	config     *config.Config
	httpClient *http.Client
}

type PMSGuestProfile struct {
	GuestID         string    `json:"guest_id"`
	ReservationID   string    `json:"reservation_id"`
	RoomNumber      string    `json:"room_number"`
	FirstName       string    `json:"first_name"`
	LastName        string    `json:"last_name"`
	Email           string    `json:"email"`
	Phone           string    `json:"phone"`
	CheckInDate     time.Time `json:"check_in_date"`
	CheckOutDate    time.Time `json:"check_out_date"`
	BreakfastPackage bool     `json:"breakfast_package"`
	PropertyID      string    `json:"property_id"`
	Status          string    `json:"status"` // checked_in, checked_out, no_show
}

type PMSChargeRequest struct {
	GuestID         string  `json:"guest_id"`
	ReservationID   string  `json:"reservation_id"`
	RoomNumber      string  `json:"room_number"`
	ChargeCode      string  `json:"charge_code"`
	Amount          float64 `json:"amount"`
	Description     string  `json:"description"`
	TransactionDate string  `json:"transaction_date"`
	DepartmentCode  string  `json:"department_code"`
	PropertyID      string  `json:"property_id"`
}

type PMSChargeResponse struct {
	Success         bool   `json:"success"`
	TransactionID   string `json:"transaction_id"`
	Status          string `json:"status"`
	Message         string `json:"message"`
	PMSConfirmation string `json:"pms_confirmation"`
}

type PMSGuestSearchResponse struct {
	Success bool              `json:"success"`
	Guests  []PMSGuestProfile `json:"guests"`
	Message string            `json:"message"`
}

func NewPMSService(config *config.Config) *PMSService {
	return &PMSService{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SearchGuests retrieves guest information from PMS by various criteria
func (s *PMSService) SearchGuests(criteria map[string]string) ([]PMSGuestProfile, error) {
	endpoint := fmt.Sprintf("%s/guests/search", s.config.PMSIntegration.BaseURL)
	
	reqBody, err := json.Marshal(criteria)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal search criteria: %w", err)
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.config.PMSIntegration.APIKey))
	req.Header.Set("X-Property-ID", s.config.PMSIntegration.PropertyID)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("PMS API request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("PMS API error: %d - %s", resp.StatusCode, string(body))
	}

	var response PMSGuestSearchResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("PMS search failed: %s", response.Message)
	}

	return response.Guests, nil
}

// GetGuestByID retrieves a specific guest from PMS
func (s *PMSService) GetGuestByID(guestID string) (*PMSGuestProfile, error) {
	endpoint := fmt.Sprintf("%s/guests/%s", s.config.PMSIntegration.BaseURL, guestID)

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.config.PMSIntegration.APIKey))
	req.Header.Set("X-Property-ID", s.config.PMSIntegration.PropertyID)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("PMS API request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("PMS API error: %d - %s", resp.StatusCode, string(body))
	}

	var guest PMSGuestProfile
	if err := json.Unmarshal(body, &guest); err != nil {
		return nil, fmt.Errorf("failed to unmarshal guest: %w", err)
	}

	return &guest, nil
}

// PostCharge posts a breakfast consumption charge to guest's room
func (s *PMSService) PostCharge(charge PMSChargeRequest) (*PMSChargeResponse, error) {
	endpoint := fmt.Sprintf("%s/charges", s.config.PMSIntegration.BaseURL)

	reqBody, err := json.Marshal(charge)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal charge request: %w", err)
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.config.PMSIntegration.APIKey))
	req.Header.Set("X-Property-ID", s.config.PMSIntegration.PropertyID)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("PMS charge API request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("PMS charge API error: %d - %s", resp.StatusCode, string(body))
	}

	var response PMSChargeResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal charge response: %w", err)
	}

	return &response, nil
}

// SyncGuest synchronizes guest data from PMS to local database
func (s *PMSService) SyncGuest(pmsGuest PMSGuestProfile) (*models.Guest, error) {
	guest := &models.Guest{
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
		PropertyID:       pmsGuest.PropertyID,
		IsActive:         pmsGuest.Status == "checked_in",
	}

	return guest, nil
}

// ValidateGuestEligibility checks if guest is eligible for breakfast consumption
func (s *PMSService) ValidateGuestEligibility(guestID string) (bool, string, error) {
	guest, err := s.GetGuestByID(guestID)
	if err != nil {
		return false, "Failed to retrieve guest information", err
	}

	// Check if guest is checked in
	if guest.Status != "checked_in" {
		return false, "Guest is not currently checked in", nil
	}

	// Check if guest has breakfast package
	if !guest.BreakfastPackage {
		return false, "Guest does not have breakfast package", nil
	}

	// Check if it's within breakfast hours (business logic)
	now := time.Now()
	if now.Hour() < 6 || now.Hour() > 11 {
		return false, "Breakfast service is not available at this time", nil
	}

	return true, "Guest is eligible for breakfast", nil
}

// ChargeBreakfast is a convenience method to charge breakfast to a guest's room
func (s *PMSService) ChargeBreakfast(guestID, reservationID, roomNumber string, amount float64, propertyID string) error {
	charge := PMSChargeRequest{
		GuestID:         guestID,
		ReservationID:   reservationID,
		RoomNumber:      roomNumber,
		ChargeCode:      "BRKFST",
		Amount:          amount,
		Description:     "Breakfast Service",
		TransactionDate: time.Now().Format("2006-01-02"),
		DepartmentCode:  "F&B",
		PropertyID:      propertyID,
	}

	_, err := s.PostCharge(charge)
	return err
}
