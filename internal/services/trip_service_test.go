package services

import (
	"testing"
	"time"

	"github.com/kemit/trip-planner/internal/models"
	"github.com/kemit/trip-planner/internal/repository"
)

func TestTripService_CreateTrip(t *testing.T) {
	db := setupTestDB(t)
	tripRepo := repository.NewTripRepository(db)
	placeRepo := repository.NewPlaceRepository(db)
	userRepo := repository.NewUserRepository(db)

	mockAI := &MockAIClient{
		MockResponse: &RecommendResponse{
			Plan: AIPlan{
				Destinations: []AIDestination{
					{
						City:     "Cairo",
						Days:     1,
						Category: "history",
						DailySchedule: []AIPlanDay{
							{
								DayNumber: 1,
								Activities: []AIPlanActivity{
									{
										Name:          "Pyramids",
										Category:      "history",
										DurationHours: 4,
										Rating:        5.0,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	tripService := NewTripService(tripRepo, placeRepo, mockAI)

	user := &models.User{FullName: "Trip User", Email: "trip@example.com"}
	userRepo.Create(user)

	input := CreateTripInput{
		StartDate:      time.Now().Format("2006-01-02"),
		EndDate:        time.Now().AddDate(0, 0, 1).Format("2006-01-02"),
		TravelersCount: 2,
		Preferences: models.Preferences{
			Interests:       []string{"history"},
			TravelCompanion: "friends",
			Budget:          "moderate",
		},
	}

	trip, err := tripService.CreateTrip(user.ID, input)
	if err != nil {
		t.Fatalf("Expected no error creating trip, got: %v", err)
	}

	if trip.Status != "planned" {
		t.Errorf("Expected trip status to be 'planned', got %s", trip.Status)
	}

	if trip.Title != "History Trip to Egypt" {
		t.Errorf("Expected trip title 'History Trip to Egypt', got %s", trip.Title)
	}
}

func TestTripService_GetAndUpdateTrip(t *testing.T) {
	db := setupTestDB(t)
	tripRepo := repository.NewTripRepository(db)
	placeRepo := repository.NewPlaceRepository(db)
	userRepo := repository.NewUserRepository(db)
	tripService := NewTripService(tripRepo, placeRepo, &MockAIClient{})

	user := &models.User{FullName: "Trip User", Email: "trip@example.com"}
	userRepo.Create(user)

	otherUser := &models.User{FullName: "Other User", Email: "other@example.com"}
	userRepo.Create(otherUser)

	trip := &models.Trip{
		UserID:         user.ID,
		Title:          "Old Title",
		Status:         "draft",
		TravelersCount: 1,
	}
	tripRepo.Create(trip)

	// Get Trip success
	fetched, err := tripService.GetTrip(trip.ID, user.ID)
	if err != nil {
		t.Fatalf("Expected no error fetching trip, got: %v", err)
	}
	if fetched.Title != "Old Title" {
		t.Errorf("Expected title 'Old Title', got %s", fetched.Title)
	}

	// Get Trip failure (access denied)
	_, err = tripService.GetTrip(trip.ID, otherUser.ID)
	if err == nil {
		t.Errorf("Expected access denied error fetching other user's trip")
	}

	// Update Trip
	updateInput := UpdateTripInput{
		Title:          "New Title",
		TravelersCount: 4,
	}
	updated, err := tripService.UpdateTrip(trip.ID, user.ID, updateInput)
	if err != nil {
		t.Fatalf("Expected no error updating trip, got: %v", err)
	}
	if updated.Title != "New Title" || updated.TravelersCount != 4 {
		t.Errorf("Expected trip to be updated")
	}
}
