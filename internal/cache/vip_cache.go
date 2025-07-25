package cache

import (
	"context"
	"fmt"
	"time"

	"hudini-breakfast-module/internal/models"
)

// VIPCache provides caching specifically for VIP-related queries
type VIPCache struct {
	cache *RedisCache
	ttl   time.Duration
}

// Cache key prefixes
const (
	VIPGuestListKey     = "hudini:vip:guests:property:%s"
	VIPGuestKey         = "hudini:vip:guest:%d"
	UpsetGuestListKey   = "hudini:vip:upset:property:%s"
	VIPMetricsKey       = "hudini:vip:metrics:property:%s"
	GuestPreferencesKey = "hudini:guest:preferences:%d"
	RoomStatusKey       = "hudini:room:status:%s:%s" // property:room_number
	DailyStatsKey       = "hudini:stats:daily:%s:%s" // property:date
)

// NewVIPCache creates a new VIP cache instance
func NewVIPCache(redisCache *RedisCache, ttl time.Duration) *VIPCache {
	return &VIPCache{
		cache: redisCache,
		ttl:   ttl,
	}
}

// GetVIPGuests retrieves cached VIP guests for a property
func (v *VIPCache) GetVIPGuests(ctx context.Context, propertyID string) ([]models.Guest, error) {
	key := fmt.Sprintf(VIPGuestListKey, propertyID)
	var guests []models.Guest
	
	err := v.cache.GetJSON(ctx, key, &guests)
	if err != nil {
		return nil, err
	}
	
	return guests, nil
}

// SetVIPGuests caches VIP guests for a property
func (v *VIPCache) SetVIPGuests(ctx context.Context, propertyID string, guests []models.Guest) error {
	key := fmt.Sprintf(VIPGuestListKey, propertyID)
	return v.cache.SetJSON(ctx, key, guests, v.ttl)
}

// GetVIPGuest retrieves a cached VIP guest by ID
func (v *VIPCache) GetVIPGuest(ctx context.Context, guestID uint) (*models.Guest, error) {
	key := fmt.Sprintf(VIPGuestKey, guestID)
	var guest models.Guest
	
	err := v.cache.GetJSON(ctx, key, &guest)
	if err != nil {
		return nil, err
	}
	
	return &guest, nil
}

// SetVIPGuest caches a VIP guest
func (v *VIPCache) SetVIPGuest(ctx context.Context, guest *models.Guest) error {
	if !guest.IsVIP {
		return nil // Only cache VIP guests
	}
	
	key := fmt.Sprintf(VIPGuestKey, guest.ID)
	return v.cache.SetJSON(ctx, key, guest, v.ttl)
}

// GetUpsetGuests retrieves cached upset guests for a property
func (v *VIPCache) GetUpsetGuests(ctx context.Context, propertyID string) ([]models.Guest, error) {
	key := fmt.Sprintf(UpsetGuestListKey, propertyID)
	var guests []models.Guest
	
	err := v.cache.GetJSON(ctx, key, &guests)
	if err != nil {
		return nil, err
	}
	
	return guests, nil
}

// SetUpsetGuests caches upset guests for a property
func (v *VIPCache) SetUpsetGuests(ctx context.Context, propertyID string, guests []models.Guest) error {
	key := fmt.Sprintf(UpsetGuestListKey, propertyID)
	return v.cache.SetJSON(ctx, key, guests, v.ttl)
}

// GetVIPMetrics retrieves cached VIP metrics for a property
func (v *VIPCache) GetVIPMetrics(ctx context.Context, propertyID string) (*VIPMetrics, error) {
	key := fmt.Sprintf(VIPMetricsKey, propertyID)
	var metrics VIPMetrics
	
	err := v.cache.GetJSON(ctx, key, &metrics)
	if err != nil {
		return nil, err
	}
	
	return &metrics, nil
}

// SetVIPMetrics caches VIP metrics for a property
func (v *VIPCache) SetVIPMetrics(ctx context.Context, propertyID string, metrics *VIPMetrics) error {
	key := fmt.Sprintf(VIPMetricsKey, propertyID)
	return v.cache.SetJSON(ctx, key, metrics, v.ttl)
}

// GetGuestPreferences retrieves cached guest preferences
func (v *VIPCache) GetGuestPreferences(ctx context.Context, guestID uint) (*models.GuestPreference, error) {
	key := fmt.Sprintf(GuestPreferencesKey, guestID)
	var prefs models.GuestPreference
	
	err := v.cache.GetJSON(ctx, key, &prefs)
	if err != nil {
		return nil, err
	}
	
	return &prefs, nil
}

