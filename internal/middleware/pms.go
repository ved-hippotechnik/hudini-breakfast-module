package middleware

import (
	"context"
	"fmt"
	"time"

	"hudini-breakfast-module/internal/config"
)

// PMSProvider defines the interface that all PMS providers must implement
type PMSProvider interface {
	// Authentication
	Authenticate(ctx context.Context, credentials PMSCredentials) error
	RefreshToken(ctx context.Context) error
	IsAuthenticated() bool

	// Guest Management
	GetGuestProfile(ctx context.Context, roomNumber string) (*GuestProfile, error)
	GetGuestByReservation(ctx context.Context, reservationID string) (*GuestProfile, error)
	GetGuestsByProperty(ctx context.Context, propertyID string) ([]GuestProfile, error)
	UpdateGuestProfile(ctx context.Context, guestID string, profile *GuestProfile) error

	// Room Management
	GetRoomStatus(ctx context.Context, roomNumber string) (*RoomStatus, error)
	GetRoomsByProperty(ctx context.Context, propertyID string) ([]RoomStatus, error)
	UpdateRoomStatus(ctx context.Context, roomNumber string, status *RoomStatus) error

	// Charges and Billing
	PostCharge(ctx context.Context, charge *ChargeRequest) (*ChargeResponse, error)
	GetCharges(ctx context.Context, guestID string) ([]Charge, error)
	VoidCharge(ctx context.Context, chargeID string) error

	// Reservations
	GetReservation(ctx context.Context, reservationID string) (*Reservation, error)
	GetReservationsByDate(ctx context.Context, date time.Time) ([]Reservation, error)
	UpdateReservation(ctx context.Context, reservationID string, reservation *Reservation) error

	// Folio Management
	GetFolio(ctx context.Context, guestID string) (*Folio, error)
	UpdateFolio(ctx context.Context, guestID string, folio *Folio) error

	// Health Check
	HealthCheck(ctx context.Context) error
}

// PMSCredentials holds authentication credentials for PMS providers
type PMSCredentials struct {
	Username   string            `json:"username"`
	Password   string            `json:"password"`
	APIKey     string            `json:"api_key"`
	ClientID   string            `json:"client_id"`
	ClientSecret string          `json:"client_secret"`
	BaseURL    string            `json:"base_url"`
	PropertyID string            `json:"property_id"`
	Additional map[string]string `json:"additional"`
}

// GuestProfile represents a guest's profile from PMS
type GuestProfile struct {
	GuestID         string            `json:"guest_id"`
	ReservationID   string            `json:"reservation_id"`
	RoomNumber      string            `json:"room_number"`
	FirstName       string            `json:"first_name"`
	LastName        string            `json:"last_name"`
	Email           string            `json:"email"`
	Phone           string            `json:"phone"`
	CheckInDate     time.Time         `json:"check_in_date"`
	CheckOutDate    time.Time         `json:"check_out_date"`
	BreakfastPackage bool             `json:"breakfast_package"`
	PropertyID      string            `json:"property_id"`
	Status          string            `json:"status"` // checked_in, checked_out, no_show
	VIPStatus       string            `json:"vip_status"`
	Preferences     map[string]string `json:"preferences"`
	LoyaltyProgram  *LoyaltyProgram   `json:"loyalty_program"`
}

