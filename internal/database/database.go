package database

import (
	"hudini-breakfast-module/internal/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func Initialize(databaseURL string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("breakfast.db"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Auto-migrate the schema
	err = db.AutoMigrate(
		&models.Property{},
		&models.Room{},
		&models.Guest{},
		&models.Staff{},
		&models.DailyBreakfastConsumption{},
		&models.OHIPTransaction{},
	)
	if err != nil {
		return nil, err
	}

	return db, nil
}
