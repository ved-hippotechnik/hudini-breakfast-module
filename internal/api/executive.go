package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"hudini-breakfast-module/internal/cache"
	"hudini-breakfast-module/internal/logging"
	"hudini-breakfast-module/internal/services"

	"github.com/gin-gonic/gin"
)

type ExecutiveHandler struct {
	breakfastService *services.BreakfastService
	guestService     *services.GuestService
}

func NewExecutiveHandler(breakfastService *services.BreakfastService, guestService *services.GuestService) *ExecutiveHandler {
	return &ExecutiveHandler{
		breakfastService: breakfastService,
		guestService:     guestService,
	}
}

// GetExecutiveKPIs returns key performance indicators for executives
func (h *ExecutiveHandler) GetExecutiveKPIs(c *gin.Context) {
	propertyID := c.Query("property_id")
	if propertyID == "" {
		propertyID = "HOTEL001"
	}

	ctx := context.Background()

	// Get VIP metrics
	vipMetrics, err := h.breakfastService.GetVIPMetrics(ctx, propertyID)
	if err != nil {
		logging.WithError(err).Error("Failed to fetch VIP metrics")
		vipMetrics = &cache.VIPMetrics{} // Use empty metrics on error
	}

	// Get today's report
	todayReport, err := h.breakfastService.GetDailyReport(propertyID, time.Now())
	if err != nil {
		logging.WithError(err).Error("Failed to fetch daily report")
		todayReport = &services.DailyBreakfastReport{}
	}

	// Calculate additional KPIs
	kpis := ExecutiveKPIs{
		TotalVIPs:        vipMetrics.TotalVIPs,
		UpsetGuests:      vipMetrics.TotalUpset,
		SatisfactionRate: calculateSatisfactionRate(vipMetrics),
		AvgServiceTime:   12, // Mock data - would come from service analytics
		Revenue:          calculateRevenue(todayReport),
		OccupancyRate:    calculateOccupancyRate(propertyID),
		Trends: KPITrends{
			VIPTrend:          TrendData{Value: 12, Direction: "up", Period: "vs last week"},
			UpsetTrend:        TrendData{Value: -8, Direction: "down", Period: "vs last week"},
			SatisfactionTrend: TrendData{Value: 5, Direction: "up", Period: "vs last month"},
			ServiceTimeTrend:  TrendData{Value: -2, Direction: "down", Period: "improvement"},
			RevenueTrend:      TrendData{Value: 18, Direction: "up", Period: "vs last month"},
			OccupancyTrend:    TrendData{Value: 3, Direction: "up", Period: "vs last week"},
		},
		LastUpdated: time.Now(),
	}

	c.JSON(http.StatusOK, kpis)
}

// GetVIPTrends returns VIP guest trends over time
func (h *ExecutiveHandler) GetVIPTrends(c *gin.Context) {
	propertyID := c.Query("property_id")
	if propertyID == "" {
		propertyID = "HOTEL001"
	}
	
	period := c.Query("period") // week, month, year
	if period == "" {
		period = "week"
	}

	// Generate trend data based on period
	var trends VIPTrends
	
	switch period {
	case "week":
		trends = VIPTrends{
			Labels:     []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"},
			VIPCounts:  []int{42, 45, 43, 47, 48, 52, 47},
			UpsetCounts: []int{2, 1, 3, 2, 1, 2, 3},
			Period:     period,
		}
	case "month":
		trends = VIPTrends{
			Labels:     []string{"Week 1", "Week 2", "Week 3", "Week 4"},
			VIPCounts:  []int{180, 195, 188, 203},
			UpsetCounts: []int{8, 6, 9, 7},
			Period:     period,
		}
	case "year":
		trends = VIPTrends{
			Labels:     []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun"},
			VIPCounts:  []int{720, 780, 810, 890, 920, 960},
			UpsetCounts: []int{35, 28, 32, 25, 22, 20},
			Period:     period,
		}
	}

	c.JSON(http.StatusOK, trends)
}

