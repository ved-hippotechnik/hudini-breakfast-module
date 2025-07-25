package api

import (
	"net/http"
	"strconv"
	"time"

	"hudini-breakfast-module/internal/audit"
	"hudini-breakfast-module/internal/services"

	"github.com/gin-gonic/gin"
)

// AuditHandler handles audit log API endpoints
type AuditHandler struct {
	auditService *services.AuditService
}

// NewAuditHandler creates a new audit handler
func NewAuditHandler(auditService *services.AuditService) *AuditHandler {
	return &AuditHandler{
		auditService: auditService,
	}
}

// GetAuditLogs retrieves audit logs with filters
// GET /api/audit/logs
func (h *AuditHandler) GetAuditLogs(c *gin.Context) {
	// Parse filters from query parameters
	filters := services.AuditFilters{
		Limit:     50, // Default limit
		Offset:    0,
		SortOrder: "desc",
	}

	// User ID filter
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		if userID, err := strconv.ParseUint(userIDStr, 10, 32); err == nil {
			uid := uint(userID)
			filters.UserID = &uid
		}
	}

	// Action filter
	if action := c.Query("action"); action != "" {
		filters.Action = action
	}

	// Resource filter
	if resource := c.Query("resource"); resource != "" {
		filters.Resource = resource
	}

	// Resource ID filter
	if resourceID := c.Query("resource_id"); resourceID != "" {
		filters.ResourceID = resourceID
	}

	// Status filter
	if status := c.Query("status"); status != "" {
		filters.Status = status
	}

	// IP address filter
	if ipAddress := c.Query("ip_address"); ipAddress != "" {
		filters.IPAddress = ipAddress
	}

	// Date range filters
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if startDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
			filters.StartDate = startDate
		}
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if endDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
			// Set to end of day
			filters.EndDate = endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		}
	}

	// Pagination
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 1000 {
			filters.Limit = limit
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			filters.Offset = offset
		}
	}

	// Sort order
	if sortOrder := c.Query("sort_order"); sortOrder == "asc" || sortOrder == "desc" {
		filters.SortOrder = sortOrder
	}

	// Get audit logs
	logs, total, err := h.auditService.GetAuditLogs(c.Request.Context(), filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve audit logs",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"logs":   logs,
			"total":  total,
			"limit":  filters.Limit,
			"offset": filters.Offset,
		},
	})
}

// GetUserActivity retrieves audit logs for a specific user
// GET /api/audit/users/:user_id/activity
func (h *AuditHandler) GetUserActivity(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID",
		})
		return
	}

	// Get limit from query parameter
	limit := 100 // Default limit
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 500 {
			limit = l
		}
	}

	// Get user activity
	logs, err := h.auditService.GetUserActivity(c.Request.Context(), uint(userID), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve user activity",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"user_id": userID,
			"logs":    logs,
			"count":   len(logs),
		},
	})
}

// GetResourceHistory retrieves audit logs for a specific resource
// GET /api/audit/resources/:resource/:resource_id/history
func (h *AuditHandler) GetResourceHistory(c *gin.Context) {
	resource := c.Param("resource")
	resourceID := c.Param("resource_id")

	if resource == "" || resourceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Resource and resource ID are required",
		})
		return
	}

	// Validate resource type
	validResources := map[string]audit.AuditResource{
		"guest":       audit.ResourceGuest,
		"room":        audit.ResourceRoom,
		"consumption": audit.ResourceConsumption,
		"staff":       audit.ResourceStaff,
		"property":    audit.ResourceProperty,
		"outlet":      audit.ResourceOutlet,
	}

	auditResource, valid := validResources[resource]
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid resource type",
		})
		return
	}

	// Get resource history
	logs, err := h.auditService.GetResourceHistory(c.Request.Context(), auditResource, resourceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve resource history",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"resource":    resource,
			"resource_id": resourceID,
			"logs":        logs,
			"count":       len(logs),
		},
	})
}

// GetAuditSummary provides a summary of audit activity
// GET /api/audit/summary
func (h *AuditHandler) GetAuditSummary(c *gin.Context) {
	// Parse date range
	var startDate, endDate time.Time
	
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if sd, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = sd
		}
	} else {
		// Default to last 7 days
		startDate = time.Now().AddDate(0, 0, -7)
	}
	
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if ed, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = ed.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		}
	} else {
		endDate = time.Now()
	}

	// Get summary data
	// This would typically involve aggregated queries
	// For now, return a basic structure
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"date_range": gin.H{
				"start": startDate,
				"end":   endDate,
			},
			"summary": gin.H{
				"total_actions":     0, // Would be populated from DB
				"failed_actions":    0,
				"unique_users":      0,
				"top_actions":       []gin.H{},
				"top_resources":     []gin.H{},
				"failed_by_resource": []gin.H{},
			},
		},
	})
}