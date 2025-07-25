package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"hudini-breakfast-module/internal/config"
	"hudini-breakfast-module/internal/logging"
	"hudini-breakfast-module/internal/middleware"
)

// OracleOHIPProvider implements the PMSProvider interface for Oracle OHIP
type OracleOHIPProvider struct {
	config      config.OHIPConfig
	httpClient  *http.Client
	accessToken string
	tokenExpiry time.Time
}

// NewOracleOHIPProvider creates a new Oracle OHIP PMS provider
func NewOracleOHIPProvider(config config.OHIPConfig) *OracleOHIPProvider {
	return &OracleOHIPProvider{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Authenticate implements PMSProvider.Authenticate
func (o *OracleOHIPProvider) Authenticate(ctx context.Context, credentials middleware.PMSCredentials) error {
	authRequest := map[string]string{
		"username": credentials.Username,
		"password": credentials.Password,
		"client_id": credentials.ClientID,
		"client_secret": credentials.ClientSecret,
		"grant_type": "client_credentials",
	}

	authData, err := json.Marshal(authRequest)
	if err != nil {
		return fmt.Errorf("failed to marshal auth request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", 
		fmt.Sprintf("%s/oauth2/token", credentials.BaseURL), 
		bytes.NewBuffer(authData))
	if err != nil {
		return fmt.Errorf("failed to create auth request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := o.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("authentication request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("authentication failed with status: %d", resp.StatusCode)
	}

	var authResponse struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&authResponse); err != nil {
		return fmt.Errorf("failed to decode auth response: %w", err)
	}

	o.accessToken = authResponse.AccessToken
	o.tokenExpiry = time.Now().Add(time.Duration(authResponse.ExpiresIn) * time.Second)

	logging.Info("Successfully authenticated with Oracle OHIP")
	return nil
}

// RefreshToken implements PMSProvider.RefreshToken
func (o *OracleOHIPProvider) RefreshToken(ctx context.Context) error {
	if time.Now().Before(o.tokenExpiry.Add(-5 * time.Minute)) {
		return nil // Token is still valid
	}

	// Re-authenticate with stored credentials
	// In a real implementation, you'd store credentials securely
	credentials := middleware.PMSCredentials{
		Username:     o.config.Username,
		Password:     o.config.Password,
		ClientID:     o.config.ClientID,
		ClientSecret: o.config.ClientSecret,
		BaseURL:      o.config.BaseURL,
	}

	return o.Authenticate(ctx, credentials)
}

// IsAuthenticated implements PMSProvider.IsAuthenticated
func (o *OracleOHIPProvider) IsAuthenticated() bool {
	return o.accessToken != "" && time.Now().Before(o.tokenExpiry)
}

// GetGuestProfile implements PMSProvider.GetGuestProfile
func (o *OracleOHIPProvider) GetGuestProfile(ctx context.Context, roomNumber string) (*middleware.GuestProfile, error) {
	if err := o.RefreshToken(ctx); err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/reservations/room/%s", o.config.BaseURL, roomNumber)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", o.accessToken))
	req.Header.Set("Accept", "application/json")

	resp, err := o.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("room not found: %s", roomNumber)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status: %d", resp.StatusCode)
	}

	var ohipResponse struct {
		ReservationID   string    `json:"reservation_id"`
		GuestID         string    `json:"guest_id"`
		RoomNumber      string    `json:"room_number"`
		FirstName       string    `json:"first_name"`
		LastName        string    `json:"last_name"`
		Email           string    `json:"email"`
		Phone           string    `json:"phone"`
		CheckInDate     time.Time `json:"check_in_date"`
		CheckOutDate    time.Time `json:"check_out_date"`
		Status          string    `json:"status"`
		PropertyID      string    `json:"property_id"`
		VIPStatus       string    `json:"vip_status"`
		Preferences     map[string]string `json:"preferences"`
		PackageInclusions []string `json:"package_inclusions"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&ohipResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Check if breakfast package is included
	breakfastPackage := false
	for _, inclusion := range ohipResponse.PackageInclusions {
		if strings.Contains(strings.ToLower(inclusion), "breakfast") {
			breakfastPackage = true
			break
		}
	}

	profile := &middleware.GuestProfile{
		GuestID:         ohipResponse.GuestID,
		ReservationID:   ohipResponse.ReservationID,
		RoomNumber:      ohipResponse.RoomNumber,
		FirstName:       ohipResponse.FirstName,
		LastName:        ohipResponse.LastName,
		Email:           ohipResponse.Email,
		Phone:           ohipResponse.Phone,
		CheckInDate:     ohipResponse.CheckInDate,
		CheckOutDate:    ohipResponse.CheckOutDate,
		BreakfastPackage: breakfastPackage,
		PropertyID:      ohipResponse.PropertyID,
		Status:          ohipResponse.Status,
		VIPStatus:       ohipResponse.VIPStatus,
		Preferences:     ohipResponse.Preferences,
	}

	return profile, nil
}

// GetGuestByReservation implements PMSProvider.GetGuestByReservation
func (o *OracleOHIPProvider) GetGuestByReservation(ctx context.Context, reservationID string) (*middleware.GuestProfile, error) {
	if err := o.RefreshToken(ctx); err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/reservations/%s", o.config.BaseURL, reservationID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", o.accessToken))
	req.Header.Set("Accept", "application/json")

	resp, err := o.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("reservation not found: %s", reservationID)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status: %d", resp.StatusCode)
	}

	var ohipResponse struct {
		ReservationID   string    `json:"reservation_id"`
		GuestID         string    `json:"guest_id"`
		RoomNumber      string    `json:"room_number"`
		FirstName       string    `json:"first_name"`
		LastName        string    `json:"last_name"`
		Email           string    `json:"email"`
		Phone           string    `json:"phone"`
		CheckInDate     time.Time `json:"check_in_date"`
		CheckOutDate    time.Time `json:"check_out_date"`
		Status          string    `json:"status"`
		PropertyID      string    `json:"property_id"`
		VIPStatus       string    `json:"vip_status"`
		Preferences     map[string]string `json:"preferences"`
		PackageInclusions []string `json:"package_inclusions"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&ohipResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Check if breakfast package is included
	breakfastPackage := false
	for _, inclusion := range ohipResponse.PackageInclusions {
		if strings.Contains(strings.ToLower(inclusion), "breakfast") {
			breakfastPackage = true
			break
		}
	}

	profile := &middleware.GuestProfile{
		GuestID:         ohipResponse.GuestID,
		ReservationID:   ohipResponse.ReservationID,
		RoomNumber:      ohipResponse.RoomNumber,
		FirstName:       ohipResponse.FirstName,
		LastName:        ohipResponse.LastName,
		Email:           ohipResponse.Email,
		Phone:           ohipResponse.Phone,
		CheckInDate:     ohipResponse.CheckInDate,
		CheckOutDate:    ohipResponse.CheckOutDate,
		BreakfastPackage: breakfastPackage,
		PropertyID:      ohipResponse.PropertyID,
		Status:          ohipResponse.Status,
		VIPStatus:       ohipResponse.VIPStatus,
		Preferences:     ohipResponse.Preferences,
	}

	return profile, nil
}

// GetGuestsByProperty implements PMSProvider.GetGuestsByProperty
func (o *OracleOHIPProvider) GetGuestsByProperty(ctx context.Context, propertyID string) ([]middleware.GuestProfile, error) {
	if err := o.RefreshToken(ctx); err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/properties/%s/reservations", o.config.BaseURL, propertyID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", o.accessToken))
	req.Header.Set("Accept", "application/json")

	resp, err := o.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status: %d", resp.StatusCode)
	}

	var ohipResponse struct {
		Reservations []struct {
			ReservationID   string    `json:"reservation_id"`
			GuestID         string    `json:"guest_id"`
			RoomNumber      string    `json:"room_number"`
			FirstName       string    `json:"first_name"`
			LastName        string    `json:"last_name"`
			Email           string    `json:"email"`
			Phone           string    `json:"phone"`
			CheckInDate     time.Time `json:"check_in_date"`
			CheckOutDate    time.Time `json:"check_out_date"`
			Status          string    `json:"status"`
			PropertyID      string    `json:"property_id"`
			VIPStatus       string    `json:"vip_status"`
			Preferences     map[string]string `json:"preferences"`
			PackageInclusions []string `json:"package_inclusions"`
		} `json:"reservations"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&ohipResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var profiles []middleware.GuestProfile
	for _, reservation := range ohipResponse.Reservations {
		// Check if breakfast package is included
		breakfastPackage := false
		for _, inclusion := range reservation.PackageInclusions {
			if strings.Contains(strings.ToLower(inclusion), "breakfast") {
				breakfastPackage = true
				break
			}
		}

		profile := middleware.GuestProfile{
			GuestID:         reservation.GuestID,
			ReservationID:   reservation.ReservationID,
			RoomNumber:      reservation.RoomNumber,
			FirstName:       reservation.FirstName,
			LastName:        reservation.LastName,
			Email:           reservation.Email,
			Phone:           reservation.Phone,
			CheckInDate:     reservation.CheckInDate,
			CheckOutDate:    reservation.CheckOutDate,
			BreakfastPackage: breakfastPackage,
			PropertyID:      reservation.PropertyID,
			Status:          reservation.Status,
			VIPStatus:       reservation.VIPStatus,
			Preferences:     reservation.Preferences,
		}

		profiles = append(profiles, profile)
	}

	return profiles, nil
}

// UpdateGuestProfile implements PMSProvider.UpdateGuestProfile
func (o *OracleOHIPProvider) UpdateGuestProfile(ctx context.Context, guestID string, profile *middleware.GuestProfile) error {
	if err := o.RefreshToken(ctx); err != nil {
		return fmt.Errorf("failed to refresh token: %w", err)
	}

	updateData := map[string]interface{}{
		"first_name": profile.FirstName,
		"last_name":  profile.LastName,
		"email":      profile.Email,
		"phone":      profile.Phone,
		"preferences": profile.Preferences,
	}

	requestData, err := json.Marshal(updateData)
	if err != nil {
		return fmt.Errorf("failed to marshal update request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/guests/%s", o.config.BaseURL, guestID)
	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewBuffer(requestData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", o.accessToken))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := o.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("update failed with status: %d", resp.StatusCode)
	}

	return nil
}

// GetRoomStatus implements PMSProvider.GetRoomStatus
func (o *OracleOHIPProvider) GetRoomStatus(ctx context.Context, roomNumber string) (*middleware.RoomStatus, error) {
	if err := o.RefreshToken(ctx); err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/rooms/%s/status", o.config.BaseURL, roomNumber)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", o.accessToken))
	req.Header.Set("Accept", "application/json")

	resp, err := o.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("room not found: %s", roomNumber)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status: %d", resp.StatusCode)
	}

	var ohipResponse struct {
		RoomNumber      string    `json:"room_number"`
		Status          string    `json:"status"`
		RoomType        string    `json:"room_type"`
		GuestID         string    `json:"guest_id"`
		ReservationID   string    `json:"reservation_id"`
		CheckInDate     time.Time `json:"check_in_date"`
		CheckOutDate    time.Time `json:"check_out_date"`
		PropertyID      string    `json:"property_id"`
		HousekeepingStatus string `json:"housekeeping_status"`
		MaintenanceStatus  string `json:"maintenance_status"`
		LastUpdated     time.Time `json:"last_updated"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&ohipResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	roomStatus := &middleware.RoomStatus{
		RoomNumber:      ohipResponse.RoomNumber,
		Status:          ohipResponse.Status,
		RoomType:        ohipResponse.RoomType,
		GuestID:         ohipResponse.GuestID,
		ReservationID:   ohipResponse.ReservationID,
		CheckInDate:     ohipResponse.CheckInDate,
		CheckOutDate:    ohipResponse.CheckOutDate,
		PropertyID:      ohipResponse.PropertyID,
		HousekeepingStatus: ohipResponse.HousekeepingStatus,
		MaintenanceStatus:  ohipResponse.MaintenanceStatus,
		LastUpdated:     ohipResponse.LastUpdated,
	}

	return roomStatus, nil
}

