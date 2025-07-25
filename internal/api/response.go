package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type APIResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Error     *APIError   `json:"error,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
	RequestID string      `json:"request_id,omitempty"`
}

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

type PaginatedResponse struct {
	APIResponse
	Pagination *PaginationMeta `json:"pagination,omitempty"`
}

type PaginationMeta struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

func SuccessResponse(c *gin.Context, data interface{}) {
	response := APIResponse{
		Success:   true,
		Data:      data,
		Timestamp: time.Now(),
		RequestID: getRequestID(c),
	}
	c.JSON(http.StatusOK, response)
}

func SuccessResponseWithMessage(c *gin.Context, message string, data interface{}) {
	response := APIResponse{
		Success:   true,
		Message:   message,
		Data:      data,
		Timestamp: time.Now(),
		RequestID: getRequestID(c),
	}
	c.JSON(http.StatusOK, response)
}

func CreatedResponse(c *gin.Context, data interface{}) {
	response := APIResponse{
		Success:   true,
		Message:   "Resource created successfully",
		Data:      data,
		Timestamp: time.Now(),
		RequestID: getRequestID(c),
	}
	c.JSON(http.StatusCreated, response)
}

func ErrorResponse(c *gin.Context, statusCode int, code, message string) {
	response := APIResponse{
		Success: false,
		Error: &APIError{
			Code:    code,
			Message: message,
		},
		Timestamp: time.Now(),
		RequestID: getRequestID(c),
	}
	c.JSON(statusCode, response)
}

func ErrorResponseWithDetails(c *gin.Context, statusCode int, code, message, details string) {
	response := APIResponse{
		Success: false,
		Error: &APIError{
			Code:    code,
			Message: message,
			Details: details,
		},
		Timestamp: time.Now(),
		RequestID: getRequestID(c),
	}
	c.JSON(statusCode, response)
}

func ValidationErrorResponse(c *gin.Context, details string) {
	ErrorResponseWithDetails(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid input data", details)
}

func NotFoundResponse(c *gin.Context, resource string) {
	ErrorResponse(c, http.StatusNotFound, "NOT_FOUND", resource+" not found")
}

func InternalErrorResponse(c *gin.Context, err error) {
	ErrorResponseWithDetails(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error", err.Error())
}

func UnauthorizedResponse(c *gin.Context) {
	ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
}

func ForbiddenResponse(c *gin.Context) {
	ErrorResponse(c, http.StatusForbidden, "FORBIDDEN", "Access denied")
}

func PaginatedSuccessResponse(c *gin.Context, data interface{}, pagination *PaginationMeta) {
	response := PaginatedResponse{
		APIResponse: APIResponse{
			Success:   true,
			Data:      data,
			Timestamp: time.Now(),
			RequestID: getRequestID(c),
		},
		Pagination: pagination,
	}
	c.JSON(http.StatusOK, response)
}

func getRequestID(c *gin.Context) string {
	if requestID, exists := c.Get("request_id"); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}