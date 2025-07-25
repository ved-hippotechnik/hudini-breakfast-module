package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port           string
	DatabaseURL    string
	RedisURL       string
	JWTSecret      string
	GinMode        string
	OHIP           OHIPConfig
	PMSIntegration PMSConfig
	PMSProviders   PMSProvidersConfig
	Security       SecurityConfig
	Database       DatabaseConfig
	Logging        LoggingConfig
}

type OHIPConfig struct {
	BaseURL      string
	ClientID     string
	ClientSecret string
	Username     string
	Password     string
	Environment  string
	Version      string
	Timeout      int
}

type PMSConfig struct {
	BaseURL     string
	APIKey      string
	PropertyID  string
	Timeout     int
	Environment string
}

type PMSProvidersConfig struct {
	DefaultProvider string                    `json:"default_provider"`
	Providers       map[string]PMSProviderConfig `json:"providers"`
}

type PMSProviderConfig struct {
	Name         string            `json:"name"`
	Type         string            `json:"type"` // oracle_ohip, opera, fidelio, etc.
	BaseURL      string            `json:"base_url"`
	Username     string            `json:"username"`
	Password     string            `json:"password"`
	APIKey       string            `json:"api_key"`
	ClientID     string            `json:"client_id"`
	ClientSecret string            `json:"client_secret"`
	PropertyID   string            `json:"property_id"`
	Timeout      int               `json:"timeout"`
	Environment  string            `json:"environment"`
	Enabled      bool              `json:"enabled"`
	Additional   map[string]string `json:"additional"`
}

type SecurityConfig struct {
	JWTExpiration     time.Duration
	JWTRefreshExpiry  time.Duration
	RateLimitRequests int
	RateLimitWindow   time.Duration
	TLSEnabled        bool
	CertFile          string
	KeyFile           string
	MinPasswordLength int
	RequireStrongAuth bool
}

type DatabaseConfig struct {
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	BackupEnabled   bool
	BackupInterval  time.Duration
	BackupPath      string
}

type LoggingConfig struct {
	Level      string
	Format     string // json, text
	Output     string // stdout, file
	FilePath   string
	MaxSize    int // MB
	MaxBackups int
	MaxAge     int // days
}

