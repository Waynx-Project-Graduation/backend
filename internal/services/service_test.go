package services

import (
	"testing"
	"time"

	"github.com/kemit/trip-planner/internal/models"
	"github.com/kemit/trip-planner/internal/utils"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	err = db.AutoMigrate(
		&models.User{},
		&models.Place{},
		&models.Trip{},
		&models.TripDestination{},
		&models.TripDay{},
		&models.TripActivity{},
		&models.SavedPlace{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}
	return db
}

func setupJWTManager() *utils.JWTManager {
	return utils.NewJWTManager("test_secret", 15*time.Minute, 24*time.Hour)
}

// MockAIClient implements TripAIProvider for TripService tests
type MockAIClient struct {
	MockResponse *RecommendResponse
	MockError    error
}

func (m *MockAIClient) GetRecommendation(req RecommendRequest) (*RecommendResponse, error) {
	return m.MockResponse, m.MockError
}