// GetServicePerformance returns service performance metrics
func (h *ExecutiveHandler) GetServicePerformance(c *gin.Context) {
	propertyID := c.Query("property_id")
	if propertyID == "" {
		propertyID = "HOTEL001"
	}
	
	period := c.Query("period") // today, week, month
	if period == "" {
		period = "today"
	}

	var performance ServicePerformance
	
	switch period {
	case "today":
		performance = ServicePerformance{
			Labels: []string{"6AM", "7AM", "8AM", "9AM", "10AM", "11AM"},
			ServiceTimes: []float64{8, 12, 15, 13, 10, 8},
			Period: period,
			AverageTime: 11,
		}
	case "week":
		performance = ServicePerformance{
			Labels: []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"},
			ServiceTimes: []float64{12, 11, 13, 12, 11, 14, 12},
			Period: period,
			AverageTime: 12.1,
		}
	case "month":
		performance = ServicePerformance{
			Labels: []string{"Week 1", "Week 2", "Week 3", "Week 4"},
			ServiceTimes: []float64{12.5, 12.2, 11.8, 12.0},
			Period: period,
			AverageTime: 12.1,
		}
	}

	c.JSON(http.StatusOK, performance)
}

// GetRevenueAnalysis returns revenue analysis data
func (h *ExecutiveHandler) GetRevenueAnalysis(c *gin.Context) {
	propertyID := c.Query("property_id")
	if propertyID == "" {
		propertyID = "HOTEL001"
	}
	
	period := c.Query("period") // week, month, quarter
	if period == "" {
		period = "week"
	}

	var revenue RevenueAnalysis
	
	switch period {
	case "week":
		revenue = RevenueAnalysis{
			Labels: []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"},
			Amounts: []float64{1850, 2100, 1950, 2200, 2350, 2800, 2450},
			Total: 15700,
			Average: 2242.86,
			Period: period,
		}
	case "month":
		revenue = RevenueAnalysis{
			Labels: []string{"Week 1", "Week 2", "Week 3", "Week 4"},
			Amounts: []float64{14200, 15300, 14800, 16200},
			Total: 60500,
			Average: 15125,
			Period: period,
		}
	case "quarter":
		revenue = RevenueAnalysis{
			Labels: []string{"Jan", "Feb", "Mar"},
			Amounts: []float64{58000, 62000, 65500},
			Total: 185500,
			Average: 61833.33,
			Period: period,
		}
	}

	c.JSON(http.StatusOK, revenue)
}

// GetGuestPreferences returns guest preference distribution
func (h *ExecutiveHandler) GetGuestPreferences(c *gin.Context) {
	propertyID := c.Query("property_id")
	if propertyID == "" {
		propertyID = "HOTEL001"
	}

	preferences := GuestPreferences{
		Labels: []string{"Continental", "American", "Healthy", "Vegetarian", "Special Diet"},
		Counts: []int{35, 25, 20, 15, 5},
		Total: 100,
	}

	c.JSON(http.StatusOK, preferences)
}

// GetUpsetVIPGuests returns VIP guests requiring attention
func (h *ExecutiveHandler) GetUpsetVIPGuests(c *gin.Context) {
	propertyID := c.Query("property_id")
	if propertyID == "" {
		propertyID = "HOTEL001"
	}

	ctx := context.Background()
	
	// Get upset guests
	upsetGuests, err := h.breakfastService.GetUpsetGuests(ctx, propertyID)
	if err != nil {
		logging.WithError(err).Error("Failed to fetch upset guests")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch upset guests"})
		return
	}

	// Transform to executive view
	var executiveView []ExecutiveVIPGuest
	for _, guest := range upsetGuests {
		execGuest := ExecutiveVIPGuest{
			ID:          guest.ID,
			Name:        guest.FirstName + " " + guest.LastName,
			Room:        guest.RoomNumber,
			Status:      "upset",
			Issue:       guest.PMSSpecialReq, // Using special requests as issue indicator
			StayDuration: calculateStayDuration(guest.CheckInDate, guest.CheckOutDate),
			Preferences: guest.PMSSpecialReq, // Using special requests for preferences
			IsGold:      guest.IsVIP, // Using VIP status instead of loyalty tier
			LastUpdated: guest.UpdatedAt,
		}
		
		// Set default issue if empty
		if execGuest.Issue == "" {
			execGuest.Issue = "Service quality concern"
		}
		
		executiveView = append(executiveView, execGuest)
	}

	c.JSON(http.StatusOK, gin.H{
		"guests": executiveView,
		"total":  len(executiveView),
	})
}