// RoomStatus represents room information from PMS
type RoomStatus struct {
	RoomNumber      string    `json:"room_number"`
	Status          string    `json:"status"` // occupied, vacant_clean, vacant_dirty, out_of_order
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

// ChargeRequest represents a charge to be posted to PMS
type ChargeRequest struct {
	GuestID         string            `json:"guest_id"`
	ReservationID   string            `json:"reservation_id"`
	RoomNumber      string            `json:"room_number"`
	ChargeCode      string            `json:"charge_code"`
	Amount          float64           `json:"amount"`
	Description     string            `json:"description"`
	TransactionDate time.Time         `json:"transaction_date"`
	DepartmentCode  string            `json:"department_code"`
	PropertyID      string            `json:"property_id"`
	Reference       string            `json:"reference"`
	TaxAmount       float64           `json:"tax_amount"`
	Metadata        map[string]string `json:"metadata"`
}

// ChargeResponse represents the response from posting a charge
type ChargeResponse struct {
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

// Charge represents a charge in the PMS
type Charge struct {
	ChargeID        string            `json:"charge_id"`
	GuestID         string            `json:"guest_id"`
	ReservationID   string            `json:"reservation_id"`
	RoomNumber      string            `json:"room_number"`
	ChargeCode      string            `json:"charge_code"`
	Amount          float64           `json:"amount"`
	Description     string            `json:"description"`
	TransactionDate time.Time         `json:"transaction_date"`
	DepartmentCode  string            `json:"department_code"`
	Status          string            `json:"status"` // posted, voided, adjusted
	Reference       string            `json:"reference"`
	TaxAmount       float64           `json:"tax_amount"`
	Metadata        map[string]string `json:"metadata"`
}

// Reservation represents a reservation from PMS
type Reservation struct {
	ReservationID   string            `json:"reservation_id"`
	GuestID         string            `json:"guest_id"`
	RoomNumber      string            `json:"room_number"`
	RoomType        string            `json:"room_type"`
	CheckInDate     time.Time         `json:"check_in_date"`
	CheckOutDate    time.Time         `json:"check_out_date"`
	Adults          int               `json:"adults"`
	Children        int               `json:"children"`
	Status          string            `json:"status"` // confirmed, cancelled, no_show, checked_in, checked_out
	RateCode        string            `json:"rate_code"`
	Rate            float64           `json:"rate"`
	PropertyID      string            `json:"property_id"`
	BreakfastPackage bool             `json:"breakfast_package"`
	Preferences     map[string]string `json:"preferences"`
	SpecialRequests []string          `json:"special_requests"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
}

// Folio represents a guest's folio from PMS
type Folio struct {
	FolioID       string    `json:"folio_id"`
	GuestID       string    `json:"guest_id"`
	ReservationID string    `json:"reservation_id"`
	RoomNumber    string    `json:"room_number"`
	Balance       float64   `json:"balance"`
	Charges       []Charge  `json:"charges"`
	Payments      []Payment `json:"payments"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Payment represents a payment in the PMS
type Payment struct {
	PaymentID     string    `json:"payment_id"`
	Amount        float64   `json:"amount"`
	PaymentMethod string    `json:"payment_method"`
	Reference     string    `json:"reference"`
	TransactionDate time.Time `json:"transaction_date"`
	Status        string    `json:"status"`
}

// LoyaltyProgram represents loyalty program information
type LoyaltyProgram struct {
	ProgramName string `json:"program_name"`
	MemberID    string `json:"member_id"`
	Level       string `json:"level"`
	Points      int    `json:"points"`
}

// PMSMiddleware provides a unified interface to different PMS providers
type PMSMiddleware struct {
	providers map[string]PMSProvider
	config    *config.Config
	logger    PMSLogger
}

// PMSLogger interface for logging
type PMSLogger interface {
	Info(args ...interface{})
	Error(args ...interface{})
	Debug(args ...interface{})
	Warn(args ...interface{})
}

// NewPMSMiddleware creates a new PMS middleware instance
func NewPMSMiddleware(config *config.Config, logger PMSLogger) *PMSMiddleware {
	return &PMSMiddleware{
		providers: make(map[string]PMSProvider),
		config:    config,
		logger:    logger,
	}
}

// RegisterProvider registers a new PMS provider
func (m *PMSMiddleware) RegisterProvider(name string, provider PMSProvider) {
	m.providers[name] = provider
	m.logger.Info(fmt.Sprintf("Registered PMS provider: %s", name))
}

// GetProvider returns a PMS provider by name
func (m *PMSMiddleware) GetProvider(name string) (PMSProvider, error) {
	provider, exists := m.providers[name]
	if !exists {
		return nil, fmt.Errorf("PMS provider not found: %s", name)
	}
	return provider, nil
}

// GetDefaultProvider returns the default PMS provider
func (m *PMSMiddleware) GetDefaultProvider() (PMSProvider, error) {
	if len(m.providers) == 0 {
		return nil, fmt.Errorf("no PMS providers registered")
	}
	
	// Return the first registered provider as default
	for _, provider := range m.providers {
		return provider, nil
	}
	
	return nil, fmt.Errorf("no default provider available")
}

// ListProviders returns all registered provider names
func (m *PMSMiddleware) ListProviders() []string {
	var names []string
	for name := range m.providers {
		names = append(names, name)
	}
	return names
}

// HealthCheck checks the health of all registered providers
func (m *PMSMiddleware) HealthCheck(ctx context.Context) map[string]error {
	results := make(map[string]error)
	
	for name, provider := range m.providers {
		results[name] = provider.HealthCheck(ctx)
	}
	
	return results
}
