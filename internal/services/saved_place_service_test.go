package services

import (
	"testing"
	"github.com/kemit/trip-planner/internal/repository"
	"github.com/kemit/trip-planner/internal/models"
)

func TestSavedPlaceService_SaveAndUnsave(t *testing.T) {
	db := setupTestDB(t)
	savedPlaceRepo := repository.NewSavedPlaceRepository(db)
	placeRepo := repository.NewPlaceRepository(db)
	userRepo := repository.NewUserRepository(db)
	svc := NewSavedPlaceService(savedPlaceRepo, placeRepo)

	user := &models.User{FullName: "Saver", Email: "saver@example.com"}
	userRepo.Create(user)

	place := &models.Place{Name: "To Save", City: "Cairo"}
	placeRepo.Create(place)

	// Save
	err := svc.SavePlace(user.ID, place.ID)
	if err != nil {
		t.Fatalf("Expected no error saving place, got: %v", err)
	}

	// Verify saved
	isSaved, _ := svc.IsSaved(user.ID, place.ID)
	if !isSaved {
		t.Errorf("Expected place to be saved")
	}

	// Save duplicate
	err = svc.SavePlace(user.ID, place.ID)
	if err == nil {
		t.Errorf("Expected error saving duplicate place")
	}

	// Unsave
	err = svc.UnsavePlace(user.ID, place.ID)
	if err != nil {
		t.Fatalf("Expected no error unsaving place, got: %v", err)
	}

	// Verify unsaved
	isSaved, _ = svc.IsSaved(user.ID, place.ID)
	if isSaved {
		t.Errorf("Expected place to not be saved")
	}

	// Unsave non-existent
	err = svc.UnsavePlace(user.ID, place.ID)
	if err == nil {
		t.Errorf("Expected error unsaving not saved place")
	}
}

func TestSavedPlaceService_ListSavedPlaces(t *testing.T) {
	db := setupTestDB(t)
	savedPlaceRepo := repository.NewSavedPlaceRepository(db)
	placeRepo := repository.NewPlaceRepository(db)
	userRepo := repository.NewUserRepository(db)
	svc := NewSavedPlaceService(savedPlaceRepo, placeRepo)

	user := &models.User{FullName: "List Saver", Email: "list@example.com"}
	userRepo.Create(user)

	place1 := &models.Place{Name: "Place 1", City: "Cairo"}
	place2 := &models.Place{Name: "Place 2", City: "Luxor"}
	placeRepo.Create(place1)
	placeRepo.Create(place2)

	svc.SavePlace(user.ID, place1.ID)
	svc.SavePlace(user.ID, place2.ID)

	places, total, err := svc.ListSavedPlaces(user.ID, 1, 10)
	if err != nil {
		t.Fatalf("Expected no error listing saved places, got: %v", err)
	}

	if total != 2 || len(places) != 2 {
		t.Errorf("Expected 2 saved places, got %d", len(places))
	}
}