// GetRoomsByProperty implements PMSProvider.GetRoomsByProperty
func (o *OracleOHIPProvider) GetRoomsByProperty(ctx context.Context, propertyID string) ([]middleware.RoomStatus, error) {
	if err := o.RefreshToken(ctx); err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/properties/%s/rooms", o.config.BaseURL, propertyID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", o.accessToken))
	req.Header.Set("Accept", "application/json")

	resp, err := o.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status: %d", resp.StatusCode)
	}

	var ohipResponse struct {
		Rooms []struct {
			RoomNumber      string    `json:"room_number"`
			Status          string    `json:"status"`
			RoomType        string    `json:"room_type"`
			GuestID         string    `json:"guest_id"`
			ReservationID   string    `json:"reservation_id"`
			CheckInDate     time.Time `json:"check_in_date"`
			CheckOutDate    time.Time `json:"check_out_date"`
			PropertyID      string    `json:"property_id"`
			HousekeepingStatus string `json:"housekeeping_status"`
			MaintenanceStatus  string `json:"maintenance_status"`
			LastUpdated     time.Time `json:"last_updated"`
		} `json:"rooms"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&ohipResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var rooms []middleware.RoomStatus
	for _, room := range ohipResponse.Rooms {
		roomStatus := middleware.RoomStatus{
			RoomNumber:      room.RoomNumber,
			Status:          room.Status,
			RoomType:        room.RoomType,
			GuestID:         room.GuestID,
			ReservationID:   room.ReservationID,
			CheckInDate:     room.CheckInDate,
			CheckOutDate:    room.CheckOutDate,
			PropertyID:      room.PropertyID,
			HousekeepingStatus: room.HousekeepingStatus,
			MaintenanceStatus:  room.MaintenanceStatus,
			LastUpdated:     room.LastUpdated,
		}

		rooms = append(rooms, roomStatus)
	}

	return rooms, nil
}

// UpdateRoomStatus implements PMSProvider.UpdateRoomStatus
func (o *OracleOHIPProvider) UpdateRoomStatus(ctx context.Context, roomNumber string, status *middleware.RoomStatus) error {
	if err := o.RefreshToken(ctx); err != nil {
		return fmt.Errorf("failed to refresh token: %w", err)
	}

	updateData := map[string]interface{}{
		"status":               status.Status,
		"housekeeping_status":  status.HousekeepingStatus,
		"maintenance_status":   status.MaintenanceStatus,
	}

	requestData, err := json.Marshal(updateData)
	if err != nil {
		return fmt.Errorf("failed to marshal update request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/rooms/%s/status", o.config.BaseURL, roomNumber)
	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewBuffer(requestData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", o.accessToken))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := o.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("update failed with status: %d", resp.StatusCode)
	}

	return nil
}

// PostCharge implements PMSProvider.PostCharge
func (o *OracleOHIPProvider) PostCharge(ctx context.Context, charge *middleware.ChargeRequest) (*middleware.ChargeResponse, error) {
	if err := o.RefreshToken(ctx); err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	chargeData := map[string]interface{}{
		"guest_id":         charge.GuestID,
		"reservation_id":   charge.ReservationID,
		"room_number":      charge.RoomNumber,
		"charge_code":      charge.ChargeCode,
		"amount":           charge.Amount,
		"description":      charge.Description,
		"transaction_date": charge.TransactionDate.Format(time.RFC3339),
		"department_code":  charge.DepartmentCode,
		"property_id":      charge.PropertyID,
		"reference":        charge.Reference,
		"tax_amount":       charge.TaxAmount,
		"metadata":         charge.Metadata,
	}

	requestData, err := json.Marshal(chargeData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal charge request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/charges", o.config.BaseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(requestData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", o.accessToken))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := o.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	var ohipResponse struct {
		Success         bool              `json:"success"`
		TransactionID   string            `json:"transaction_id"`
		Status          string            `json:"status"`
		Message         string            `json:"message"`
		ErrorCode       string            `json:"error_code"`
		Amount          float64           `json:"amount"`
		Balance         float64           `json:"balance"`
		Timestamp       time.Time         `json:"timestamp"`
		Reference       string            `json:"reference"`
		Metadata        map[string]string `json:"metadata"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&ohipResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return &middleware.ChargeResponse{
			Success:   false,
			Status:    "failed",
			Message:   ohipResponse.Message,
			ErrorCode: ohipResponse.ErrorCode,
		}, nil
	}

	chargeResponse := &middleware.ChargeResponse{
		Success:       ohipResponse.Success,
		TransactionID: ohipResponse.TransactionID,
		Status:        ohipResponse.Status,
		Message:       ohipResponse.Message,
		ErrorCode:     ohipResponse.ErrorCode,
		Amount:        ohipResponse.Amount,
		Balance:       ohipResponse.Balance,
		Timestamp:     ohipResponse.Timestamp,
		Reference:     ohipResponse.Reference,
		Metadata:      ohipResponse.Metadata,
	}

	return chargeResponse, nil
}

