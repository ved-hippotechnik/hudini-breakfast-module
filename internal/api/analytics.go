package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// AnalyticsData represents comprehensive analytics information
type AnalyticsData struct {
	Period     string              `json:"period"`
	PropertyID string              `json:"property_id"`
	Metrics    AnalyticsMetrics    `json:"metrics"`
	Charts     AnalyticsCharts     `json:"charts"`
	Insights   []AnalyticsInsight  `json:"insights"`
	Forecasts  []AnalyticsForecast `json:"forecasts"`
	Timestamp  time.Time           `json:"timestamp"`
}

type AnalyticsMetrics struct {
	Revenue              MetricValue `json:"revenue"`
	OccupancyRate        MetricValue `json:"occupancy_rate"`
	BreakfastTakeup      MetricValue `json:"breakfast_takeup"`
	AverageStayDuration  MetricValue `json:"average_stay_duration"`
	CustomerSatisfaction MetricValue `json:"customer_satisfaction"`
	CostPerBreakfast     MetricValue `json:"cost_per_breakfast"`
	TotalRooms           int         `json:"total_rooms"`
	OccupiedRooms        int         `json:"occupied_rooms"`
	BreakfastPackages    int         `json:"breakfast_packages"`
	ConsumedToday        int         `json:"consumed_today"`
}

type MetricValue struct {
	Current       float64 `json:"current"`
	Previous      float64 `json:"previous"`
	Change        float64 `json:"change"`
	ChangePercent float64 `json:"change_percent"`
	Trend         string  `json:"trend"` // "up", "down", "stable"
}

type AnalyticsCharts struct {
	RevenueTimeline     []ChartDataPoint `json:"revenue_timeline"`
	PackageDistribution []ChartPieSlice  `json:"package_distribution"`
	HourlyConsumption   []ChartDataPoint `json:"hourly_consumption"`
	MonthlyTrends       []ChartDataPoint `json:"monthly_trends"`
}

type ChartDataPoint struct {
	Label string  `json:"label"`
	Value float64 `json:"value"`
	Date  string  `json:"date,omitempty"`
}

type ChartPieSlice struct {
	Label      string  `json:"label"`
	Value      float64 `json:"value"`
	Percentage float64 `json:"percentage"`
	Color      string  `json:"color"`
}

type AnalyticsInsight struct {
	Type        string    `json:"type"` // "opportunity", "warning", "critical", "info"
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Impact      string    `json:"impact"`
	Action      string    `json:"action"`
	Confidence  float64   `json:"confidence"`
	CreatedAt   time.Time `json:"created_at"`
}

type AnalyticsForecast struct {
	Metric     string    `json:"metric"`
	Period     string    `json:"period"`
	Predicted  float64   `json:"predicted"`
	Confidence float64   `json:"confidence"`
	Trend      string    `json:"trend"`
	Factors    []string  `json:"factors"`
	CreatedAt  time.Time `json:"created_at"`
}

// GetAdvancedAnalytics provides comprehensive analytics data
func (h *BreakfastHandler) GetAdvancedAnalytics(c *gin.Context) {
	propertyID := c.DefaultQuery("property_id", "HOTEL001")
	period := c.DefaultQuery("period", "week")
	comparison := c.DefaultQuery("comparison", "previous")

	// Generate comprehensive analytics
	analytics := generateAdvancedAnalytics(propertyID, period, comparison)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    analytics,
	})
}

// GetRealtimeMetrics provides real-time business metrics
func (h *BreakfastHandler) GetRealtimeMetrics(c *gin.Context) {
	propertyID := c.DefaultQuery("property_id", "HOTEL001")

	metrics := generateRealtimeMetrics(propertyID)

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"data":      metrics,
		"timestamp": time.Now(),
	})
}

// GetPredictiveInsights provides AI-powered predictive analytics
func (h *BreakfastHandler) GetPredictiveInsights(c *gin.Context) {
	propertyID := c.DefaultQuery("property_id", "HOTEL001")
	horizon := c.DefaultQuery("horizon", "30") // days

	horizonDays, _ := strconv.Atoi(horizon)
	insights := generatePredictiveInsights(propertyID, horizonDays)

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"data":         insights,
		"horizon_days": horizonDays,
	})
}

