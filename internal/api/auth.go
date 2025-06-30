package api

import (
	"net/http"
	"strings"
	"time"

	"hudini-breakfast-module/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthHandler struct {
	db        *gorm.DB
	jwtSecret string
}

func NewAuthHandler(db *gorm.DB, jwtSecret string) *AuthHandler {
	return &AuthHandler{
		db:        db,
		jwtSecret: jwtSecret,
	}
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required,min=6"`
	FirstName  string `json:"first_name" binding:"required"`
	LastName   string `json:"last_name" binding:"required"`
	Role       string `json:"role"` // staff, manager, admin
	PropertyID string `json:"property_id" binding:"required"`
}

type Claims struct {
	StaffID    uint   `json:"staff_id"`
	Email      string `json:"email"`
	Role       string `json:"role"`
	PropertyID string `json:"property_id"`
	jwt.RegisteredClaims
}

// POST /api/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate role
	validRoles := map[string]bool{
		"staff":   true,
		"manager": true,
		"admin":   true,
	}
	if req.Role == "" {
		req.Role = "staff" // Default role
	}
	if !validRoles[req.Role] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role"})
		return
	}

	// Check if staff member already exists
	var existingStaff models.Staff
	if err := h.db.Where("email = ?", req.Email).First(&existingStaff).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Staff member already exists"})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Create staff member
	staff := models.Staff{
		Email:      req.Email,
		Password:   string(hashedPassword),
		FirstName:  req.FirstName,
		LastName:   req.LastName,
		Role:       req.Role,
		PropertyID: req.PropertyID,
		IsActive:   true,
	}

	if err := h.db.Create(&staff).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create staff member"})
		return
	}

	// Generate JWT token
	token, err := h.generateToken(staff)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Staff member registered successfully",
		"token":   token,
		"staff": gin.H{
			"id":          staff.ID,
			"email":       staff.Email,
			"first_name":  staff.FirstName,
			"last_name":   staff.LastName,
			"role":        staff.Role,
			"property_id": staff.PropertyID,
		},
	})
}

// POST /api/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find staff member
	var staff models.Staff
	if err := h.db.Where("email = ? AND is_active = ?", req.Email, true).First(&staff).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(staff.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate JWT token
	token, err := h.generateToken(staff)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"staff": gin.H{
			"id":          staff.ID,
			"email":       staff.Email,
			"first_name":  staff.FirstName,
			"last_name":   staff.LastName,
			"role":        staff.Role,
			"property_id": staff.PropertyID,
		},
	})
}

// GET /api/auth/me
func (h *AuthHandler) GetProfile(c *gin.Context) {
	staffID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var staff models.Staff
	if err := h.db.First(&staff, staffID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Staff member not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"staff": gin.H{
			"id":          staff.ID,
			"email":       staff.Email,
			"first_name":  staff.FirstName,
			"last_name":   staff.LastName,
			"role":        staff.Role,
			"property_id": staff.PropertyID,
			"created_at":  staff.CreatedAt,
		},
	})
}

func (h *AuthHandler) generateToken(staff models.Staff) (string, error) {
	claims := Claims{
		StaffID:    staff.ID,
		Email:      staff.Email,
		Role:       staff.Role,
		PropertyID: staff.PropertyID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.jwtSecret))
}

// AuthMiddleware validates JWT tokens
func (h *AuthHandler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		// Parse and validate token
		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(h.jwtSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(*Claims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		// Verify staff is still active
		var staff models.Staff
		if err := h.db.Where("id = ? AND is_active = ?", claims.StaffID, true).First(&staff).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Staff member not found or inactive"})
			c.Abort()
			return
		}

		// Set user context
		c.Set("user_id", claims.StaffID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)
		c.Set("property_id", claims.PropertyID)

		c.Next()
	}
}

// RequireRole middleware checks if user has required role
func (h *AuthHandler) RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User role not found"})
			c.Abort()
			return
		}

		role := userRole.(string)
		for _, requiredRole := range roles {
			if role == requiredRole {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		c.Abort()
	}
}
