package database

import (
	"log"
	"os"

	"github.com/kemit/trip-planner/internal/config"
	"github.com/kemit/trip-planner/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"github.com/glebarez/sqlite"
)

// Connect opens a SQLite database at the path specified in config.
func Connect(cfg *config.Config) *gorm.DB {
	// Use verbose SQL logging only in debug mode
	logLevel := logger.Warn
	if cfg.GinMode == "debug" || os.Getenv("GIN_MODE") == "debug" || os.Getenv("GIN_MODE") == "" {
		logLevel = logger.Info
	}

	db, err := gorm.Open(sqlite.Open(cfg.DBPath), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Enable WAL mode and foreign keys for better performance & integrity
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get underlying DB: %v", err)
	}
	sqlDB.Exec("PRAGMA journal_mode=WAL")
	sqlDB.Exec("PRAGMA foreign_keys=ON")

	// Limit concurrent writes to prevent "database is locked" crashes
	sqlDB.SetMaxOpenConns(1)

	log.Println("Database connected successfully (SQLite)")
	return db
}

func AutoMigrate(db *gorm.DB) {
	err := db.AutoMigrate(
		&models.User{},
		&models.Place{},
		&models.Trip{},
		&models.TripDestination{},
		&models.TripDay{},
		&models.TripActivity{},
		&models.Notification{},
		&models.ChatSession{},
		&models.ChatMessage{},
		&models.SavedPlace{},
		&models.PlaceReview{},
		&models.TripExpense{},
		&models.TripMember{},
	)
	if err != nil {
		log.Fatalf("Failed to auto-migrate: %v", err)
	}
	log.Println("Database migration completed")
}