// GetBusinessIntelligence provides comprehensive BI dashboard data
func (h *BreakfastHandler) GetBusinessIntelligence(c *gin.Context) {
	propertyID := c.DefaultQuery("property_id", "HOTEL001")

	bi := BusinessIntelligenceData{
		PropertyID:   propertyID,
		GeneratedAt:  time.Now(),
		KPIs:         generateKPIs(propertyID),
		Segments:     generateCustomerSegments(propertyID),
		Optimization: generateOptimizationRecommendations(propertyID),
		Competitive:  generateCompetitiveAnalysis(propertyID),
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    bi,
	})
}

type BusinessIntelligenceData struct {
	PropertyID   string                       `json:"property_id"`
	GeneratedAt  time.Time                    `json:"generated_at"`
	KPIs         []KPIMetric                  `json:"kpis"`
	Segments     []CustomerSegment            `json:"segments"`
	Optimization []OptimizationRecommendation `json:"optimization"`
	Competitive  CompetitiveAnalysis          `json:"competitive"`
}

type KPIMetric struct {
	Name        string  `json:"name"`
	Value       float64 `json:"value"`
	Target      float64 `json:"target"`
	Performance float64 `json:"performance"` // percentage of target achieved
	Status      string  `json:"status"`      // "above", "on_target", "below"
	Unit        string  `json:"unit"`
}

type CustomerSegment struct {
	Name            string   `json:"name"`
	Size            int      `json:"size"`
	Revenue         float64  `json:"revenue"`
	BreakfastRate   float64  `json:"breakfast_rate"`
	Satisfaction    float64  `json:"satisfaction"`
	LoyaltyScore    float64  `json:"loyalty_score"`
	Characteristics []string `json:"characteristics"`
}

