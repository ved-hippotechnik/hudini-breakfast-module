package api

import (
	"context"
	"net/http"
	"runtime"
	"time"

	"hudini-breakfast-module/internal/cache"
	"hudini-breakfast-module/internal/database"
	"hudini-breakfast-module/internal/websocket"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type HealthHandler struct {
	db             *gorm.DB
	cache          *cache.RedisCache
	distributedHub *websocket.DistributedHub
	startTime      time.Time
	serverID       string
}

func NewHealthHandler(db *gorm.DB, cache *cache.RedisCache, hub *websocket.DistributedHub, serverID string) *HealthHandler {
	return &HealthHandler{
		db:             db,
		cache:          cache,
		distributedHub: hub,
		startTime:      time.Now(),
		serverID:       serverID,
	}
}

// Health check response
type HealthResponse struct {
	Status    string                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	ServerID  string                 `json:"server_id"`
	Uptime    string                 `json:"uptime"`
	Checks    map[string]HealthCheckStatus `json:"checks"`
}

// Individual health check
type HealthCheckStatus struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// GET /health - Basic health check
func (h *HealthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"server_id": h.serverID,
	})
}

// GET /health/detailed - Detailed health check
func (h *HealthHandler) DetailedHealth(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().UTC(),
		ServerID:  h.serverID,
		Uptime:    time.Since(h.startTime).String(),
		Checks:    make(map[string]HealthCheckStatus),
	}

	// Check database
	response.Checks["database"] = h.checkDatabase(ctx)
	
	// Check Redis
	response.Checks["redis"] = h.checkRedis(ctx)
	
	// Check WebSocket
	response.Checks["websocket"] = h.checkWebSocket(ctx)
	
	// Check memory usage
	response.Checks["memory"] = h.checkMemory()

	// Determine overall status
	for _, check := range response.Checks {
		if check.Status != "healthy" {
			response.Status = "degraded"
			if check.Status == "unhealthy" {
				response.Status = "unhealthy"
				c.JSON(http.StatusServiceUnavailable, response)
				return
			}
		}
	}

	c.JSON(http.StatusOK, response)
}

// GET /metrics - Prometheus-compatible metrics
func (h *HealthHandler) Metrics(c *gin.Context) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	dbStats, _ := database.GetStats(h.db)
	
	metrics := gin.H{
		"go_memstats_alloc_bytes":         m.Alloc,
		"go_memstats_sys_bytes":           m.Sys,
		"go_memstats_heap_alloc_bytes":    m.HeapAlloc,
		"go_memstats_heap_inuse_bytes":    m.HeapInuse,
		"go_memstats_heap_objects":        m.HeapObjects,
		"go_goroutines":                   runtime.NumGoroutine(),
		"db_open_connections":             dbStats.OpenConnections,
		"db_in_use":                       dbStats.InUse,
		"db_idle":                         dbStats.Idle,
		"websocket_connected_clients":     h.getWebSocketClients(),
		"uptime_seconds":                  time.Since(h.startTime).Seconds(),
		"server_id":                       h.serverID,
	}

	c.JSON(http.StatusOK, metrics)
}

// GET /ready - Readiness probe
func (h *HealthHandler) Ready(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Check critical dependencies
	if err := h.db.WithContext(ctx).Exec("SELECT 1").Error; err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "not ready",
			"reason": "database unavailable",
		})
		return
	}

	if h.cache != nil {
		if err := h.cache.Health(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "not ready",
				"reason": "cache unavailable",
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
	})
}

// GET /live - Liveness probe
func (h *HealthHandler) Live(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "alive",
	})
}

// Helper functions

func (h *HealthHandler) checkDatabase(ctx context.Context) HealthCheckStatus {
	var result int
	err := h.db.WithContext(ctx).Raw("SELECT 1").Scan(&result).Error
	
	if err != nil {
		return HealthCheckStatus{
			Status:  "unhealthy",
			Message: err.Error(),
		}
	}

	dbStats, _ := database.GetStats(h.db)
	if dbStats.OpenConnections > 20 {
		return HealthCheckStatus{
			Status:  "degraded",
			Message: "high connection count",
		}
	}

	return HealthCheckStatus{Status: "healthy"}
}

func (h *HealthHandler) checkRedis(ctx context.Context) HealthCheckStatus {
	if h.cache == nil {
		return HealthCheckStatus{
			Status:  "disabled",
			Message: "Redis not configured",
		}
	}

	err := h.cache.Health(ctx)
	if err != nil {
		return HealthCheckStatus{
			Status:  "unhealthy",
			Message: err.Error(),
		}
	}

	return HealthCheckStatus{Status: "healthy"}
}

func (h *HealthHandler) checkWebSocket(ctx context.Context) HealthCheckStatus {
	if h.distributedHub == nil {
		return HealthCheckStatus{
			Status:  "degraded",
			Message: "WebSocket hub not available",
		}
	}

	err := h.distributedHub.Health(ctx)
	if err != nil {
		return HealthCheckStatus{
			Status:  "unhealthy",
			Message: err.Error(),
		}
	}

	return HealthCheckStatus{Status: "healthy"}
}

func (h *HealthHandler) checkMemory() HealthCheckStatus {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Alert if memory usage is high (>1GB)
	if m.Alloc > 1024*1024*1024 {
		return HealthCheckStatus{
			Status:  "degraded",
			Message: "high memory usage",
		}
	}

	return HealthCheckStatus{Status: "healthy"}
}

func (h *HealthHandler) getWebSocketClients() int {
	if h.distributedHub != nil {
		count, _ := h.distributedHub.GetConnectedClients(context.Background())
		return count
	}
	return 0
}