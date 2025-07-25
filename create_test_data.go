package main

import (
	"fmt"
	"log"
	"time"

	"hudini-breakfast-module/internal/database"
	"hudini-breakfast-module/internal/models"

	"gorm.io/gorm"
)

const testPasswordHash = "$2a$14$test" // This should be properly hashed in production

func main() {
	// Initialize database
	db, err := database.InitDatabase()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Create test data
	createTestData(db)
}

func createTestData(db *gorm.DB) {
	// Create a property
	property := models.Property{
		PropertyID: "PROP001",
		Name:       "Grand Hotel Downtown",
		Address:    "123 Main Street, Downtown",
		TotalRooms: 50,
		FloorCount: 5,
	}
	db.FirstOrCreate(&property, models.Property{PropertyID: property.PropertyID})

	// Create staff for testing
	staff := []models.Staff{
		{
			Username:     "admin",
			PasswordHash: testPasswordHash,
			FirstName:    "John",
			LastName:     "Admin",
			Email:        "admin@grandhotel.com",
			Role:         "admin",
			PropertyID:   "PROP001",
			IsActive:     true,
		},
		{
			Username:     "manager",
			PasswordHash: testPasswordHash,
			FirstName:    "Jane",
			LastName:     "Manager",
			Email:        "manager@grandhotel.com",
			Role:         "manager",
			PropertyID:   "PROP001",
			IsActive:     true,
		},
		{
			Username:     "frontdesk",
			PasswordHash: testPasswordHash,
			FirstName:    "Mike",
			LastName:     "Johnson",
			Email:        "frontdesk@grandhotel.com",
			Role:         "staff",
			PropertyID:   "PROP001",
			IsActive:     true,
		},
	}

	for _, s := range staff {
		db.FirstOrCreate(&s, models.Staff{Username: s.Username})
	}

	// Create rooms (10 rooms per floor, 5 floors)
	rooms := []models.Room{}
	roomTypes := []string{"standard", "deluxe", "suite"}

	for floor := 1; floor <= 5; floor++ {
		for roomNum := 1; roomNum <= 10; roomNum++ {
			roomNumber := fmt.Sprintf("%d%02d", floor, roomNum)
			roomType := roomTypes[roomNum%3]
			maxOccupancy := 2
			if roomType == "suite" {
				maxOccupancy = 4
			} else if roomType == "deluxe" {
				maxOccupancy = 3
			}

			room := models.Room{
				PropertyID:   "PROP001",
				RoomNumber:   roomNumber,
				Floor:        floor,
				RoomType:     roomType,
				MaxOccupancy: maxOccupancy,
				Status:       "available",
			}
			rooms = append(rooms, room)
		}
	}

	for _, room := range rooms {
		db.FirstOrCreate(&room, models.Room{
			PropertyID: room.PropertyID,
			RoomNumber: room.RoomNumber,
		})
	}

	// Create some test guests with various scenarios
	guests := []models.Guest{
		// Guests with breakfast packages - consumed
		{
			PMSGuestID:       "GUEST001",
			ReservationID:    "RES001",
			RoomNumber:       "101",
			PropertyID:       "PROP001",
			FirstName:        "Alice",
			LastName:         "Johnson",
			Email:            "alice@email.com",
			Phone:            "+1234567890",
			CheckInDate:      time.Now().AddDate(0, 0, -1),
			CheckOutDate:     time.Now().AddDate(0, 0, 2),
			AdultCount:       2,
			ChildCount:       0,
			BreakfastPackage: true,
			BreakfastCount:   2,
			IsActive:         true,
		},
		{
			PMSGuestID:       "GUEST002",
			ReservationID:    "RES002",
			RoomNumber:       "102",
			PropertyID:       "PROP001",
			FirstName:        "Bob",
			LastName:         "Smith",
			Email:            "bob@email.com",
			Phone:            "+1234567891",
			CheckInDate:      time.Now().AddDate(0, 0, -2),
			CheckOutDate:     time.Now().AddDate(0, 0, 1),
			AdultCount:       1,
			ChildCount:       1,
			BreakfastPackage: true,
			BreakfastCount:   2,
			IsActive:         true,
		},
		// Guests with breakfast packages - not consumed
		{
			PMSGuestID:       "GUEST003",
			ReservationID:    "RES003",
			RoomNumber:       "103",
			PropertyID:       "PROP001",
			FirstName:        "Carol",
			LastName:         "Davis",
			Email:            "carol@email.com",
			Phone:            "+1234567892",
			CheckInDate:      time.Now(),
			CheckOutDate:     time.Now().AddDate(0, 0, 3),
			AdultCount:       2,
			ChildCount:       1,
			BreakfastPackage: true,
			BreakfastCount:   3,
			IsActive:         true,
		},
		{
			PMSGuestID:       "GUEST004",
			ReservationID:    "RES004",
			RoomNumber:       "201",
			PropertyID:       "PROP001",
			FirstName:        "David",
			LastName:         "Wilson",
			Email:            "david@email.com",
			Phone:            "+1234567893",
			CheckInDate:      time.Now(),
			CheckOutDate:     time.Now().AddDate(0, 0, 2),
			AdultCount:       2,
			ChildCount:       0,
			BreakfastPackage: true,
			BreakfastCount:   2,
			IsActive:         true,
		},
		// Guests without breakfast packages
		{
			PMSGuestID:       "GUEST005",
			ReservationID:    "RES005",
			RoomNumber:       "202",
			PropertyID:       "PROP001",
			FirstName:        "Eva",
			LastName:         "Brown",
			Email:            "eva@email.com",
			Phone:            "+1234567894",
			CheckInDate:      time.Now().AddDate(0, 0, -1),
			CheckOutDate:     time.Now().AddDate(0, 0, 1),
			AdultCount:       1,
			ChildCount:       0,
			BreakfastPackage: false,
			BreakfastCount:   0,
			IsActive:         true,
		},
		{
			PMSGuestID:       "GUEST006",
			ReservationID:    "RES006",
			RoomNumber:       "301",
			PropertyID:       "PROP001",
			FirstName:        "Frank",
			LastName:         "Miller",
			Email:            "frank@email.com",
			Phone:            "+1234567895",
			CheckInDate:      time.Now(),
			CheckOutDate:     time.Now().AddDate(0, 0, 4),
			AdultCount:       2,
			ChildCount:       2,
			BreakfastPackage: false,
			BreakfastCount:   0,
			IsActive:         true,
		},
		// More guests with breakfast packages
		{
			PMSGuestID:       "GUEST007",
			ReservationID:    "RES007",
			RoomNumber:       "302",
			PropertyID:       "PROP001",
			FirstName:        "Grace",
			LastName:         "Taylor",
			Email:            "grace@email.com",
			Phone:            "+1234567896",
			CheckInDate:      time.Now().AddDate(0, 0, -1),
			CheckOutDate:     time.Now().AddDate(0, 0, 2),
			AdultCount:       2,
			ChildCount:       0,
			BreakfastPackage: true,
			BreakfastCount:   2,
			IsActive:         true,
		},
		{
			PMSGuestID:       "GUEST008",
			ReservationID:    "RES008",
			RoomNumber:       "401",
			PropertyID:       "PROP001",
			FirstName:        "Henry",
			LastName:         "Anderson",
			Email:            "henry@email.com",
			Phone:            "+1234567897",
			CheckInDate:      time.Now(),
			CheckOutDate:     time.Now().AddDate(0, 0, 1),
			AdultCount:       1,
			ChildCount:       0,
			BreakfastPackage: true,
			BreakfastCount:   1,
			IsActive:         true,
		},
		{
			PMSGuestID:       "GUEST009",
			ReservationID:    "RES009",
			RoomNumber:       "501",
			PropertyID:       "PROP001",
			FirstName:        "Iris",
			LastName:         "Thompson",
			Email:            "iris@email.com",
			Phone:            "+1234567898",
			CheckInDate:      time.Now().AddDate(0, 0, -2),
			CheckOutDate:     time.Now().AddDate(0, 0, 3),
			AdultCount:       2,
			ChildCount:       1,
			BreakfastPackage: true,
			BreakfastCount:   3,
			OHIPNumber:       "1234567890",
			IsActive:         true,
		},
	}

	for _, guest := range guests {
		db.FirstOrCreate(&guest, models.Guest{PMSGuestID: guest.PMSGuestID})
	}

	// Set some rooms to maintenance/out of order
	db.Model(&models.Room{}).Where("room_number IN ?", []string{"105", "205"}).Update("status", "maintenance")
	db.Model(&models.Room{}).Where("room_number IN ?", []string{"305"}).Update("status", "out_of_order")

	// Create some breakfast consumption records
	today := time.Now().Truncate(24 * time.Hour)
	consumptions := []models.DailyBreakfastConsumption{
		{
			PropertyID:      "PROP001",
			RoomNumber:      "101",
			ConsumptionDate: today,
			Status:          "consumed",
			PaymentMethod:   "room_charge",
			ConsumedAt:      &time.Time{},
			StaffID:         1, // First staff member
			Notes:           "Consumed at breakfast restaurant",
		},
		{
			PropertyID:      "PROP001",
			RoomNumber:      "102",
			ConsumptionDate: today,
			Status:          "consumed",
			PaymentMethod:   "room_charge",
			ConsumedAt:      &time.Time{},
			StaffID:         2, // Second staff member
			Notes:           "Room service breakfast",
		},
		{
			PropertyID:      "PROP001",
			RoomNumber:      "302",
			ConsumptionDate: today,
			Status:          "consumed",
			PaymentMethod:   "cash",
			ConsumedAt:      &time.Time{},
			StaffID:         3, // Third staff member
			Notes:           "Cash payment at restaurant",
		},
	}

	// Set consumed_at to current time for consumed breakfasts
	now := time.Now()
	for i := range consumptions {
		consumptions[i].ConsumedAt = &now
	}

	for _, consumption := range consumptions {
		db.FirstOrCreate(&consumption, models.DailyBreakfastConsumption{
			PropertyID:      consumption.PropertyID,
			RoomNumber:      consumption.RoomNumber,
			ConsumptionDate: consumption.ConsumptionDate,
		})
	}

	log.Println("✅ Test data created successfully!")
	log.Println("📊 Summary:")
	log.Printf("   - 1 Property: %s", property.Name)
	log.Printf("   - %d Rooms (across %d floors)", len(rooms), property.FloorCount)
	log.Printf("   - %d Staff members", len(staff))
	log.Printf("   - %d Guests", len(guests))
	log.Printf("   - %d Breakfast consumptions recorded", len(consumptions))
	log.Println("\n🏨 Room Status Overview:")
	log.Println("   - Rooms 101, 102, 302: Breakfast consumed today")
	log.Println("   - Rooms 103, 201, 302, 401, 501: Breakfast packages (not consumed)")
	log.Println("   - Rooms 202, 301: Occupied (no breakfast)")
	log.Println("   - Room 105, 205: Maintenance")
	log.Println("   - Room 305: Out of order")
	log.Println("   - Other rooms: Vacant/Available")
}