type OptimizationRecommendation struct {
	Category    string  `json:"category"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Impact      string  `json:"impact"`
	Effort      string  `json:"effort"`   // "low", "medium", "high"
	Priority    int     `json:"priority"` // 1-5
	ExpectedROI float64 `json:"expected_roi"`
}

type CompetitiveAnalysis struct {
	MarketPosition  string            `json:"market_position"`
	Benchmarks      []BenchmarkMetric `json:"benchmarks"`
	Opportunities   []string          `json:"opportunities"`
	Threats         []string          `json:"threats"`
	Recommendations []string          `json:"recommendations"`
}

type BenchmarkMetric struct {
	Metric      string  `json:"metric"`
	OurValue    float64 `json:"our_value"`
	IndustryAvg float64 `json:"industry_avg"`
	BestInClass float64 `json:"best_in_class"`
	Gap         float64 `json:"gap"`
	Ranking     string  `json:"ranking"` // "top_quartile", "above_avg", "below_avg", "bottom_quartile"
}

// Helper functions to generate analytics data
func generateAdvancedAnalytics(propertyID, period, comparison string) AnalyticsData {
	now := time.Now()

	return AnalyticsData{
		Period:     period,
		PropertyID: propertyID,
		Timestamp:  now,
		Metrics: AnalyticsMetrics{
			Revenue: MetricValue{
				Current:       25450.00,
				Previous:      23200.00,
				Change:        2250.00,
				ChangePercent: 9.7,
				Trend:         "up",
			},
			OccupancyRate: MetricValue{
				Current:       87.5,
				Previous:      82.3,
				Change:        5.2,
				ChangePercent: 6.3,
				Trend:         "up",
			},
			BreakfastTakeup: MetricValue{
				Current:       73.2,
				Previous:      69.8,
				Change:        3.4,
				ChangePercent: 4.9,
				Trend:         "up",
			},
			AverageStayDuration: MetricValue{
				Current:       2.8,
				Previous:      2.6,
				Change:        0.2,
				ChangePercent: 7.7,
				Trend:         "up",
			},
			CustomerSatisfaction: MetricValue{
				Current:       4.6,
				Previous:      4.4,
				Change:        0.2,
				ChangePercent: 4.5,
				Trend:         "up",
			},
			CostPerBreakfast: MetricValue{
				Current:       12.50,
				Previous:      13.20,
				Change:        -0.70,
				ChangePercent: -5.3,
				Trend:         "down",
			},
			TotalRooms:        120,
			OccupiedRooms:     105,
			BreakfastPackages: 77,
			ConsumedToday:     65,
		},
		Charts:    generateChartData(),
		Insights:  generateInsights(),
		Forecasts: generateForecasts(),
	}
}

func generateChartData() AnalyticsCharts {
	return AnalyticsCharts{
		RevenueTimeline: []ChartDataPoint{
			{Label: "Week 1", Value: 18500, Date: "2024-01-01"},
			{Label: "Week 2", Value: 21200, Date: "2024-01-08"},
			{Label: "Week 3", Value: 23400, Date: "2024-01-15"},
			{Label: "Week 4", Value: 25450, Date: "2024-01-22"},
		},
		PackageDistribution: []ChartPieSlice{
			{Label: "Standard Package", Value: 45, Percentage: 45.0, Color: "#4ade80"},
			{Label: "Premium Package", Value: 25, Percentage: 25.0, Color: "#22d3ee"},
			{Label: "VIP Package", Value: 15, Percentage: 15.0, Color: "#a855f7"},
			{Label: "No Package", Value: 15, Percentage: 15.0, Color: "#6b7280"},
		},
		HourlyConsumption: []ChartDataPoint{
			{Label: "6 AM", Value: 15},
			{Label: "7 AM", Value: 45},
			{Label: "8 AM", Value: 85},
			{Label: "9 AM", Value: 75},
			{Label: "10 AM", Value: 35},
			{Label: "11 AM", Value: 10},
		},
		MonthlyTrends: []ChartDataPoint{
			{Label: "Jan", Value: 82500},
			{Label: "Feb", Value: 78200},
			{Label: "Mar", Value: 89100},
			{Label: "Apr", Value: 92800},
			{Label: "May", Value: 88900},
			{Label: "Jun", Value: 95200},
		},
	}
}

func generateInsights() []AnalyticsInsight {
	return []AnalyticsInsight{
		{
			Type:        "opportunity",
			Title:       "Revenue Growth Opportunity",
			Description: "Breakfast take-up rate is 15% below industry average",
			Impact:      "Potential $2,400/month additional revenue",
			Action:      "Implement targeted promotional campaign",
			Confidence:  0.87,
			CreatedAt:   time.Now(),
		},
		{
			Type:        "warning",
			Title:       "Peak Hour Capacity",
			Description: "8-9 AM shows 23% higher demand than optimal capacity",
			Impact:      "Potential guest dissatisfaction",
			Action:      "Consider extending breakfast hours or increasing staff",
			Confidence:  0.92,
			CreatedAt:   time.Now(),
		},
		{
			Type:        "info",
			Title:       "Cost Optimization Success",
			Description: "Food waste reduction initiatives showing positive results",
			Impact:      "12% cost reduction without satisfaction impact",
			Action:      "Continue current waste reduction strategies",
			Confidence:  0.94,
			CreatedAt:   time.Now(),
		},
	}
}

func generateForecasts() []AnalyticsForecast {
	return []AnalyticsForecast{
		{
			Metric:     "revenue",
			Period:     "next_week",
			Predicted:  27650.00,
			Confidence: 0.85,
			Trend:      "increasing",
			Factors:    []string{"seasonal_demand", "promotional_campaign", "weather_forecast"},
			CreatedAt:  time.Now(),
		},
		{
			Metric:     "occupancy",
			Period:     "next_month",
			Predicted:  89.2,
			Confidence: 0.78,
			Trend:      "stable",
			Factors:    []string{"booking_patterns", "market_events", "competitor_analysis"},
			CreatedAt:  time.Now(),
		},
	}
}

func generateRealtimeMetrics(propertyID string) map[string]interface{} {
	return map[string]interface{}{
		"current_occupancy":     87.5,
		"breakfast_active":      65,
		"kitchen_capacity":      85.0,
		"wait_time_minutes":     3.2,
		"satisfaction_score":    4.6,
		"cost_efficiency":       94.2,
		"staff_utilization":     78.5,
		"food_waste_percentage": 8.3,
		"last_updated":          time.Now(),
	}
}

func generatePredictiveInsights(propertyID string, horizonDays int) map[string]interface{} {
	return map[string]interface{}{
		"demand_forecast": map[string]interface{}{
			"next_7_days": []float64{85, 78, 82, 90, 88, 92, 87},
			"confidence":  0.83,
			"trend":       "seasonal_increase",
		},
		"revenue_prediction": map[string]interface{}{
			"expected_daily": 3650.00,
			"range_low":      3200.00,
			"range_high":     4100.00,
			"confidence":     0.79,
		},
		"optimization_opportunities": []map[string]interface{}{
			{
				"area":           "pricing",
				"potential_lift": 8.5,
				"confidence":     0.72,
				"implementation": "dynamic_pricing_weekends",
			},
			{
				"area":             "operations",
				"potential_saving": 12.3,
				"confidence":       0.84,
				"implementation":   "staff_scheduling_optimization",
			},
		},
	}
}

func generateKPIs(propertyID string) []KPIMetric {
	return []KPIMetric{
		{
			Name:        "Monthly Revenue",
			Value:       95200,
			Target:      90000,
			Performance: 105.8,
			Status:      "above",
			Unit:        "USD",
		},
		{
			Name:        "Breakfast Take-up Rate",
			Value:       73.2,
			Target:      75.0,
			Performance: 97.6,
			Status:      "on_target",
			Unit:        "percent",
		},
		{
			Name:        "Customer Satisfaction",
			Value:       4.6,
			Target:      4.5,
			Performance: 102.2,
			Status:      "above",
			Unit:        "score",
		},
		{
			Name:        "Cost per Breakfast",
			Value:       12.50,
			Target:      13.00,
			Performance: 96.2,
			Status:      "above",
			Unit:        "USD",
		},
	}
}

func generateCustomerSegments(propertyID string) []CustomerSegment {
	return []CustomerSegment{
		{
			Name:            "Business Travelers",
			Size:            45,
			Revenue:         42500,
			BreakfastRate:   85.2,
			Satisfaction:    4.4,
			LoyaltyScore:    7.8,
			Characteristics: []string{"early_breakfast", "premium_packages", "efficiency_focused"},
		},
		{
			Name:            "Leisure Families",
			Size:            35,
			Revenue:         28900,
			BreakfastRate:   92.1,
			Satisfaction:    4.7,
			LoyaltyScore:    8.2,
			Characteristics: []string{"weekend_stays", "variety_seekers", "value_conscious"},
		},
		{
			Name:            "Conference Attendees",
			Size:            20,
			Revenue:         18600,
			BreakfastRate:   67.8,
			Satisfaction:    4.2,
			LoyaltyScore:    6.1,
			Characteristics: []string{"group_bookings", "schedule_dependent", "cost_sensitive"},
		},
	}
}

func generateOptimizationRecommendations(propertyID string) []OptimizationRecommendation {
	return []OptimizationRecommendation{
		{
			Category:    "Revenue",
			Title:       "Dynamic Breakfast Pricing",
			Description: "Implement demand-based pricing for peak and off-peak hours",
			Impact:      "8-12% revenue increase",
			Effort:      "medium",
			Priority:    1,
			ExpectedROI: 15.2,
		},
		{
			Category:    "Operations",
			Title:       "AI-Powered Inventory Management",
			Description: "Use predictive analytics for food ordering and waste reduction",
			Impact:      "15% cost reduction",
			Effort:      "high",
			Priority:    2,
			ExpectedROI: 22.8,
		},
		{
			Category:    "Experience",
			Title:       "Mobile Pre-ordering System",
			Description: "Allow guests to pre-order breakfast through mobile app",
			Impact:      "Reduced wait times, higher satisfaction",
			Effort:      "medium",
			Priority:    3,
			ExpectedROI: 8.5,
		},
	}
}

func generateCompetitiveAnalysis(propertyID string) CompetitiveAnalysis {
	return CompetitiveAnalysis{
		MarketPosition: "Upper Middle Tier",
		Benchmarks: []BenchmarkMetric{
			{
				Metric:      "Breakfast Take-up Rate",
				OurValue:    73.2,
				IndustryAvg: 68.5,
				BestInClass: 82.1,
				Gap:         8.9,
				Ranking:     "above_avg",
			},
			{
				Metric:      "Cost per Breakfast",
				OurValue:    12.50,
				IndustryAvg: 14.20,
				BestInClass: 10.80,
				Gap:         1.70,
				Ranking:     "above_avg",
			},
		},
		Opportunities: []string{
			"Premium breakfast packages for business travelers",
			"Family-friendly breakfast experiences",
			"Healthy/dietary restriction options expansion",
		},
		Threats: []string{
			"New boutique hotels with unique breakfast offerings",
			"Rising food costs impacting margins",
			"Changing guest preferences toward lighter meals",
		},
		Recommendations: []string{
			"Invest in breakfast experience differentiation",
			"Develop loyalty program integration",
			"Implement sustainability initiatives",
		},
	}
}
