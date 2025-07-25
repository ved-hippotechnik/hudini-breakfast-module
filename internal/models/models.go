package models

import (
	"time"

	"gorm.io/gorm"
)

type Property struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	PropertyID   string    `json:"property_id" gorm:"uniqueIndex;not null"`
	Name         string    `json:"name" gorm:"not null"`
	Address      string    `json:"address"`
	TotalRooms   int       `json:"total_rooms"`
	FloorCount   int       `json:"floor_count"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Room struct {
	ID              uint      `json:"id" gorm:"primaryKey"`
	PropertyID      string    `json:"property_id" gorm:"not null"`
	Property        Property  `json:"property" gorm:"foreignKey:PropertyID;references:PropertyID"`
	RoomNumber      string    `json:"room_number" gorm:"not null"`
	Floor           int       `json:"floor"`
	RoomType        string    `json:"room_type"` // standard, deluxe, suite
	MaxOccupancy    int       `json:"max_occupancy"`
	Status          string    `json:"status" gorm:"default:'available'"` // available, occupied, maintenance, out_of_order
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type Guest struct {
	ID              uint      `json:"id" gorm:"primaryKey"`
	PMSGuestID      string    `json:"pms_guest_id" gorm:"uniqueIndex;not null"`
	ReservationID   string    `json:"reservation_id" gorm:"not null"`
	RoomNumber      string    `json:"room_number" gorm:"not null"`
	Room            Room      `json:"room" gorm:"foreignKey:RoomNumber,PropertyID;references:RoomNumber,PropertyID"`
	FirstName       string    `json:"first_name" gorm:"not null"`
	LastName        string    `json:"last_name" gorm:"not null"`
	Email           string    `json:"email"`
	Phone           string    `json:"phone"`
	CheckInDate     time.Time `json:"check_in_date"`
	CheckOutDate    time.Time `json:"check_out_date"`
	AdultCount      int       `json:"adult_count" gorm:"default:1"`
	ChildCount      int       `json:"child_count" gorm:"default:0"`
	BreakfastPackage bool     `json:"breakfast_package" gorm:"default:false"`
	BreakfastCount   int      `json:"breakfast_count" gorm:"default:0"` // Number of breakfasts included
	OHIPNumber      string    `json:"ohip_number"`
	PropertyID      string    `json:"property_id" gorm:"not null"`
	IsActive        bool      `json:"is_active" gorm:"default:true"`
	// VIP and Special Guest Fields
	IsVIP           bool      `json:"is_vip" gorm:"default:false"`
	IsUpset         bool      `json:"is_upset" gorm:"default:false"`
	SpecialNotes    string    `json:"special_notes" gorm:"type:text"`
	HandlingInstr   string    `json:"handling_instructions" gorm:"type:text"`
	PMSSpecialReq   string    `json:"pms_special_requests" gorm:"type:text"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"-" gorm:"index"`
}

type Staff struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Email       string    `json:"email" gorm:"uniqueIndex;not null"`
	Password    string    `json:"-" gorm:"not null"`
	FirstName   string    `json:"first_name" gorm:"not null"`
	LastName    string    `json:"last_name" gorm:"not null"`
	Role        string    `json:"role" gorm:"default:'staff'"` // staff, manager, admin
	PropertyID  string    `json:"property_id" gorm:"not null"`
	IsActive    bool      `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

type DailyBreakfastConsumption struct {
	ID               uint             `json:"id" gorm:"primaryKey"`
	PropertyID       string           `json:"property_id" gorm:"not null"`
	RoomNumber       string           `json:"room_number" gorm:"not null"`
	Room             Room             `json:"room" gorm:"foreignKey:RoomNumber,PropertyID;references:RoomNumber,PropertyID"`
	GuestID          uint             `json:"guest_id" gorm:"not null"`
	Guest            Guest            `json:"guest" gorm:"foreignKey:GuestID"`
	ConsumptionDate  time.Time        `json:"consumption_date" gorm:"not null;index"` // Date only (YYYY-MM-DD)
	ConsumedAt       *time.Time       `json:"consumed_at,omitempty"` // Actual timestamp when consumed
	ConsumedBy       *uint            `json:"consumed_by,omitempty"` // Staff member who marked it
	Staff            *Staff           `json:"staff,omitempty" gorm:"foreignKey:ConsumedBy"`
	Status           string           `json:"status" gorm:"default:'available'"` // available, consumed, no_show
	Notes            string           `json:"notes"`
	PaymentMethod    string           `json:"payment_method"` // room_charge, ohip, comp, cash
	OHIPCovered      bool             `json:"ohip_covered" gorm:"default:false"`
	OHIPTransaction  *OHIPTransaction `json:"ohip_transaction,omitempty" gorm:"foreignKey:ConsumptionID"`
	PMSPosted        bool             `json:"pms_posted" gorm:"default:false"`
	PMSTransactionID string           `json:"pms_transaction_id"`
	Amount           float64          `json:"amount"`
	CreatedAt        time.Time        `json:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at"`
	DeletedAt        gorm.DeletedAt   `json:"-" gorm:"index"`
}

