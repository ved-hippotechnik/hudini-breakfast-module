package pms

import (
	"context"
	"time"
)

// PMSProvider defines the interface for PMS integrations
type PMSProvider interface {
	// Authentication
	Authenticate(ctx context.Context, credentials PMSCredentials) error
	RefreshToken(ctx context.Context) error
	
	// Guest operations
	GetGuestProfile(ctx context.Context, roomNumber string) (*GuestProfile, error)
	GetGuestByReservation(ctx context.Context, reservationID string) (*GuestProfile, error)
	GetGuestsByProperty(ctx context.Context, propertyID string) ([]GuestProfile, error)
	UpdateGuestProfile(ctx context.Context, guestID string, profile *GuestProfile) error
	
	// Room operations
	GetRoomStatus(ctx context.Context, roomNumber string) (*RoomStatus, error)
	GetRoomsByProperty(ctx context.Context, propertyID string) ([]RoomStatus, error)
	UpdateRoomStatus(ctx context.Context, roomNumber string, status *RoomStatus) error
	
	// Charge operations
	PostCharge(ctx context.Context, charge *ChargeRequest) (*ChargeResponse, error)
	ReverseCharge(ctx context.Context, transactionID string) (*ChargeResponse, error)
	GetCharges(ctx context.Context, roomNumber string, startDate, endDate time.Time) ([]ChargeDetail, error)
	
	// Transaction operations
	GetTransaction(ctx context.Context, transactionID string) (*Transaction, error)
	GetTransactionsByRoom(ctx context.Context, roomNumber string, startDate, endDate time.Time) ([]Transaction, error)
	
	// Health check
	HealthCheck(ctx context.Context) error
}

// PMSCredentials holds authentication credentials for PMS
type PMSCredentials struct {
	Username     string
	Password     string
	ClientID     string
	ClientSecret string
	BaseURL      string
	PropertyID   string
}

// GuestProfile represents guest information from PMS
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
	Status          string    `json:"status"` // occupied, vacant, out_of_order
	RoomType        string    `json:"room_type"`
	Floor           int       `json:"floor"`
	GuestID         string    `json:"guest_id"`
	PropertyID      string    `json:"property_id"`
	HousekeepingStatus string `json:"housekeeping_status"`
	MaintenanceStatus  string `json:"maintenance_status"`
	LastUpdated     time.Time `json:"last_updated"`
}

// ChargeRequest represents a charge to be posted to PMS
type ChargeRequest struct {
	GuestID         string    `json:"guest_id"`
	RoomNumber      string    `json:"room_number"`
	ChargeCode      string    `json:"charge_code"`
	Amount          float64   `json:"amount"`
	Description     string    `json:"description"`
	TransactionDate time.Time `json:"transaction_date"`
	DepartmentCode  string    `json:"department_code"`
	PropertyID      string    `json:"property_id"`
	Reference       string    `json:"reference"`
	Tax             float64   `json:"tax"`
	ServiceCharge   float64   `json:"service_charge"`
}

// ChargeResponse represents the response from posting a charge
type ChargeResponse struct {
	TransactionID string    `json:"transaction_id"`
	Status        string    `json:"status"`
	Message       string    `json:"message"`
	PostedAt      time.Time `json:"posted_at"`
}

// ChargeDetail represents detailed charge information
type ChargeDetail struct {
	TransactionID   string    `json:"transaction_id"`
	ChargeCode      string    `json:"charge_code"`
	Amount          float64   `json:"amount"`
	Description     string    `json:"description"`
	PostedDate      time.Time `json:"posted_date"`
	DepartmentCode  string    `json:"department_code"`
	Reference       string    `json:"reference"`
	Tax             float64   `json:"tax"`
	ServiceCharge   float64   `json:"service_charge"`
}

// Transaction represents a PMS transaction
type Transaction struct {
	TransactionID   string    `json:"transaction_id"`
	RoomNumber      string    `json:"room_number"`
	GuestID         string    `json:"guest_id"`
	Type            string    `json:"type"` // charge, payment, adjustment
	Amount          float64   `json:"amount"`
	Description     string    `json:"description"`
	TransactionDate time.Time `json:"transaction_date"`
	PostedDate      time.Time `json:"posted_date"`
	Status          string    `json:"status"`
	Reference       string    `json:"reference"`
}

// LoyaltyProgram represents guest loyalty information
type LoyaltyProgram struct {
	ProgramName   string `json:"program_name"`
	MemberNumber  string `json:"member_number"`
	TierLevel     string `json:"tier_level"`
	Points        int    `json:"points"`
}

// Logger interface for PMS middleware
type Logger interface {
	Info(args ...interface{})
	Error(args ...interface{})
	Debug(args ...interface{})
	Warn(args ...interface{})
}