package services

import (
	"context"
	"fmt"
	"time"

	"hudini-breakfast-module/internal/config"
	"hudini-breakfast-module/internal/middleware"
	"hudini-breakfast-module/internal/models"
)

// PMSIntegrationService provides integration with various PMS systems
type PMSIntegrationService struct {
	middleware     *middleware.PMSMiddleware
	config         *config.Config
	logger         Logger
	defaultProvider middleware.PMSProvider
}

// Logger interface for the service
type Logger interface {
	Info(args ...interface{})
	Error(args ...interface{})
	Debug(args ...interface{})
	Warn(args ...interface{})
}

// NewPMSIntegrationService creates a new PMS integration service
func NewPMSIntegrationService(cfg *config.Config, logger Logger) *PMSIntegrationService {
	// Create the middleware
	pmsMiddleware := middleware.NewPMSMiddleware(cfg, logger)
	
	service := &PMSIntegrationService{
		middleware: pmsMiddleware,
		config:     cfg,
		logger:     logger,
	}
	
	// Initialize providers
	service.initializeProviders()
	
	return service
}

// initializeProviders registers all configured PMS providers
func (s *PMSIntegrationService) initializeProviders() {
	for name, providerConfig := range s.config.PMSProviders.Providers {
		if !providerConfig.Enabled {
			s.logger.Info(fmt.Sprintf("Skipping disabled PMS provider: %s", name))
			continue
		}
		
		switch providerConfig.Type {
		case "oracle_ohip":
			s.registerOracleOHIPProvider(name, providerConfig)
		case "opera":
			s.registerOperaProvider(name, providerConfig)
		case "fidelio":
			s.registerFidelioProvider(name, providerConfig)
		default:
			s.logger.Warn(fmt.Sprintf("Unknown PMS provider type: %s", providerConfig.Type))
		}
	}
	
	// Set default provider
	if s.config.PMSProviders.DefaultProvider != "" {
		provider, err := s.middleware.GetProvider(s.config.PMSProviders.DefaultProvider)
		if err != nil {
			s.logger.Error(fmt.Sprintf("Failed to set default provider: %v", err))
		} else {
			s.defaultProvider = provider
			s.logger.Info(fmt.Sprintf("Default PMS provider set to: %s", s.config.PMSProviders.DefaultProvider))
		}
	}
}

// registerOracleOHIPProvider registers Oracle OHIP provider
func (s *PMSIntegrationService) registerOracleOHIPProvider(name string, providerConfig config.PMSProviderConfig) {
	ohipConfig := config.OHIPConfig{
		BaseURL:      providerConfig.BaseURL,
		ClientID:     providerConfig.ClientID,
		ClientSecret: providerConfig.ClientSecret,
		Username:     providerConfig.Username,
		Password:     providerConfig.Password,
		Environment:  providerConfig.Environment,
		Timeout:      providerConfig.Timeout,
	}
	
	provider := NewOracleOHIPProvider(ohipConfig)
	s.middleware.RegisterProvider(name, provider)
	
	// Authenticate the provider
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	credentials := middleware.PMSCredentials{
		Username:     providerConfig.Username,
		Password:     providerConfig.Password,
		ClientID:     providerConfig.ClientID,
		ClientSecret: providerConfig.ClientSecret,
		BaseURL:      providerConfig.BaseURL,
		PropertyID:   providerConfig.PropertyID,
	}
	
	if err := provider.Authenticate(ctx, credentials); err != nil {
		s.logger.Error(fmt.Sprintf("Failed to authenticate Oracle OHIP provider %s: %v", name, err))
	} else {
		s.logger.Info(fmt.Sprintf("Successfully authenticated Oracle OHIP provider: %s", name))
	}
}

// registerOperaProvider registers Oracle Opera provider (placeholder)
func (s *PMSIntegrationService) registerOperaProvider(name string, providerConfig config.PMSProviderConfig) {
	s.logger.Info(fmt.Sprintf("Opera provider registration not implemented yet: %s", name))
	// TODO: Implement Opera provider when needed
}

// registerFidelioProvider registers Fidelio provider (placeholder)
func (s *PMSIntegrationService) registerFidelioProvider(name string, providerConfig config.PMSProviderConfig) {
	s.logger.Info(fmt.Sprintf("Fidelio provider registration not implemented yet: %s", name))
	// TODO: Implement Fidelio provider when needed
}