// GetCharges implements PMSProvider.GetCharges
func (o *OracleOHIPProvider) GetCharges(ctx context.Context, guestID string) ([]middleware.Charge, error) {
	if err := o.RefreshToken(ctx); err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/guests/%s/charges", o.config.BaseURL, guestID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", o.accessToken))
	req.Header.Set("Accept", "application/json")

	resp, err := o.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status: %d", resp.StatusCode)
	}

	var ohipResponse struct {
		Charges []struct {
			ChargeID        string            `json:"charge_id"`
			GuestID         string            `json:"guest_id"`
			ReservationID   string            `json:"reservation_id"`
			RoomNumber      string            `json:"room_number"`
			ChargeCode      string            `json:"charge_code"`
			Amount          float64           `json:"amount"`
			Description     string            `json:"description"`
			TransactionDate time.Time         `json:"transaction_date"`
			DepartmentCode  string            `json:"department_code"`
			Status          string            `json:"status"`
			Reference       string            `json:"reference"`
			TaxAmount       float64           `json:"tax_amount"`
			Metadata        map[string]string `json:"metadata"`
		} `json:"charges"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&ohipResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var charges []middleware.Charge
	for _, charge := range ohipResponse.Charges {
		chargeItem := middleware.Charge{
			ChargeID:        charge.ChargeID,
			GuestID:         charge.GuestID,
			ReservationID:   charge.ReservationID,
			RoomNumber:      charge.RoomNumber,
			ChargeCode:      charge.ChargeCode,
			Amount:          charge.Amount,
			Description:     charge.Description,
			TransactionDate: charge.TransactionDate,
			DepartmentCode:  charge.DepartmentCode,
			Status:          charge.Status,
			Reference:       charge.Reference,
			TaxAmount:       charge.TaxAmount,
			Metadata:        charge.Metadata,
		}

		charges = append(charges, chargeItem)
	}

	return charges, nil
}

// VoidCharge implements PMSProvider.VoidCharge
func (o *OracleOHIPProvider) VoidCharge(ctx context.Context, chargeID string) error {
	if err := o.RefreshToken(ctx); err != nil {
		return fmt.Errorf("failed to refresh token: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/charges/%s/void", o.config.BaseURL, chargeID)
	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", o.accessToken))
	req.Header.Set("Accept", "application/json")

	resp, err := o.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("void charge failed with status: %d", resp.StatusCode)
	}

	return nil
}

// GetReservation implements PMSProvider.GetReservation
func (o *OracleOHIPProvider) GetReservation(ctx context.Context, reservationID string) (*middleware.Reservation, error) {
	if err := o.RefreshToken(ctx); err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/reservations/%s", o.config.BaseURL, reservationID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", o.accessToken))
	req.Header.Set("Accept", "application/json")

	resp, err := o.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("reservation not found: %s", reservationID)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status: %d", resp.StatusCode)
	}

	var ohipResponse struct {
		ReservationID   string            `json:"reservation_id"`
		GuestID         string            `json:"guest_id"`
		RoomNumber      string            `json:"room_number"`
		RoomType        string            `json:"room_type"`
		CheckInDate     time.Time         `json:"check_in_date"`
		CheckOutDate    time.Time         `json:"check_out_date"`
		Adults          int               `json:"adults"`
		Children        int               `json:"children"`
		Status          string            `json:"status"`
		RateCode        string            `json:"rate_code"`
		Rate            float64           `json:"rate"`
		PropertyID      string            `json:"property_id"`
		Preferences     map[string]string `json:"preferences"`
		SpecialRequests []string          `json:"special_requests"`
		CreatedAt       time.Time         `json:"created_at"`
		UpdatedAt       time.Time         `json:"updated_at"`
		PackageInclusions []string        `json:"package_inclusions"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&ohipResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Check if breakfast package is included
	breakfastPackage := false
	for _, inclusion := range ohipResponse.PackageInclusions {
		if strings.Contains(strings.ToLower(inclusion), "breakfast") {
			breakfastPackage = true
			break
		}
	}

	reservation := &middleware.Reservation{
		ReservationID:   ohipResponse.ReservationID,
		GuestID:         ohipResponse.GuestID,
		RoomNumber:      ohipResponse.RoomNumber,
		RoomType:        ohipResponse.RoomType,
		CheckInDate:     ohipResponse.CheckInDate,
		CheckOutDate:    ohipResponse.CheckOutDate,
		Adults:          ohipResponse.Adults,
		Children:        ohipResponse.Children,
		Status:          ohipResponse.Status,
		RateCode:        ohipResponse.RateCode,
		Rate:            ohipResponse.Rate,
		PropertyID:      ohipResponse.PropertyID,
		BreakfastPackage: breakfastPackage,
		Preferences:     ohipResponse.Preferences,
		SpecialRequests: ohipResponse.SpecialRequests,
		CreatedAt:       ohipResponse.CreatedAt,
		UpdatedAt:       ohipResponse.UpdatedAt,
	}

	return reservation, nil
}

