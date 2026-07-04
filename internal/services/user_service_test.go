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
	newName := "New Name"
	newCity := "Alexandria"
	newAvatar := "http://example.com/avatar.jpg"
	profileInput := UpdateProfileInput{
		FullName:  &newName,
		City:      &newCity,
		AvatarURL: &newAvatar,
	}
	updatedUser, err := userService.UpdateProfile(user.ID, profileInput)
	if err != nil {
		t.Fatalf("Expected no error updating profile, got: %v", err)
	}
	if updatedUser.FullName != "New Name" || updatedUser.City != "Alexandria" {
		t.Errorf("Profile not updated correctly")
	}

	// Update Preferences
	interests := []string{"history", "food"}
	companion := "family"
	prefInput := UpdatePreferencesInput{
		Interests:       &interests,
		TravelCompanion: &companion,
	}
	updatedUser, err = userService.UpdatePreferences(user.ID, prefInput)
	if err != nil {
		t.Fatalf("Expected no error updating preferences, got: %v", err)
	}
	if len(updatedUser.Preferences.Interests) != 2 || updatedUser.Preferences.TravelCompanion != "family" {
		t.Errorf("Preferences not updated correctly")
	}

	// Partial update should preserve existing preferences (bug #4 regression).
	budget := "medium"
	partial := UpdatePreferencesInput{Budget: &budget}
	updatedUser, err = userService.UpdatePreferences(user.ID, partial)
	if err != nil {
		t.Fatalf("Expected no error on partial preferences update, got: %v", err)
	}
	if updatedUser.Preferences.Budget != "medium" || len(updatedUser.Preferences.Interests) != 2 {
		t.Errorf("Partial preference update should preserve interests; got %+v", updatedUser.Preferences)
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