func Load() *Config {
	if err := ValidateRequiredEnvVars(); err != nil {
		log.Fatalf("Configuration validation failed: %v", err)
	}

	jwtExpiration, _ := time.ParseDuration(getEnvOrDefault("JWT_EXPIRATION", "24h"))
	jwtRefreshExpiry, _ := time.ParseDuration(getEnvOrDefault("JWT_REFRESH_EXPIRY", "168h")) // 7 days
	rateLimitRequests, _ := strconv.Atoi(getEnvOrDefault("RATE_LIMIT_REQUESTS", "100"))
	rateLimitWindow, _ := time.ParseDuration(getEnvOrDefault("RATE_LIMIT_WINDOW", "1m"))
	
	maxOpenConns, _ := strconv.Atoi(getEnvOrDefault("DB_MAX_OPEN_CONNS", "25"))
	maxIdleConns, _ := strconv.Atoi(getEnvOrDefault("DB_MAX_IDLE_CONNS", "5"))
	connMaxLifetime, _ := time.ParseDuration(getEnvOrDefault("DB_CONN_MAX_LIFETIME", "5m"))
	backupInterval, _ := time.ParseDuration(getEnvOrDefault("DB_BACKUP_INTERVAL", "24h"))

	ohipTimeout, _ := strconv.Atoi(getEnvOrDefault("OHIP_TIMEOUT", "30"))
	pmsTimeout, _ := strconv.Atoi(getEnvOrDefault("PMS_TIMEOUT", "30"))
	minPasswordLength, _ := strconv.Atoi(getEnvOrDefault("MIN_PASSWORD_LENGTH", "8"))

	return &Config{
		Port:        getEnvOrDefault("PORT", "3001"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		RedisURL:    getEnvOrDefault("REDIS_URL", ""),
		JWTSecret:   os.Getenv("JWT_SECRET"),
		GinMode:     getEnvOrDefault("GIN_MODE", "debug"),
		OHIP: OHIPConfig{
			BaseURL:      getEnvOrDefault("OHIP_BASE_URL", ""),
			ClientID:     getEnvOrDefault("OHIP_CLIENT_ID", ""),
			ClientSecret: getEnvOrDefault("OHIP_CLIENT_SECRET", ""),
			Username:     getEnvOrDefault("OHIP_USERNAME", ""),
			Password:     getEnvOrDefault("OHIP_PASSWORD", ""),
			Environment:  getEnvOrDefault("OHIP_ENVIRONMENT", "sandbox"),
			Version:      getEnvOrDefault("OHIP_VERSION", "v1"),
			Timeout:      ohipTimeout,
		},
		PMSIntegration: PMSConfig{
			BaseURL:     getEnvOrDefault("PMS_BASE_URL", ""),
			APIKey:      getEnvOrDefault("PMS_API_KEY", ""),
			PropertyID:  getEnvOrDefault("PMS_PROPERTY_ID", ""),
			Timeout:     pmsTimeout,
			Environment: getEnvOrDefault("PMS_ENVIRONMENT", "sandbox"),
		},
		PMSProviders: PMSProvidersConfig{
			DefaultProvider: getEnvOrDefault("PMS_DEFAULT_PROVIDER", "oracle_ohip"),
			Providers: map[string]PMSProviderConfig{
				"oracle_ohip": {
					Name:         "Oracle OHIP",
					Type:         "oracle_ohip",
					BaseURL:      getEnvOrDefault("ORACLE_OHIP_BASE_URL", ""),
					Username:     getEnvOrDefault("ORACLE_OHIP_USERNAME", ""),
					Password:     getEnvOrDefault("ORACLE_OHIP_PASSWORD", ""),
					ClientID:     getEnvOrDefault("ORACLE_OHIP_CLIENT_ID", ""),
					ClientSecret: getEnvOrDefault("ORACLE_OHIP_CLIENT_SECRET", ""),
					PropertyID:   getEnvOrDefault("ORACLE_OHIP_PROPERTY_ID", ""),
					Timeout:      getEnvInt("ORACLE_OHIP_TIMEOUT", 30),
					Environment:  getEnvOrDefault("ORACLE_OHIP_ENVIRONMENT", "sandbox"),
					Enabled:      getEnvBool("ORACLE_OHIP_ENABLED", true),
					Additional:   make(map[string]string),
				},
				"opera": {
					Name:         "Oracle Opera",
					Type:         "opera",
					BaseURL:      getEnvOrDefault("OPERA_BASE_URL", ""),
					Username:     getEnvOrDefault("OPERA_USERNAME", ""),
					Password:     getEnvOrDefault("OPERA_PASSWORD", ""),
					ClientID:     getEnvOrDefault("OPERA_CLIENT_ID", ""),
					ClientSecret: getEnvOrDefault("OPERA_CLIENT_SECRET", ""),
					PropertyID:   getEnvOrDefault("OPERA_PROPERTY_ID", ""),
					Timeout:      getEnvInt("OPERA_TIMEOUT", 30),
					Environment:  getEnvOrDefault("OPERA_ENVIRONMENT", "sandbox"),
					Enabled:      getEnvBool("OPERA_ENABLED", false),
					Additional:   make(map[string]string),
				},
				"fidelio": {
					Name:         "Fidelio",
					Type:         "fidelio",
					BaseURL:      getEnvOrDefault("FIDELIO_BASE_URL", ""),
					Username:     getEnvOrDefault("FIDELIO_USERNAME", ""),
					Password:     getEnvOrDefault("FIDELIO_PASSWORD", ""),
					APIKey:       getEnvOrDefault("FIDELIO_API_KEY", ""),
					PropertyID:   getEnvOrDefault("FIDELIO_PROPERTY_ID", ""),
					Timeout:      getEnvInt("FIDELIO_TIMEOUT", 30),
					Environment:  getEnvOrDefault("FIDELIO_ENVIRONMENT", "sandbox"),
					Enabled:      getEnvBool("FIDELIO_ENABLED", false),
					Additional:   make(map[string]string),
				},
			},
		},
		Security: SecurityConfig{
			JWTExpiration:     jwtExpiration,
			JWTRefreshExpiry:  jwtRefreshExpiry,
			RateLimitRequests: rateLimitRequests,
			RateLimitWindow:   rateLimitWindow,
			TLSEnabled:        getEnvBool("TLS_ENABLED", false),
			CertFile:          getEnvOrDefault("TLS_CERT_FILE", ""),
			KeyFile:           getEnvOrDefault("TLS_KEY_FILE", ""),
			MinPasswordLength: minPasswordLength,
			RequireStrongAuth: getEnvBool("REQUIRE_STRONG_AUTH", true),
		},
		Database: DatabaseConfig{
			MaxOpenConns:    maxOpenConns,
			MaxIdleConns:    maxIdleConns,
			ConnMaxLifetime: connMaxLifetime,
			BackupEnabled:   getEnvBool("DB_BACKUP_ENABLED", false),
			BackupInterval:  backupInterval,
			BackupPath:      getEnvOrDefault("DB_BACKUP_PATH", "./backups"),
		},
		Logging: LoggingConfig{
			Level:      getEnvOrDefault("LOG_LEVEL", "info"),
			Format:     getEnvOrDefault("LOG_FORMAT", "json"),
			Output:     getEnvOrDefault("LOG_OUTPUT", "stdout"),
			FilePath:   getEnvOrDefault("LOG_FILE_PATH", "./logs/app.log"),
			MaxSize:    getEnvInt("LOG_MAX_SIZE", 100),
			MaxBackups: getEnvInt("LOG_MAX_BACKUPS", 3),
			MaxAge:     getEnvInt("LOG_MAX_AGE", 28),
		},
	}
}

func ValidateRequiredEnvVars() error {
	required := []string{
		"DATABASE_URL",
		"JWT_SECRET",
	}

	for _, env := range required {
		if os.Getenv(env) == "" {
			return fmt.Errorf("required environment variable %s is not set", env)
		}
	}

	// Validate JWT secret strength
	if len(os.Getenv("JWT_SECRET")) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters long for security")
	}

	return nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// Legacy function for backward compatibility
func getEnv(key, defaultValue string) string {
	return getEnvOrDefault(key, defaultValue)
}

func IsDevelopment() bool {
	return getEnvOrDefault("GIN_MODE", "debug") == "debug"
}

func IsProduction() bool {
	return getEnvOrDefault("GIN_MODE", "debug") == "release"
}
