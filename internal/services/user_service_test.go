package services

import (
	"testing"
	"github.com/kemit/trip-planner/internal/repository"
	"github.com/kemit/trip-planner/internal/models"
)

func TestUserService_UpdateProfileAndPreferences(t *testing.T) {
	db := setupTestDB(t)
	userRepo := repository.NewUserRepository(db)
	savedPlaceRepo := repository.NewSavedPlaceRepository(db)
	chatRepo := repository.NewChatRepository(db)
	userService := NewUserService(userRepo, savedPlaceRepo, chatRepo)

	// Create user
	user := &models.User{
		FullName: "Original Name",
		Email:    "user@example.com",
	}
	userRepo.Create(user)

	// Update Profile
	profileInput := UpdateProfileInput{
		FullName:  "New Name",
		City:      "Alexandria",
		AvatarURL: "http://example.com/avatar.jpg",
	}
	updatedUser, err := userService.UpdateProfile(user.ID, profileInput)
	if err != nil {
		t.Fatalf("Expected no error updating profile, got: %v", err)
	}
	if updatedUser.FullName != "New Name" || updatedUser.City != "Alexandria" {
		t.Errorf("Profile not updated correctly")
	}

	// Update Preferences
	prefInput := UpdatePreferencesInput{
		Interests:       []string{"history", "food"},
		TravelCompanion: "family",
	}
	updatedUser, err = userService.UpdatePreferences(user.ID, prefInput)
	if err != nil {
		t.Fatalf("Expected no error updating preferences, got: %v", err)
	}
	if len(updatedUser.Preferences.Interests) != 2 || updatedUser.Preferences.TravelCompanion != "family" {
		t.Errorf("Preferences not updated correctly")
	}
}

func TestUserService_GetStats(t *testing.T) {
	db := setupTestDB(t)
	userRepo := repository.NewUserRepository(db)
	savedPlaceRepo := repository.NewSavedPlaceRepository(db)
	chatRepo := repository.NewChatRepository(db)
	userService := NewUserService(userRepo, savedPlaceRepo, chatRepo)

	user := &models.User{FullName: "Stat User", Email: "stats@example.com", ExplorerPoints: 100}
	userRepo.Create(user)

	stats, err := userService.GetStats(user.ID)
	if err != nil {
		t.Fatalf("Expected no error getting stats, got: %v", err)
	}

	if stats.ExplorerPoints != 100 {
		t.Errorf("Expected 100 points, got %d", stats.ExplorerPoints)
	}
	if stats.DestinationsVisited != 0 || stats.SavedPlacesCount != 0 {
		t.Errorf("Expected 0 for empty repos")
	}
}
