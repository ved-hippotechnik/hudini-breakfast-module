package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"hudini-breakfast-module/internal/config"
	"hudini-breakfast-module/internal/database"
	"hudini-breakfast-module/internal/models"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := database.Initialize(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	log.Println("🎯 Creating dummy data for Hudini Breakfast Module...")

	// Create sample staff members
	createSampleStaff(db)

	// Create sample rooms
	createSampleRooms(db)

	// Create sample guests
	createSampleGuests(db)

	// Create sample consumption data
	createSampleConsumption(db)

	log.Println("✅ Dummy data creation completed!")
	log.Println("📋 You can now login with these test accounts:")
	log.Println("   Admin: admin@hotel.com / password123")
	log.Println("   Manager: manager@hotel.com / password123")
	log.Println("   Staff: staff@hotel.com / password123")
	log.Println("🏨 Hotel has 50 rooms with various occupancy and breakfast statuses")
}

func createSampleStaff(db *gorm.DB) {
	log.Println("👥 Creating sample staff members...")

	staffMembers := []struct {
		Email     string
		Password  string
		FirstName string
		LastName  string
		Role      string
	}{
		{"admin@hotel.com", "password123", "John", "Admin", "admin"},
		{"manager@hotel.com", "password123", "Sarah", "Manager", "manager"},
		{"staff@hotel.com", "password123", "Mike", "Staff", "staff"},
		{"reception@hotel.com", "password123", "Emma", "Reception", "staff"},
		{"housekeeping@hotel.com", "password123", "Lisa", "Housekeeping", "staff"},
	}

	for _, member := range staffMembers {
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(member.Password), bcrypt.DefaultCost)
		
		staff := models.Staff{
			Email:      member.Email,
			Password:   string(hashedPassword),
			FirstName:  member.FirstName,
			LastName:   member.LastName,
			Role:       member.Role,
			PropertyID: "HOTEL001",
			IsActive:   true,
		}

		if err := db.Create(&staff).Error; err != nil {
			log.Printf("Failed to create staff %s: %v", member.Email, err)
		} else {
			log.Printf("   ✓ Created staff: %s %s (%s)", member.FirstName, member.LastName, member.Role)
		}
	}
}

func createSampleRooms(db *gorm.DB) {
	log.Println("🏨 Creating sample rooms...")

	roomTypes := []string{"standard", "deluxe", "suite", "presidential"}
	roomStatuses := []string{"available", "occupied", "maintenance", "out_of_order"}

	for floor := 1; floor <= 5; floor++ {
		for roomNum := 1; roomNum <= 10; roomNum++ {
			roomNumber := fmt.Sprintf("%d%02d", floor, roomNum)
			
			room := models.Room{
				PropertyID:   "HOTEL001",
				RoomNumber:   roomNumber,
				Floor:        floor,
				RoomType:     roomTypes[rand.Intn(len(roomTypes))],
				MaxOccupancy: 2 + rand.Intn(3), // 2-4 people
				Status:       roomStatuses[rand.Intn(len(roomStatuses))],
			}

			if err := db.Create(&room).Error; err != nil {
				log.Printf("Failed to create room %s: %v", roomNumber, err)
			}
		}
	}
	log.Println("   ✓ Created 50 rooms across 5 floors")
}