// GetGuestProfile retrieves guest profile from PMS
func (s *PMSIntegrationService) GetGuestProfile(ctx context.Context, roomNumber string) (*models.Guest, error) {
	if s.defaultProvider == nil {
		return nil, fmt.Errorf("no default PMS provider configured")
	}
	
	profile, err := s.defaultProvider.GetGuestProfile(ctx, roomNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get guest profile: %w", err)
	}
	
	// Convert to internal model
	guest := &models.Guest{
		PMSGuestID:       profile.GuestID,
		ReservationID:    profile.ReservationID,
		RoomNumber:       profile.RoomNumber,
		FirstName:        profile.FirstName,
		LastName:         profile.LastName,
		Email:            profile.Email,
		Phone:            profile.Phone,
		CheckInDate:      profile.CheckInDate,
		CheckOutDate:     profile.CheckOutDate,
		BreakfastPackage: profile.BreakfastPackage,
		PropertyID:       profile.PropertyID,
	}
	
	return guest, nil
}

// GetGuestByReservation retrieves guest by reservation ID
func (s *PMSIntegrationService) GetGuestByReservation(ctx context.Context, reservationID string) (*models.Guest, error) {
	if s.defaultProvider == nil {
		return nil, fmt.Errorf("no default PMS provider configured")
	}
	
	profile, err := s.defaultProvider.GetGuestByReservation(ctx, reservationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get guest by reservation: %w", err)
	}
	
	// Convert to internal model
	guest := &models.Guest{
		PMSGuestID:       profile.GuestID,
		ReservationID:    profile.ReservationID,
		RoomNumber:       profile.RoomNumber,
		FirstName:        profile.FirstName,
		LastName:         profile.LastName,
		Email:            profile.Email,
		Phone:            profile.Phone,
		CheckInDate:      profile.CheckInDate,
		CheckOutDate:     profile.CheckOutDate,
		BreakfastPackage: profile.BreakfastPackage,
		PropertyID:       profile.PropertyID,
	}
	
	return guest, nil
}

// GetAllGuests retrieves all guests from PMS
func (s *PMSIntegrationService) GetAllGuests(ctx context.Context, propertyID string) ([]models.Guest, error) {
	if s.defaultProvider == nil {
		return nil, fmt.Errorf("no default PMS provider configured")
	}
	
	profiles, err := s.defaultProvider.GetGuestsByProperty(ctx, propertyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get guests: %w", err)
	}
	
	// Convert to internal models
	var guests []models.Guest
	for _, profile := range profiles {
		guest := models.Guest{
			PMSGuestID:       profile.GuestID,
			ReservationID:    profile.ReservationID,
			RoomNumber:       profile.RoomNumber,
			FirstName:        profile.FirstName,
			LastName:         profile.LastName,
			Email:            profile.Email,
			Phone:            profile.Phone,
			CheckInDate:      profile.CheckInDate,
			CheckOutDate:     profile.CheckOutDate,
			BreakfastPackage: profile.BreakfastPackage,
			PropertyID:       profile.PropertyID,
		}
		guests = append(guests, guest)
	}
	
	return guests, nil
}

// GetRoomStatus retrieves room status from PMS
func (s *PMSIntegrationService) GetRoomStatus(ctx context.Context, roomNumber string) (*models.Room, error) {
	if s.defaultProvider == nil {
		return nil, fmt.Errorf("no default PMS provider configured")
	}
	
	roomStatus, err := s.defaultProvider.GetRoomStatus(ctx, roomNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get room status: %w", err)
	}
	
	// Convert to internal model
	room := &models.Room{
		RoomNumber: roomStatus.RoomNumber,
		Status:     roomStatus.Status,
		RoomType:   roomStatus.RoomType,
		PropertyID: roomStatus.PropertyID,
		UpdatedAt:  roomStatus.LastUpdated,
	}
	
	return room, nil
}

// GetAllRooms retrieves all rooms from PMS
func (s *PMSIntegrationService) GetAllRooms(ctx context.Context, propertyID string) ([]models.Room, error) {
	if s.defaultProvider == nil {
		return nil, fmt.Errorf("no default PMS provider configured")
	}
	
	roomStatuses, err := s.defaultProvider.GetRoomsByProperty(ctx, propertyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get rooms: %w", err)
	}
	
	// Convert to internal models
	var rooms []models.Room
	for _, roomStatus := range roomStatuses {
		room := models.Room{
			RoomNumber: roomStatus.RoomNumber,
			Status:     roomStatus.Status,
			RoomType:   roomStatus.RoomType,
			PropertyID: roomStatus.PropertyID,
			UpdatedAt:  roomStatus.LastUpdated,
		}
		rooms = append(rooms, room)
	}
	
	return rooms, nil
}

