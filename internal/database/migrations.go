package database

import (
	"fmt"
	"hudini-breakfast-module/internal/logging"
	
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// CreateIndexes creates performance-optimized indexes for the database
func CreateIndexes(db *gorm.DB) error {
	logging.Info("Creating database indexes for performance optimization")
	
	indexes := []struct {
		Table string
		Name  string
		SQL   string
	}{
		// Guest indexes
		{
			Table: "guests",
			Name:  "idx_guests_vip",
			SQL:   "CREATE INDEX IF NOT EXISTS idx_guests_vip ON guests(is_vip, property_id) WHERE is_vip = 1",
		},
		{
			Table: "guests",
			Name:  "idx_guests_upset",
			SQL:   "CREATE INDEX IF NOT EXISTS idx_guests_upset ON guests(is_upset, property_id) WHERE is_upset = 1",
		},
		{
			Table: "guests",
			Name:  "idx_guests_active_dates",
			SQL:   "CREATE INDEX IF NOT EXISTS idx_guests_active_dates ON guests(property_id, check_in_date, check_out_date) WHERE is_active = 1",
		},
		{
			Table: "guests",
			Name:  "idx_guests_room_property",
			SQL:   "CREATE INDEX IF NOT EXISTS idx_guests_room_property ON guests(room_number, property_id)",
		},
		
		// Daily consumption indexes
		{
			Table: "daily_breakfast_consumptions",
			Name:  "idx_consumption_date_property",
			SQL:   "CREATE INDEX IF NOT EXISTS idx_consumption_date_property ON daily_breakfast_consumptions(consumption_date, property_id)",
		},
		{
			Table: "daily_breakfast_consumptions",
			Name:  "idx_consumption_room_date",
			SQL:   "CREATE INDEX IF NOT EXISTS idx_consumption_room_date ON daily_breakfast_consumptions(room_number, consumption_date)",
		},
		{
			Table: "daily_breakfast_consumptions",
			Name:  "idx_consumption_status",
			SQL:   "CREATE INDEX IF NOT EXISTS idx_consumption_status ON daily_breakfast_consumptions(status, consumption_date) WHERE deleted_at IS NULL",
		},
		
		// Room indexes
		{
			Table: "rooms",
			Name:  "idx_rooms_property_status",
			SQL:   "CREATE INDEX IF NOT EXISTS idx_rooms_property_status ON rooms(property_id, status)",
		},
		{
			Table: "rooms",
			Name:  "idx_rooms_floor",
			SQL:   "CREATE INDEX IF NOT EXISTS idx_rooms_floor ON rooms(property_id, floor)",
		},
		
		// Guest preferences indexes
		{
			Table: "guest_preferences",
			Name:  "idx_preferences_guest",
			SQL:   "CREATE INDEX IF NOT EXISTS idx_preferences_guest ON guest_preferences(guest_id) WHERE deleted_at IS NULL",
		},
		
		// Staff comments indexes
		{
			Table: "staff_comments",
			Name:  "idx_comments_guest",
			SQL:   "CREATE INDEX IF NOT EXISTS idx_comments_guest ON staff_comments(guest_id) WHERE guest_id IS NOT NULL",
		},
		{
			Table: "staff_comments",
			Name:  "idx_comments_unresolved",
			SQL:   "CREATE INDEX IF NOT EXISTS idx_comments_unresolved ON staff_comments(category, is_resolved) WHERE is_resolved = 0",
		},
		
		// Outlet indexes
		{
			Table: "outlets",
			Name:  "idx_outlets_property_active",
			SQL:   "CREATE INDEX IF NOT EXISTS idx_outlets_property_active ON outlets(property_id, is_active) WHERE is_active = 1",
		},
		
		// Audit log indexes
		{
			Table: "audit_logs",
			Name:  "idx_audit_user_action",
			SQL:   "CREATE INDEX IF NOT EXISTS idx_audit_user_action ON audit_logs(user_id, action, created_at DESC)",
		},
		{
			Table: "audit_logs",
			Name:  "idx_audit_resource",
			SQL:   "CREATE INDEX IF NOT EXISTS idx_audit_resource ON audit_logs(resource, resource_id, created_at DESC)",
		},
		{
			Table: "audit_logs",
			Name:  "idx_audit_date_range",
			SQL:   "CREATE INDEX IF NOT EXISTS idx_audit_date_range ON audit_logs(created_at DESC) WHERE status = 'failed'",
		},
		{
			Table: "audit_logs",
			Name:  "idx_audit_ip_tracking",
			SQL:   "CREATE INDEX IF NOT EXISTS idx_audit_ip_tracking ON audit_logs(ip_address, created_at DESC)",
		},
	}
	
	// Execute each index creation
	for _, idx := range indexes {
		logging.WithFields(logrus.Fields{
			"table": idx.Table,
			"index": idx.Name,
		}).Debug("Creating index")
		
		if err := db.Exec(idx.SQL).Error; err != nil {
			logging.WithFields(logrus.Fields{
				"table": idx.Table,
				"index": idx.Name,
				"error": err.Error(),
			}).Error("Failed to create index")
			// Continue with other indexes even if one fails
		} else {
			logging.WithFields(logrus.Fields{
				"table": idx.Table,
				"index": idx.Name,
			}).Info("Index created successfully")
		}
	}
	
	// Analyze tables for query optimization
	tables := []string{"guests", "rooms", "daily_breakfast_consumptions", "guest_preferences", "staff_comments", "outlets", "audit_logs"}
	for _, table := range tables {
		if err := db.Exec(fmt.Sprintf("ANALYZE %s", table)).Error; err != nil {
			logging.WithFields(logrus.Fields{
				"table": table,
				"error": err.Error(),
			}).Warn("Failed to analyze table")
		}
	}
	
	logging.Info("Database indexing completed")
	return nil
}

// DropIndexes removes all custom indexes (for rollback)
func DropIndexes(db *gorm.DB) error {
	indexes := []string{
		"idx_guests_vip",
		"idx_guests_upset",
		"idx_guests_active_dates",
		"idx_guests_room_property",
		"idx_consumption_date_property",
		"idx_consumption_room_date",
		"idx_consumption_status",
		"idx_rooms_property_status",
		"idx_rooms_floor",
		"idx_preferences_guest",
		"idx_comments_guest",
		"idx_comments_unresolved",
		"idx_outlets_property_active",
		"idx_audit_user_action",
		"idx_audit_resource",
		"idx_audit_date_range",
		"idx_audit_ip_tracking",
	}
	
	for _, idx := range indexes {
		if err := db.Exec(fmt.Sprintf("DROP INDEX IF EXISTS %s", idx)).Error; err != nil {
			logging.WithFields(logrus.Fields{
				"index": idx,
				"error": err.Error(),
			}).Warn("Failed to drop index")
		}
	}
	
	return nil
}