func createSampleGuests(db *gorm.DB) {
	log.Println("👨‍👩‍👧‍👦 Creating sample guests...")

	firstNames := []string{"John", "Sarah", "Mike", "Emma", "David", "Lisa", "Tom", "Anna", "Chris", "Maria", "James", "Kate", "Alex", "Nina", "Paul"}
	lastNames := []string{"Smith", "Johnson", "Williams", "Brown", "Jones", "Garcia", "Miller", "Davis", "Rodriguez", "Martinez", "Hernandez", "Lopez", "Gonzalez", "Wilson", "Anderson"}
	
	// Get occupied rooms
	var occupiedRooms []models.Room
	db.Where("status = ? AND property_id = ?", "occupied", "HOTEL001").Find(&occupiedRooms)

	if len(occupiedRooms) == 0 {
		// If no occupied rooms, mark some as occupied
		var availableRooms []models.Room
		db.Where("status = ? AND property_id = ?", "available", "HOTEL001").Limit(20).Find(&availableRooms)
		
		for i := 0; i < len(availableRooms) && i < 20; i++ {
			availableRooms[i].Status = "occupied"
			db.Save(&availableRooms[i])
			occupiedRooms = append(occupiedRooms, availableRooms[i])
		}
	}

	guestCount := 0
	for _, room := range occupiedRooms {
		if guestCount >= 25 { // Limit to 25 guests
			break
		}

		// Create 1-2 guests per room
		numGuests := 1 + rand.Intn(2)
		for i := 0; i < numGuests; i++ {
			checkInDate := time.Now().AddDate(0, 0, -rand.Intn(7)) // Checked in within last 7 days
			checkOutDate := checkInDate.AddDate(0, 0, 1+rand.Intn(14)) // Staying 1-14 days
			
			hasBreakfast := rand.Float32() < 0.7 // 70% chance of breakfast package
			breakfastCount := 0
			if hasBreakfast {
				breakfastCount = 1 + rand.Intn(3) // 1-3 breakfast count
			}

			guest := models.Guest{
				PMSGuestID:      fmt.Sprintf("PMS%d", 1000+guestCount),
				ReservationID:   fmt.Sprintf("RES%d", 2000+guestCount),
				RoomNumber:      room.RoomNumber,
				FirstName:       firstNames[rand.Intn(len(firstNames))],
				LastName:        lastNames[rand.Intn(len(lastNames))],
				Email:           fmt.Sprintf("guest%d@email.com", guestCount+1),
				Phone:           fmt.Sprintf("+1-555-%04d", rand.Intn(10000)),
				CheckInDate:     checkInDate,
				CheckOutDate:    checkOutDate,
				AdultCount:      1 + rand.Intn(3), // 1-3 adults
				ChildCount:      rand.Intn(3),     // 0-2 children
				BreakfastPackage: hasBreakfast,
				BreakfastCount:  breakfastCount,
				PropertyID:      "HOTEL001",
				IsActive:        true,
			}

			// Some guests might have OHIP
			if rand.Float32() < 0.3 { // 30% chance
				guest.OHIPNumber = fmt.Sprintf("OHIP%d", 5000+guestCount)
			}

			if err := db.Create(&guest).Error; err != nil {
				log.Printf("Failed to create guest: %v", err)
			} else {
				guestCount++
			}
		}
	}
	
	log.Printf("   ✓ Created %d guests in occupied rooms", guestCount)
}

func createSampleConsumption(db *gorm.DB) {
	log.Println("🍳 Creating sample breakfast consumption data...")

	// Get guests with breakfast packages
	var guestsWithBreakfast []models.Guest
	db.Where("breakfast_package = ? AND property_id = ?", true, "HOTEL001").Find(&guestsWithBreakfast)

	if len(guestsWithBreakfast) == 0 {
		log.Println("   No guests with breakfast packages found")
		return
	}

	// Get staff members for consumption tracking
	var staff []models.Staff
	db.Where("property_id = ?", "HOTEL001").Find(&staff)

	if len(staff) == 0 {
		log.Println("   No staff members found for consumption tracking")
		return
	}

	paymentMethods := []string{"room_charge", "ohip", "comp", "cash"}

	consumptionCount := 0

	// Create consumption records for the last 7 days
	for days := 0; days < 7; days++ {
		consumptionDate := time.Now().AddDate(0, 0, -days)
		
		for _, guest := range guestsWithBreakfast {
			// Check if guest was staying on this date
			if consumptionDate.Before(guest.CheckInDate) || consumptionDate.After(guest.CheckOutDate) {
				continue
			}

			// 80% chance of consuming breakfast on any given day
			if rand.Float32() < 0.8 {
				paymentMethod := paymentMethods[rand.Intn(len(paymentMethods))]
				
				// Use OHIP if guest has OHIP number
				if guest.OHIPNumber != "" && rand.Float32() < 0.5 {
					paymentMethod = "ohip"
				}

				consumedAt := consumptionDate.Add(time.Hour * time.Duration(7+rand.Intn(4))) // Between 7-11 AM
				staffMember := staff[rand.Intn(len(staff))]
				
				consumption := models.DailyBreakfastConsumption{
					PropertyID:      "HOTEL001",
					RoomNumber:      guest.RoomNumber,
					GuestID:         guest.ID,
					ConsumptionDate: consumptionDate,
					ConsumedAt:      &consumedAt,
					ConsumedBy:      &staffMember.ID,
					Status:          "consumed",
					Notes:           "Sample consumption data",
					PaymentMethod:   paymentMethod,
					OHIPCovered:     paymentMethod == "ohip",
					PMSPosted:       rand.Float32() < 0.9, // 90% posted to PMS
					PMSTransactionID: fmt.Sprintf("TXN%d", 8000+consumptionCount),
					Amount:          15.00 + rand.Float64()*10.00, // $15-25
				}

				if err := db.Create(&consumption).Error; err != nil {
					log.Printf("Failed to create consumption record: %v", err)
				} else {
					consumptionCount++
				}
			}
		}
	}

	log.Printf("   ✓ Created %d breakfast consumption records over 7 days", consumptionCount)
}