// PostBreakfastCharge posts a breakfast charge to PMS
func (s *PMSIntegrationService) PostBreakfastCharge(ctx context.Context, guestID, roomNumber string, amount float64) error {
	if s.defaultProvider == nil {
		return fmt.Errorf("no default PMS provider configured")
	}
	
	charge := &middleware.ChargeRequest{
		GuestID:         guestID,
		RoomNumber:      roomNumber,
		ChargeCode:      "BREAKFAST",
		Amount:          amount,
		Description:     "Breakfast Package Charge",
		TransactionDate: time.Now(),
		DepartmentCode:  "F&B",
		PropertyID:      s.config.PMSIntegration.PropertyID,
		Reference:       fmt.Sprintf("BREAKFAST-%s-%s", roomNumber, time.Now().Format("20060102")),
	}
	
	response, err := s.defaultProvider.PostCharge(ctx, charge)
	if err != nil {
		return fmt.Errorf("failed to post breakfast charge: %w", err)
	}
	
	if !response.Success {
		return fmt.Errorf("charge posting failed: %s", response.Message)
	}
	
	s.logger.Info(fmt.Sprintf("Successfully posted breakfast charge for room %s: %s", roomNumber, response.TransactionID))
	return nil
}

// SyncRoomData synchronizes room data with PMS
func (s *PMSIntegrationService) SyncRoomData(ctx context.Context, propertyID string) error {
	if s.defaultProvider == nil {
		return fmt.Errorf("no default PMS provider configured")
	}
	
	// Get all rooms from PMS
	rooms, err := s.GetAllRooms(ctx, propertyID)
	if err != nil {
		return fmt.Errorf("failed to sync room data: %w", err)
	}
	
	// Get all guests from PMS
	guests, err := s.GetAllGuests(ctx, propertyID)
	if err != nil {
		return fmt.Errorf("failed to sync guest data: %w", err)
	}
	
	s.logger.Info(fmt.Sprintf("Synced %d rooms and %d guests from PMS", len(rooms), len(guests)))
	
	// TODO: Update local database with synced data
	// This would involve calling the appropriate database services
	
	return nil
}

// HealthCheck checks the health of all PMS providers
func (s *PMSIntegrationService) HealthCheck(ctx context.Context) map[string]error {
	return s.middleware.HealthCheck(ctx)
}

// GetProviderNames returns all registered provider names
func (s *PMSIntegrationService) GetProviderNames() []string {
	return s.middleware.ListProviders()
}

// SwitchProvider switches to a different PMS provider
func (s *PMSIntegrationService) SwitchProvider(providerName string) error {
	provider, err := s.middleware.GetProvider(providerName)
	if err != nil {
		return fmt.Errorf("failed to switch provider: %w", err)
	}
	
	if !provider.IsAuthenticated() {
		return fmt.Errorf("provider %s is not authenticated", providerName)
	}
	
	s.defaultProvider = provider
	s.logger.Info(fmt.Sprintf("Switched to PMS provider: %s", providerName))
	return nil
}

// GetProviderWithName returns a specific provider by name
func (s *PMSIntegrationService) GetProviderWithName(name string) (middleware.PMSProvider, error) {
	return s.middleware.GetProvider(name)
}

// RefreshProviderTokens refreshes authentication tokens for all providers
func (s *PMSIntegrationService) RefreshProviderTokens(ctx context.Context) error {
	providers := s.middleware.ListProviders()
	
	for _, providerName := range providers {
		provider, err := s.middleware.GetProvider(providerName)
		if err != nil {
			s.logger.Error(fmt.Sprintf("Failed to get provider %s: %v", providerName, err))
			continue
		}
		
		if err := provider.RefreshToken(ctx); err != nil {
			s.logger.Error(fmt.Sprintf("Failed to refresh token for provider %s: %v", providerName, err))
		} else {
			s.logger.Debug(fmt.Sprintf("Successfully refreshed token for provider: %s", providerName))
		}
	}
	
	return nil
}

// StartTokenRefreshScheduler starts a background scheduler to refresh tokens
func (s *PMSIntegrationService) StartTokenRefreshScheduler(ctx context.Context) {
	ticker := time.NewTicker(15 * time.Minute) // Refresh every 15 minutes
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Token refresh scheduler stopped")
			return
		case <-ticker.C:
			if err := s.RefreshProviderTokens(ctx); err != nil {
				s.logger.Error(fmt.Sprintf("Token refresh failed: %v", err))
			}
		}
	}
}