type OHIPTransaction struct {
	ID                string    `json:"id" gorm:"primaryKey"`
	ConsumptionID     uint      `json:"consumption_id" gorm:"not null"`
	OHIPNumber        string    `json:"ohip_number" gorm:"not null"`
	TransactionType   string    `json:"transaction_type"` // claim, refund, adjustment
	Amount            float64   `json:"amount" gorm:"not null"`
	Status            string    `json:"status"`           // pending, approved, denied, processed
	OHIPResponseCode  string    `json:"ohip_response_code"`
	OHIPMessage       string    `json:"ohip_message"`
	SubmittedAt       time.Time `json:"submitted_at"`
	ProcessedAt       *time.Time `json:"processed_at,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// GuestPreference represents guest preferences and dietary requirements
type GuestPreference struct {
	ID              uint           `json:"id" gorm:"primaryKey"`
	GuestID         uint           `json:"guest_id" gorm:"not null;uniqueIndex"`
	Guest           Guest          `json:"guest" gorm:"foreignKey:GuestID"`
	SeatingPref     string         `json:"seating_preference"` // window, booth, patio, quiet
	DietaryRestr    string         `json:"dietary_restrictions" gorm:"type:text"` // JSON array stored as text
	FavoriteDishes  string         `json:"favorite_dishes" gorm:"type:text"` // JSON array stored as text
	Allergies       string         `json:"allergies" gorm:"type:text"` // JSON array stored as text
	SpecialInstr    string         `json:"special_instructions" gorm:"type:text"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"-" gorm:"index"`
}

// Outlet represents a dining outlet that accepts breakfast packages
type Outlet struct {
	ID              uint           `json:"id" gorm:"primaryKey"`
	PropertyID      string         `json:"property_id" gorm:"not null"`
	Property        Property       `json:"property" gorm:"foreignKey:PropertyID;references:PropertyID"`
	Name            string         `json:"name" gorm:"not null"`
	Location        string         `json:"location"`
	AcceptsPackage  bool           `json:"accepts_breakfast_package" gorm:"default:true"`
	OpenTime        string         `json:"open_time"` // Format: "06:30"
	CloseTime       string         `json:"close_time"` // Format: "10:30"
	Capacity        int            `json:"capacity"`
	MenuType        string         `json:"menu_type"` // buffet, a_la_carte, continental
	IsActive        bool           `json:"is_active" gorm:"default:true"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"-" gorm:"index"`
}

// StaffComment represents categorized comments on guests or consumption
type StaffComment struct {
	ID              uint           `json:"id" gorm:"primaryKey"`
	GuestID         *uint          `json:"guest_id,omitempty"`
	Guest           *Guest         `json:"guest,omitempty" gorm:"foreignKey:GuestID"`
	ConsumptionID   *uint          `json:"consumption_id,omitempty"`
	Consumption     *DailyBreakfastConsumption `json:"consumption,omitempty" gorm:"foreignKey:ConsumptionID"`
	StaffID         uint           `json:"staff_id" gorm:"not null"`
	Staff           Staff          `json:"staff" gorm:"foreignKey:StaffID"`
	Category        string         `json:"category" gorm:"not null"` // dietary, preference, complaint, compliment, general
	Comment         string         `json:"comment" gorm:"type:text;not null"`
	IsResolved      bool           `json:"is_resolved" gorm:"default:false"`
	ResolvedBy      *uint          `json:"resolved_by,omitempty"`
	ResolvedAt      *time.Time     `json:"resolved_at,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}

// RoomBreakfastStatus represents the current breakfast status for a room
type RoomBreakfastStatus struct {
	PropertyID       string    `json:"property_id"`
	RoomNumber       string    `json:"room_number"`
	Floor            int       `json:"floor"`
	RoomType         string    `json:"room_type"`
	Status           string    `json:"status"` // available, occupied, maintenance, out_of_order
	HasGuest         bool      `json:"has_guest"`
	GuestName        string    `json:"guest_name"`
	BreakfastPackage bool      `json:"breakfast_package"`
	BreakfastCount   int       `json:"breakfast_count"`
	ConsumedToday    bool      `json:"consumed_today"`
	ConsumedAt       *time.Time `json:"consumed_at,omitempty"`
	ConsumedBy       string    `json:"consumed_by"`
	CheckInDate      *time.Time `json:"check_in_date,omitempty"`
	CheckOutDate     *time.Time `json:"check_out_date,omitempty"`
	// VIP Status Fields
	IsVIP            bool      `json:"is_vip"`
	IsUpset          bool      `json:"is_upset"`
	SpecialRequests  string    `json:"special_requests"`
}

// AuditLog represents system audit logs for compliance and security
type AuditLog struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	UserID     *uint     `json:"user_id" gorm:"index"`
	User       *Staff    `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Action     string    `json:"action" gorm:"not null;index"`
	Resource   string    `json:"resource" gorm:"not null;index"`
	ResourceID string    `json:"resource_id" gorm:"index"`
	OldValues  string    `json:"old_values" gorm:"type:text"`
	NewValues  string    `json:"new_values" gorm:"type:text"`
	IPAddress  string    `json:"ip_address" gorm:"index"`
	UserAgent  string    `json:"user_agent"`
	Status     string    `json:"status" gorm:"default:'success';index"` // success, failed
	Error      string    `json:"error" gorm:"type:text"`
	CreatedAt  time.Time `json:"created_at" gorm:"index"`
}
