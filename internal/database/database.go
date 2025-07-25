package database

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"hudini-breakfast-module/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DatabaseConfig struct {
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

func Initialize(databaseURL string) (*gorm.DB, error) {
	return InitializeWithConfig(databaseURL, DatabaseConfig{
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
	})
}

func InitializeWithConfig(databaseURL string, config DatabaseConfig) (*gorm.DB, error) {
	// Configure GORM logger
	gormLogger := logger.Default.LogMode(logger.Info)

	// Determine database type and create appropriate dialect
	var dialector gorm.Dialector
	if strings.HasPrefix(databaseURL, "postgres://") || strings.HasPrefix(databaseURL, "postgresql://") {
		dialector = postgres.Open(databaseURL)
	} else if strings.HasPrefix(databaseURL, "sqlite://") {
		// Remove sqlite:// prefix for SQLite
		sqliteFile := strings.TrimPrefix(databaseURL, "sqlite://")
		dialector = sqlite.Open(sqliteFile)
	} else {
		// Default to SQLite for backward compatibility (assume it's a file path)
		dialector = sqlite.Open(databaseURL)
	}

	// Open database connection
	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: gormLogger,
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		PrepareStmt: true, // Prepare statements for better performance
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying SQL database
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying SQL database: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)

	// Test the connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Auto-migrate the schema
	err = db.AutoMigrate(
		&models.Property{},
		&models.Room{},
		&models.Guest{},
		&models.Staff{},
		&models.DailyBreakfastConsumption{},
		&models.OHIPTransaction{},
		&models.GuestPreference{},
		&models.Outlet{},
		&models.StaffComment{},
		&models.AuditLog{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate database schema: %w", err)
	}

	// Create performance indexes
	if err := CreateIndexes(db); err != nil {
		// Log error but don't fail initialization
		fmt.Printf("Warning: Failed to create some indexes: %v\n", err)
	}

	return db, nil
}

func Close(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying SQL database: %w", err)
	}
	return sqlDB.Close()
}

func GetStats(db *gorm.DB) (sql.DBStats, error) {
	sqlDB, err := db.DB()
	if err != nil {
		return sql.DBStats{}, fmt.Errorf("failed to get underlying SQL database: %w", err)
	}
	return sqlDB.Stats(), nil
}
