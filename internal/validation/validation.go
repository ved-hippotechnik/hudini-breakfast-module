package validation

import (
	"fmt"
	"net/http"
	"strings"

	"hudini-breakfast-module/internal/logging"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

// ValidateJSON validates JSON request body against a struct
func ValidateJSON(model interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := c.ShouldBindJSON(model); err != nil {
			logging.WithFields(logrus.Fields{
				"handler": "ValidateJSON",
				"error":   err.Error(),
			}).Warn("JSON validation failed")
			
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "VALIDATION_ERROR",
					"message": formatValidationErrors(err),
				},
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// ValidateQuery validates query parameters
func ValidateQuery(requiredParams ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var missingParams []string
		
		for _, param := range requiredParams {
			if value := c.Query(param); value == "" {
				missingParams = append(missingParams, param)
			}
		}
		
		if len(missingParams) > 0 {
			logging.WithFields(logrus.Fields{
				"handler":        "ValidateQuery",
				"missing_params": missingParams,
			}).Warn("Query validation failed")
			
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "VALIDATION_ERROR",
					"message": fmt.Sprintf("Missing required query parameters: %s", strings.Join(missingParams, ", ")),
				},
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// ValidateStruct validates a struct using validator tags
func ValidateStruct(s interface{}) error {
	return validate.Struct(s)
}

// ValidatePropertyID validates property ID format
func ValidatePropertyID() gin.HandlerFunc {
	return func(c *gin.Context) {
		propertyID := c.Query("property_id")
		if propertyID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "VALIDATION_ERROR",
					"message": "property_id is required",
				},
			})
			c.Abort()
			return
		}
		
		// Basic validation - could be enhanced with regex
		if len(propertyID) < 3 || len(propertyID) > 50 {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "VALIDATION_ERROR",
					"message": "property_id must be between 3 and 50 characters",
				},
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// ValidateRoomNumber validates room number format
func ValidateRoomNumber() gin.HandlerFunc {
	return func(c *gin.Context) {
		roomNumber := c.Param("room_number")
		if roomNumber == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "VALIDATION_ERROR",
					"message": "room_number is required",
				},
			})
			c.Abort()
			return
		}
		
		// Basic validation - could be enhanced with regex
		if len(roomNumber) < 1 || len(roomNumber) > 10 {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "VALIDATION_ERROR",
					"message": "room_number must be between 1 and 10 characters",
				},
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// ValidateDateFormat validates date format (YYYY-MM-DD)
func ValidateDateFormat(paramName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		dateStr := c.Query(paramName)
		if dateStr == "" {
			c.Next()
			return
		}
		
		// Simple regex for YYYY-MM-DD format
		if !isValidDateFormat(dateStr) {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "VALIDATION_ERROR",
					"message": fmt.Sprintf("%s must be in YYYY-MM-DD format", paramName),
				},
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// RequestSizeLimit limits request body size
func RequestSizeLimit(maxSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > maxSize {
			logging.WithFields(logrus.Fields{
				"handler":       "RequestSizeLimit",
				"content_length": c.Request.ContentLength,
				"max_size":      maxSize,
			}).Warn("Request size exceeded limit")
			
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "REQUEST_TOO_LARGE",
					"message": "Request body too large",
				},
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// Helper functions
func formatValidationErrors(err error) string {
	var messages []string
	
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			messages = append(messages, fmt.Sprintf("Field '%s' %s", e.Field(), getValidationMessage(e)))
		}
	} else {
		messages = append(messages, err.Error())
	}
	
	return strings.Join(messages, "; ")
}

func getValidationMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "is required"
	case "email":
		return "must be a valid email address"
	case "min":
		return fmt.Sprintf("must be at least %s characters long", e.Param())
	case "max":
		return fmt.Sprintf("must be at most %s characters long", e.Param())
	case "numeric":
		return "must be a number"
	default:
		return fmt.Sprintf("failed validation for tag '%s'", e.Tag())
	}
}

func isValidDateFormat(dateStr string) bool {
	// Simple regex for YYYY-MM-DD format
	if len(dateStr) != 10 {
		return false
	}
	
	// Basic checks for format
	if dateStr[4] != '-' || dateStr[7] != '-' {
		return false
	}
	
	// Check if all other characters are digits
	for i, char := range dateStr {
		if i == 4 || i == 7 {
			continue
		}
		if char < '0' || char > '9' {
			return false
		}
	}
	
	return true
}