// SetGuestPreferences caches guest preferences
func (v *VIPCache) SetGuestPreferences(ctx context.Context, prefs *models.GuestPreference) error {
	key := fmt.Sprintf(GuestPreferencesKey, prefs.GuestID)
	return v.cache.SetJSON(ctx, key, prefs, v.ttl*2) // Longer TTL for preferences
}

// GetRoomStatus retrieves cached room status
func (v *VIPCache) GetRoomStatus(ctx context.Context, propertyID, roomNumber string) (*models.RoomBreakfastStatus, error) {
	key := fmt.Sprintf(RoomStatusKey, propertyID, roomNumber)
	var status models.RoomBreakfastStatus
	
	err := v.cache.GetJSON(ctx, key, &status)
	if err != nil {
		return nil, err
	}
	
	return &status, nil
}

// SetRoomStatus caches room status
func (v *VIPCache) SetRoomStatus(ctx context.Context, status *models.RoomBreakfastStatus) error {
	key := fmt.Sprintf(RoomStatusKey, status.PropertyID, status.RoomNumber)
	return v.cache.SetJSON(ctx, key, status, 1*time.Minute) // Short TTL for real-time data
}

// InvalidateGuestCache invalidates all cache entries for a guest
func (v *VIPCache) InvalidateGuestCache(ctx context.Context, guestID uint, propertyID string) error {
	keys := []string{
		fmt.Sprintf(VIPGuestKey, guestID),
		fmt.Sprintf(GuestPreferencesKey, guestID),
		fmt.Sprintf(VIPGuestListKey, propertyID),
		fmt.Sprintf(UpsetGuestListKey, propertyID),
		fmt.Sprintf(VIPMetricsKey, propertyID),
	}
	
	return v.cache.Delete(ctx, keys...)
}

// InvalidatePropertyCache invalidates all cache entries for a property
func (v *VIPCache) InvalidatePropertyCache(ctx context.Context, propertyID string) error {
	// In a production system, we'd use Redis SCAN to find all keys with the property prefix
	// For now, we'll invalidate known key patterns
	keys := []string{
		fmt.Sprintf(VIPGuestListKey, propertyID),
		fmt.Sprintf(UpsetGuestListKey, propertyID),
		fmt.Sprintf(VIPMetricsKey, propertyID),
		fmt.Sprintf(DailyStatsKey, propertyID, time.Now().Format("2006-01-02")),
	}
	
	return v.cache.Delete(ctx, keys...)
}

// GetDailyStats retrieves cached daily statistics
func (v *VIPCache) GetDailyStats(ctx context.Context, propertyID string, date string) (*DailyStats, error) {
	key := fmt.Sprintf(DailyStatsKey, propertyID, date)
	var stats DailyStats
	
	err := v.cache.GetJSON(ctx, key, &stats)
	if err != nil {
		return nil, err
	}
	
	return &stats, nil
}

// SetDailyStats caches daily statistics
func (v *VIPCache) SetDailyStats(ctx context.Context, propertyID string, date string, stats *DailyStats) error {
	key := fmt.Sprintf(DailyStatsKey, propertyID, date)
	// Cache until end of day
	endOfDay := time.Now().Truncate(24*time.Hour).Add(24*time.Hour).Sub(time.Now())
	return v.cache.SetJSON(ctx, key, stats, endOfDay)
}

// VIPMetrics represents aggregated VIP metrics
type VIPMetrics struct {
	TotalVIPs           int     `json:"total_vips"`
	TotalUpset          int     `json:"total_upset"`
	VIPBreakfastRate    float64 `json:"vip_breakfast_rate"`
	AverageStayDuration float64 `json:"average_stay_duration"`
	TopPreferences      []string `json:"top_preferences"`
	LastUpdated         time.Time `json:"last_updated"`
}

// DailyStats represents daily statistics
type DailyStats struct {
	Date                  string  `json:"date"`
	TotalGuests          int     `json:"total_guests"`
	TotalBreakfasts      int     `json:"total_breakfasts"`
	VIPBreakfasts        int     `json:"vip_breakfasts"`
	ConsumptionRate      float64 `json:"consumption_rate"`
	PeakHour             string  `json:"peak_hour"`
	AverageServiceTime   float64 `json:"average_service_time"`
	LastUpdated          time.Time `json:"last_updated"`
}