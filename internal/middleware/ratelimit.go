package middleware

import (
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type RateLimiter struct {
	requests map[string][]time.Time
	mutex    sync.RWMutex
	limit    int
	window   time.Duration
}

func NewRateLimiter() *RateLimiter {
	// Get configuration from environment
	limit, _ := strconv.Atoi(os.Getenv("RATE_LIMIT_REQUESTS"))
	if limit <= 0 {
		limit = 100 // Default
	}

	window, _ := strconv.Atoi(os.Getenv("RATE_LIMIT_WINDOW"))
	if window <= 0 {
		window = 60 // Default 60 seconds
	}

	rl := &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   time.Duration(window) * time.Second,
	}

	// Cleanup old entries every minute
	go rl.cleanup()

	return rl
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mutex.Lock()
		now := time.Now()
		for ip, requests := range rl.requests {
			filtered := make([]time.Time, 0)
			for _, req := range requests {
				if now.Sub(req) < rl.window {
					filtered = append(filtered, req)
				}
			}
			if len(filtered) == 0 {
				delete(rl.requests, ip)
			} else {
				rl.requests[ip] = filtered
			}
		}
		rl.mutex.Unlock()
	}
}

func (rl *RateLimiter) Allow(ip string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	requests := rl.requests[ip]

	// Remove old requests
	filtered := make([]time.Time, 0)
	for _, req := range requests {
		if now.Sub(req) < rl.window {
			filtered = append(filtered, req)
		}
	}

	// Check if limit exceeded
	if len(filtered) >= rl.limit {
		return false
	}

	// Add current request
	filtered = append(filtered, now)
	rl.requests[ip] = filtered

	return true
}

func RateLimit() gin.HandlerFunc {
	rateLimiter := NewRateLimiter()

	return func(c *gin.Context) {
		ip := c.ClientIP()

		if !rateLimiter.Allow(ip) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded. Please try again later.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
