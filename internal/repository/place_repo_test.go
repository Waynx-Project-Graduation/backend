package repository

import (
	"testing"

	"github.com/kemit/trip-planner/internal/models"
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

	err = db.AutoMigrate(&models.Place{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}
	return db
}

func TestPlaceRepository_CreateAndFindByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPlaceRepository(db)

	place := &models.Place{
		Name: "Pyramids of Giza",
		City: "Cairo",
	}

	err := repo.Create(place)
	if err != nil {
		t.Fatalf("Expected no error creating place, got: %v", err)
	}

	if place.ID == 0 {
		t.Fatalf("Expected place ID to be set, got 0")
	}

	foundPlace, err := repo.FindByID(place.ID)
	if err != nil {
		t.Fatalf("Expected no error finding place, got: %v", err)
	}

	if foundPlace.Name != "Pyramids of Giza" {
		t.Errorf("Expected place name 'Pyramids of Giza', got '%s'", foundPlace.Name)
	}
}

func TestPlaceRepository_UpdateThumbnail(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPlaceRepository(db)

	place := &models.Place{
		Name: "Egyptian Museum",
		City: "Cairo",
	}
	repo.Create(place)

	err := repo.UpdateThumbnail(place.ID, "https://example.com/image.jpg")
	if err != nil {
		t.Fatalf("Expected no error updating thumbnail, got: %v", err)
	}

	foundPlace, _ := repo.FindByID(place.ID)
	if foundPlace.ThumbnailURL != "https://example.com/image.jpg" {
		t.Errorf("Expected thumbnail URL to be updated")
	}
}

func TestPlaceRepository_ListWithFilters(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPlaceRepository(db)

	repo.Create(&models.Place{Name: "Place A", City: "Cairo", Category: "history", Rating: 4.5})
	repo.Create(&models.Place{Name: "Place B", City: "Luxor", Category: "history", Rating: 4.8})
	repo.Create(&models.Place{Name: "Place C", City: "Cairo", Category: "food", Rating: 4.0})

	// Test by City
	places, total, err := repo.ListWithFilters(PlaceFilter{
		Cities:  []string{"Cairo"},
		Page:    1,
		PerPage: 10,
	})
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if total != 2 || len(places) != 2 {
		t.Errorf("Expected 2 places in Cairo, got %d", len(places))
	}

	// Test by Category
	places, total, err = repo.ListWithFilters(PlaceFilter{
		Category: "history",
		Page:     1,
		PerPage:  10,
	})
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if total != 2 || len(places) != 2 {
		t.Errorf("Expected 2 places in history category, got %d", len(places))
	}
}

func TestPlaceRepository_Search(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPlaceRepository(db)

	repo.Create(&models.Place{Name: "Luxor Temple", City: "Luxor", Category: "history", Rating: 4.7})
	repo.Create(&models.Place{Name: "Karnak Temple", City: "Luxor", Category: "history", Rating: 4.8})
	repo.Create(&models.Place{Name: "Egyptian Museum", City: "Cairo", Category: "history", Rating: 4.5})

	places, err := repo.Search("temple", 10)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if len(places) != 2 {
		t.Errorf("Expected 2 search results, got %d", len(places))
	}
}