// GetExecutiveAlerts returns active alerts for executives
func (h *ExecutiveHandler) GetExecutiveAlerts(c *gin.Context) {
	propertyID := c.Query("property_id")
	if propertyID == "" {
		propertyID = "HOTEL001"
	}

	ctx := context.Background()
	
	// Get upset guest count
	upsetGuests, _ := h.breakfastService.GetUpsetGuests(ctx, propertyID)
	upsetCount := len(upsetGuests)

	var alerts []ExecutiveAlert
	
	if upsetCount > 0 {
		alerts = append(alerts, ExecutiveAlert{
			Type:     "upset_guests",
			Severity: "high",
			Message:  fmt.Sprintf("%d VIP guests require immediate attention", upsetCount),
			Count:    upsetCount,
			Action:   "View upset guests",
		})
	}

	// Add other alerts based on business rules
	// Low occupancy alert
	occupancyRate := calculateOccupancyRate(propertyID)
	if occupancyRate < 70 {
		alerts = append(alerts, ExecutiveAlert{
			Type:     "low_occupancy",
			Severity: "medium",
			Message:  fmt.Sprintf("Occupancy rate at %d%%, below target", occupancyRate),
			Count:    1,
			Action:   "Review pricing strategy",
		})
	}

	// Service time alert
	avgServiceTime := 12.0 // This would come from actual metrics
	if avgServiceTime > 15 {
		alerts = append(alerts, ExecutiveAlert{
			Type:     "service_delay",
			Severity: "medium",
			Message:  "Average service time exceeding 15 minutes",
			Count:    1,
			Action:   "Review staffing levels",
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"alerts": alerts,
		"total":  len(alerts),
	})
}

// Helper functions
func calculateSatisfactionRate(metrics *cache.VIPMetrics) float64 {
	if metrics.TotalVIPs == 0 {
		return 100.0
	}
	return float64(metrics.TotalVIPs-metrics.TotalUpset) / float64(metrics.TotalVIPs) * 100
}

func calculateRevenue(report *services.DailyBreakfastReport) float64 {
	// Base breakfast price * consumed count
	basePrice := 25.0
	return float64(report.TotalConsumed) * basePrice
}

func calculateOccupancyRate(propertyID string) int {
	// This would query actual room occupancy
	// For demo, return a realistic value
	return 87
}

func calculateStayDuration(checkIn, checkOut time.Time) string {
	duration := checkOut.Sub(checkIn)
	nights := int(duration.Hours() / 24)
	if nights == 1 {
		return "1 night"
	}
	return fmt.Sprintf("%d nights", nights)
}

// Response structures
type ExecutiveKPIs struct {
	TotalVIPs        int       `json:"total_vips"`
	UpsetGuests      int       `json:"upset_guests"`
	SatisfactionRate float64   `json:"satisfaction_rate"`
	AvgServiceTime   int       `json:"avg_service_time"`
	Revenue          float64   `json:"revenue"`
	OccupancyRate    int       `json:"occupancy_rate"`
	Trends           KPITrends `json:"trends"`
	LastUpdated      time.Time `json:"last_updated"`
}

type KPITrends struct {
	VIPTrend          TrendData `json:"vip_trend"`
	UpsetTrend        TrendData `json:"upset_trend"`
	SatisfactionTrend TrendData `json:"satisfaction_trend"`
	ServiceTimeTrend  TrendData `json:"service_time_trend"`
	RevenueTrend      TrendData `json:"revenue_trend"`
	OccupancyTrend    TrendData `json:"occupancy_trend"`
}

type TrendData struct {
	Value     float64 `json:"value"`
	Direction string  `json:"direction"`
	Period    string  `json:"period"`
}

type VIPTrends struct {
	Labels      []string `json:"labels"`
	VIPCounts   []int    `json:"vip_counts"`
	UpsetCounts []int    `json:"upset_counts"`
	Period      string   `json:"period"`
}

type ServicePerformance struct {
	Labels       []string  `json:"labels"`
	ServiceTimes []float64 `json:"service_times"`
	Period       string    `json:"period"`
	AverageTime  float64   `json:"average_time"`
}

type RevenueAnalysis struct {
	Labels  []string  `json:"labels"`
	Amounts []float64 `json:"amounts"`
	Total   float64   `json:"total"`
	Average float64   `json:"average"`
	Period  string    `json:"period"`
}

type GuestPreferences struct {
	Labels []string `json:"labels"`
	Counts []int    `json:"counts"`
	Total  int      `json:"total"`
}

type ExecutiveVIPGuest struct {
	ID           uint      `json:"id"`
	Name         string    `json:"name"`
	Room         string    `json:"room"`
	Status       string    `json:"status"`
	Issue        string    `json:"issue"`
	StayDuration string    `json:"stay_duration"`
	Preferences  string    `json:"preferences"`
	IsGold       bool      `json:"is_gold"`
	LastUpdated  time.Time `json:"last_updated"`
}

type ExecutiveAlert struct {
	Type     string `json:"type"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
	Count    int    `json:"count"`
	Action   string `json:"action"`
}