// GetReservationsByDate implements PMSProvider.GetReservationsByDate
func (o *OracleOHIPProvider) GetReservationsByDate(ctx context.Context, date time.Time) ([]middleware.Reservation, error) {
	if err := o.RefreshToken(ctx); err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/reservations?date=%s", o.config.BaseURL, date.Format("2006-01-02"))
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", o.accessToken))
	req.Header.Set("Accept", "application/json")

	resp, err := o.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status: %d", resp.StatusCode)
	}

	var ohipResponse struct {
		Reservations []struct {
			ReservationID   string            `json:"reservation_id"`
			GuestID         string            `json:"guest_id"`
			RoomNumber      string            `json:"room_number"`
			RoomType        string            `json:"room_type"`
			CheckInDate     time.Time         `json:"check_in_date"`
			CheckOutDate    time.Time         `json:"check_out_date"`
			Adults          int               `json:"adults"`
			Children        int               `json:"children"`
			Status          string            `json:"status"`
			RateCode        string            `json:"rate_code"`
			Rate            float64           `json:"rate"`
			PropertyID      string            `json:"property_id"`
			Preferences     map[string]string `json:"preferences"`
			SpecialRequests []string          `json:"special_requests"`
			CreatedAt       time.Time         `json:"created_at"`
			UpdatedAt       time.Time         `json:"updated_at"`
			PackageInclusions []string        `json:"package_inclusions"`
		} `json:"reservations"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&ohipResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var reservations []middleware.Reservation
	for _, reservation := range ohipResponse.Reservations {
		// Check if breakfast package is included
		breakfastPackage := false
		for _, inclusion := range reservation.PackageInclusions {
			if strings.Contains(strings.ToLower(inclusion), "breakfast") {
				breakfastPackage = true
				break
			}
		}

		reservationItem := middleware.Reservation{
			ReservationID:   reservation.ReservationID,
			GuestID:         reservation.GuestID,
			RoomNumber:      reservation.RoomNumber,
			RoomType:        reservation.RoomType,
			CheckInDate:     reservation.CheckInDate,
			CheckOutDate:    reservation.CheckOutDate,
			Adults:          reservation.Adults,
			Children:        reservation.Children,
			Status:          reservation.Status,
			RateCode:        reservation.RateCode,
			Rate:            reservation.Rate,
			PropertyID:      reservation.PropertyID,
			BreakfastPackage: breakfastPackage,
			Preferences:     reservation.Preferences,
			SpecialRequests: reservation.SpecialRequests,
			CreatedAt:       reservation.CreatedAt,
			UpdatedAt:       reservation.UpdatedAt,
		}

		reservations = append(reservations, reservationItem)
	}

	return reservations, nil
}

// UpdateReservation implements PMSProvider.UpdateReservation
func (o *OracleOHIPProvider) UpdateReservation(ctx context.Context, reservationID string, reservation *middleware.Reservation) error {
	if err := o.RefreshToken(ctx); err != nil {
		return fmt.Errorf("failed to refresh token: %w", err)
	}

	updateData := map[string]interface{}{
		"room_number":      reservation.RoomNumber,
		"room_type":        reservation.RoomType,
		"check_in_date":    reservation.CheckInDate.Format(time.RFC3339),
		"check_out_date":   reservation.CheckOutDate.Format(time.RFC3339),
		"adults":           reservation.Adults,
		"children":         reservation.Children,
		"status":           reservation.Status,
		"rate_code":        reservation.RateCode,
		"rate":             reservation.Rate,
		"preferences":      reservation.Preferences,
		"special_requests": reservation.SpecialRequests,
	}

	requestData, err := json.Marshal(updateData)
	if err != nil {
		return fmt.Errorf("failed to marshal update request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/reservations/%s", o.config.BaseURL, reservationID)
	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewBuffer(requestData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", o.accessToken))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := o.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("update failed with status: %d", resp.StatusCode)
	}

	return nil
}

// GetFolio implements PMSProvider.GetFolio
func (o *OracleOHIPProvider) GetFolio(ctx context.Context, guestID string) (*middleware.Folio, error) {
	if err := o.RefreshToken(ctx); err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/guests/%s/folio", o.config.BaseURL, guestID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", o.accessToken))
	req.Header.Set("Accept", "application/json")

	resp, err := o.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("folio not found for guest: %s", guestID)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status: %d", resp.StatusCode)
	}

	var ohipResponse struct {
		FolioID       string    `json:"folio_id"`
		GuestID       string    `json:"guest_id"`
		ReservationID string    `json:"reservation_id"`
		RoomNumber    string    `json:"room_number"`
		Balance       float64   `json:"balance"`
		Status        string    `json:"status"`
		CreatedAt     time.Time `json:"created_at"`
		UpdatedAt     time.Time `json:"updated_at"`
		Charges       []struct {
			ChargeID        string            `json:"charge_id"`
			GuestID         string            `json:"guest_id"`
			ReservationID   string            `json:"reservation_id"`
			RoomNumber      string            `json:"room_number"`
			ChargeCode      string            `json:"charge_code"`
			Amount          float64           `json:"amount"`
			Description     string            `json:"description"`
			TransactionDate time.Time         `json:"transaction_date"`
			DepartmentCode  string            `json:"department_code"`
			Status          string            `json:"status"`
			Reference       string            `json:"reference"`
			TaxAmount       float64           `json:"tax_amount"`
			Metadata        map[string]string `json:"metadata"`
		} `json:"charges"`
		Payments []struct {
			PaymentID     string    `json:"payment_id"`
			Amount        float64   `json:"amount"`
			PaymentMethod string    `json:"payment_method"`
			Reference     string    `json:"reference"`
			TransactionDate time.Time `json:"transaction_date"`
			Status        string    `json:"status"`
		} `json:"payments"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&ohipResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert charges
	var charges []middleware.Charge
	for _, charge := range ohipResponse.Charges {
		chargeItem := middleware.Charge{
			ChargeID:        charge.ChargeID,
			GuestID:         charge.GuestID,
			ReservationID:   charge.ReservationID,
			RoomNumber:      charge.RoomNumber,
			ChargeCode:      charge.ChargeCode,
			Amount:          charge.Amount,
			Description:     charge.Description,
			TransactionDate: charge.TransactionDate,
			DepartmentCode:  charge.DepartmentCode,
			Status:          charge.Status,
			Reference:       charge.Reference,
			TaxAmount:       charge.TaxAmount,
			Metadata:        charge.Metadata,
		}
		charges = append(charges, chargeItem)
	}

	// Convert payments
	var payments []middleware.Payment
	for _, payment := range ohipResponse.Payments {
		paymentItem := middleware.Payment{
			PaymentID:     payment.PaymentID,
			Amount:        payment.Amount,
			PaymentMethod: payment.PaymentMethod,
			Reference:     payment.Reference,
			TransactionDate: payment.TransactionDate,
			Status:        payment.Status,
		}
		payments = append(payments, paymentItem)
	}

	folio := &middleware.Folio{
		FolioID:       ohipResponse.FolioID,
		GuestID:       ohipResponse.GuestID,
		ReservationID: ohipResponse.ReservationID,
		RoomNumber:    ohipResponse.RoomNumber,
		Balance:       ohipResponse.Balance,
		Status:        ohipResponse.Status,
		CreatedAt:     ohipResponse.CreatedAt,
		UpdatedAt:     ohipResponse.UpdatedAt,
		Charges:       charges,
		Payments:      payments,
	}

	return folio, nil
}

// UpdateFolio implements PMSProvider.UpdateFolio
func (o *OracleOHIPProvider) UpdateFolio(ctx context.Context, guestID string, folio *middleware.Folio) error {
	if err := o.RefreshToken(ctx); err != nil {
		return fmt.Errorf("failed to refresh token: %w", err)
	}

	updateData := map[string]interface{}{
		"status": folio.Status,
	}

	requestData, err := json.Marshal(updateData)
	if err != nil {
		return fmt.Errorf("failed to marshal update request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/guests/%s/folio", o.config.BaseURL, guestID)
	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewBuffer(requestData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", o.accessToken))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := o.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("update failed with status: %d", resp.StatusCode)
	}

	return nil
}

// HealthCheck implements PMSProvider.HealthCheck
func (o *OracleOHIPProvider) HealthCheck(ctx context.Context) error {
	if err := o.RefreshToken(ctx); err != nil {
		return fmt.Errorf("failed to refresh token: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/health", o.config.BaseURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", o.accessToken))
	req.Header.Set("Accept", "application/json")

	resp, err := o.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("health check request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed with status: %d", resp.StatusCode)
	}

	return nil
}
