package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type AnalyticsData struct {
	Success bool                   `json:"success"`
	Data    map[string]interface{} `json:"data"`
}

func main() {
	r := gin.Default()

	// CORS middleware
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization"}
	r.Use(cors.New(config))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().Format(time.RFC3339),
			"service":   "breakfast-module",
		})
	})

	// Analytics endpoints
	r.GET("/api/demo/analytics/advanced", func(c *gin.Context) {
		data := generateAdvancedAnalytics()
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    data,
		})
	})

	r.GET("/api/demo/analytics/realtime", func(c *gin.Context) {
		data := generateRealtimeMetrics()
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    data,
		})
	})

	r.GET("/api/demo/rooms/breakfast-status", func(c *gin.Context) {
		rooms := generateRoomData()
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"rooms":   rooms,
		})
	})

	// Serve static files
	r.StaticFile("/analytics-dashboard.html", "./analytics-dashboard.html")
	r.StaticFile("/analytics-dashboard-clean.html", "./analytics-dashboard-clean.html")
	r.StaticFile("/room-grid-dashboard.html", "./room-grid-dashboard.html")
	r.StaticFile("/enhanced-dashboard.html", "./enhanced-dashboard.html")
	r.StaticFile("/simple-frontend.html", "./simple-frontend.html")
	
	// Default route to clean dashboard
	r.GET("/", func(c *gin.Context) {
		c.Redirect(302, "/analytics-dashboard-clean.html")
	})
	
	// Start server
	fmt.Println("🏨 Hudini Breakfast Module Server starting on :8080")
	fmt.Println("📊 Analytics Dashboard: http://localhost:8080/analytics-dashboard.html")
	fmt.Println("🏨 Enhanced Dashboard: http://localhost:8080/enhanced-dashboard.html")
	fmt.Println("🏠 Room Grid Dashboard: http://localhost:8080/room-grid-dashboard.html")
	fmt.Println("🔍 Health Check: http://localhost:8080/health")
	
	r.Run(":8080")
}

func generateAdvancedAnalytics() map[string]interface{} {
	return map[string]interface{}{
		"period":     "week",
		"property_id": "HOTEL001",
		"timestamp":  time.Now().Format(time.RFC3339),
		"metrics": map[string]interface{}{
			"revenue": map[string]interface{}{
				"current":        25450.00,
				"previous":       23200.00,
				"change":         2250.00,
				"change_percent": 9.7,
				"trend":          "up",
			},
			"occupancy_rate": map[string]interface{}{
				"current":        87.5,
				"previous":       82.3,
				"change":         5.2,
				"change_percent": 6.3,
				"trend":          "up",
			},
			"breakfast_takeup": map[string]interface{}{
				"current":        73.2,
				"previous":       69.8,
				"change":         3.4,
				"change_percent": 4.9,
				"trend":          "up",
			},
			"average_stay_duration": map[string]interface{}{
				"current":        2.8,
				"previous":       2.6,
				"change":         0.2,
				"change_percent": 7.7,
				"trend":          "up",
			},
			"customer_satisfaction": map[string]interface{}{
				"current":        4.6,
				"previous":       4.4,
				"change":         0.2,
				"change_percent": 4.5,
				"trend":          "up",
			},
			"cost_per_breakfast": map[string]interface{}{
				"current":        12.50,
				"previous":       13.20,
				"change":         -0.70,
				"change_percent": -5.3,
				"trend":          "down",
			},
			"total_rooms":        120,
			"occupied_rooms":     105,
			"breakfast_packages": 77,
			"consumed_today":     65,
		},
		"charts": map[string]interface{}{
			"revenue_timeline": []map[string]interface{}{
				{"label": "Week 1", "value": 18500, "date": "2024-01-01"},
				{"label": "Week 2", "value": 21200, "date": "2024-01-08"},
				{"label": "Week 3", "value": 23400, "date": "2024-01-15"},
				{"label": "Week 4", "value": 25450, "date": "2024-01-22"},
			},
			"package_distribution": []map[string]interface{}{
				{"label": "Standard Package", "value": 45, "percentage": 45.0, "color": "#4ade80"},
				{"label": "Premium Package", "value": 25, "percentage": 25.0, "color": "#22d3ee"},
				{"label": "VIP Package", "value": 15, "percentage": 15.0, "color": "#a855f7"},
				{"label": "No Package", "value": 15, "percentage": 15.0, "color": "#6b7280"},
			},
			"hourly_consumption": []map[string]interface{}{
				{"label": "6 AM", "value": 15},
				{"label": "7 AM", "value": 45},
				{"label": "8 AM", "value": 85},
				{"label": "9 AM", "value": 75},
				{"label": "10 AM", "value": 35},
				{"label": "11 AM", "value": 10},
			},
		},
		"insights": []map[string]interface{}{
			{
				"type":        "opportunity",
				"title":       "Revenue Growth Opportunity",
				"description": "Breakfast take-up rate is 15% below industry average",
				"impact":      "Potential $2,400/month additional revenue",
				"action":      "Implement targeted promotional campaign",
				"confidence":  0.87,
				"created_at":  time.Now().Format(time.RFC3339),
			},
			{
				"type":        "warning",
				"title":       "Peak Hour Capacity",
				"description": "8-9 AM shows 23% higher demand than optimal capacity",
				"impact":      "Potential guest dissatisfaction",
				"action":      "Consider extending breakfast hours or increasing staff",
				"confidence":  0.92,
				"created_at":  time.Now().Format(time.RFC3339),
			},
		},
	}
}

func generateRealtimeMetrics() map[string]interface{} {
	return map[string]interface{}{
		"current_occupancy":     87.5 + rand.Float64()*5,
		"breakfast_active":      65 + rand.Intn(10),
		"kitchen_capacity":      85.0 + rand.Float64()*10,
		"wait_time_minutes":     3.2 + rand.Float64()*2,
		"satisfaction_score":    4.6 + rand.Float64()*0.3,
		"cost_efficiency":       94.2 + rand.Float64()*5,
		"staff_utilization":     78.5 + rand.Float64()*10,
		"food_waste_percentage": 8.3 + rand.Float64()*2,
		"last_updated":          time.Now().Format(time.RFC3339),
	}
}

func generateRoomData() []map[string]interface{} {
	rooms := []map[string]interface{}{}
	guestNames := []string{"John Smith", "Jane Doe", "Bob Johnson", "Alice Brown", "Charlie Wilson", "Diana Davis"}
	
	for floor := 1; floor <= 3; floor++ {
		for room := 1; room <= 20; room++ {
			roomNumber := fmt.Sprintf("%d%02d", floor, room)
			hasGuest := rand.Float64() > 0.3
			hasBreakfast := hasGuest && rand.Float64() > 0.4
			isConsumed := hasBreakfast && rand.Float64() > 0.6
			
			roomData := map[string]interface{}{
				"room_number":       roomNumber,
				"floor":             floor,
				"room_type":         "Standard",
				"status":            "available",
				"has_guest":         hasGuest,
				"guest_name":        "",
				"breakfast_package": hasBreakfast,
				"consumed_today":    isConsumed,
				"check_in_date":     nil,
				"check_out_date":    nil,
			}
			
			if hasGuest {
				roomData["status"] = "occupied"
				roomData["guest_name"] = guestNames[rand.Intn(len(guestNames))]
				roomData["check_in_date"] = time.Now().Format(time.RFC3339)
				roomData["check_out_date"] = time.Now().Add(48 * time.Hour).Format(time.RFC3339)
			}
			
			rooms = append(rooms, roomData)
		}
	}
	
	return rooms
}