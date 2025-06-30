package config

import (
	"os"
	"strconv"
)

type Config struct {
	DatabaseURL     string
	JWTSecret       string
	Port            string
	OHIP            OHIPConfig
	PMSIntegration  PMSConfig
}

type OHIPConfig struct {
	BaseURL      string
	ClientID     string
	ClientSecret string
	Environment  string // sandbox, production
	Version      string
	Timeout      int
}

type PMSConfig struct {
	BaseURL    string
	APIKey     string
	PropertyID string
	Timeout    int
	Environment string // sandbox, production
}

func Load() *Config {
	ohipTimeout, _ := strconv.Atoi(getEnv("OHIP_TIMEOUT", "30"))
	pmsTimeout, _ := strconv.Atoi(getEnv("PMS_TIMEOUT", "30"))
	
	return &Config{
		DatabaseURL: getEnv("DATABASE_URL", "oracle://user:password@localhost:1521/XE"),
		JWTSecret:   getEnv("JWT_SECRET", "your-secret-key"),
		Port:        getEnv("PORT", "8080"),
		OHIP: OHIPConfig{
			BaseURL:      getEnv("OHIP_BASE_URL", "https://api.oraclehospitality.com"),
			ClientID:     getEnv("OHIP_CLIENT_ID", ""),
			ClientSecret: getEnv("OHIP_CLIENT_SECRET", ""),
			Environment:  getEnv("OHIP_ENVIRONMENT", "sandbox"),
			Version:      getEnv("OHIP_VERSION", "v1"),
			Timeout:      ohipTimeout,
		},
		PMSIntegration: PMSConfig{
			BaseURL:     getEnv("PMS_BASE_URL", "https://api.pms-provider.com"),
			APIKey:      getEnv("PMS_API_KEY", ""),
			PropertyID:  getEnv("PMS_PROPERTY_ID", ""),
			Timeout:     pmsTimeout,
			Environment: getEnv("PMS_ENVIRONMENT", "sandbox"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
