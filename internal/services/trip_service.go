package services

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/kemit/trip-planner/internal/models"
	"github.com/kemit/trip-planner/internal/repository"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type TripAIProvider interface {
	GetRecommendation(req RecommendRequest) (*RecommendResponse, error)
}

type TripService struct {
	tripRepo  *repository.TripRepository
	placeRepo *repository.PlaceRepository
	aiClient  TripAIProvider
}

func NewTripService(tripRepo *repository.TripRepository, placeRepo *repository.PlaceRepository, aiClient TripAIProvider) *TripService {
	return &TripService{
		tripRepo:  tripRepo,
		placeRepo: placeRepo,
		aiClient:  aiClient,
	}
}

// ── Input / Output Types ────────────────────────────────────────────────────

// CreateTripInput — the frontend sends the same preferences JSON object
// as stored in the user profile. The AI uses it to pick cities.
type CreateTripInput struct {
	StartDate      string            `json:"start_date" binding:"required"` // YYYY-MM-DD
	EndDate        string            `json:"end_date" binding:"required"`   // YYYY-MM-DD
	TravelersCount int               `json:"travelers_count" binding:"required,gte=1"`
	Preferences    models.Preferences `json:"preferences" binding:"required"`
}

type UpdateTripInput struct {
	Title          string `json:"title"`
	StartDate      string `json:"start_date"`
	EndDate        string `json:"end_date"`
	TravelersCount int    `json:"travelers_count"`
	Status         string `json:"status"`
}

type UpdateActivityInput struct {
	ActivityName  string  `json:"activity_name"`
	Description   string  `json:"description"`
	StartTime     string  `json:"start_time"`
	EndTime       string  `json:"end_time"`
	DurationHours int     `json:"duration_hours"`
	EstimatedCost float64 `json:"estimated_cost"`
	ActivityType  string  `json:"activity_type"`
}

// ── Create Trip ─────────────────────────────────────────────────────────────

func (s *TripService) CreateTrip(userID uuid.UUID, input CreateTripInput) (*models.Trip, error) {
	startDate, err := time.Parse("2006-01-02", input.StartDate)
	if err != nil {
		return nil, errors.New("invalid start_date format, use YYYY-MM-DD")
	}

	endDate, err := time.Parse("2006-01-02", input.EndDate)
	if err != nil {
		return nil, errors.New("invalid end_date format, use YYYY-MM-DD")
	}

	if endDate.Before(startDate) {
		return nil, errors.New("end_date must be after start_date")
	}
	
	if input.TravelersCount < 1 {
		return nil, errors.New("travelers_count must be at least 1")
	}
	if len(input.Preferences.Interests) == 0 {
		return nil, errors.New("at least one interest must be provided")
	}
	if input.Preferences.TravelCompanion == "" {
		return nil, errors.New("travel_companion must be specified")
	}
	if input.Preferences.Budget == "" {
		return nil, errors.New("budget must be specified")
	}

	numDays := int(endDate.Sub(startDate).Hours()/24) + 1

	// Build a human-readable title from interests
	titleCaser := cases.Title(language.English)
	title := "Trip to Egypt"
	if len(input.Preferences.Interests) > 0 {
		title = fmt.Sprintf("%s Trip to Egypt", titleCaser.String(input.Preferences.Interests[0]))
	}
	if len(input.Preferences.Interests) > 1 {
		title = fmt.Sprintf("%s & %s Trip to Egypt",
			titleCaser.String(input.Preferences.Interests[0]),
			titleCaser.String(input.Preferences.Interests[1]))
	}

	// Create the trip parent record (status = draft)
	trip := &models.Trip{
		UserID:         userID,
		Title:          title,
		StartDate:      startDate,
		EndDate:        endDate,
		TravelersCount: input.TravelersCount,
		Preferences:    input.Preferences,
		Status:         "draft",
	}

	if err := s.tripRepo.Create(trip); err != nil {
		return nil, errors.New("failed to create trip")
	}

	// Call AI service for recommendations + plan
	aiReq := RecommendRequest{
		Interests:        input.Preferences.Interests,
		TravelCompanion:  input.Preferences.TravelCompanion,
		Budget:           input.Preferences.Budget,
		AgeGroup:         input.Preferences.AgeGroup,
		CrowdPreference:  input.Preferences.CrowdPreference,
		Season:           input.Preferences.Season,
		TripDurationDays: numDays,
	}

	aiResp, err := s.aiClient.GetRecommendation(aiReq)
	if err != nil {
		log.Printf("AI service error (trip %s kept as draft): %v", trip.ID, err)
		// Return trip as draft — AI is unavailable
		return s.tripRepo.FindByID(trip.ID)
	}

	// Save the AI-generated multi-city itinerary
	if err := s.saveItinerary(trip.ID, startDate, aiResp); err != nil {
		log.Printf("Failed to save itinerary for trip %s: %v", trip.ID, err)
		return s.tripRepo.FindByID(trip.ID)
	}

	// Update trip status to planned
	trip.Status = "planned"
	_ = s.tripRepo.Update(trip)

	return s.tripRepo.FindByID(trip.ID)
}

// saveItinerary persists the AI plan as Destinations → Days → Activities
func (s *TripService) saveItinerary(tripID uuid.UUID, startDate time.Time, aiResp *RecommendResponse) error {
	if len(aiResp.Plan.Destinations) == 0 {
		return errors.New("AI returned an empty plan")
	}

	tx := s.tripRepo.DB().Begin()
	if tx.Error != nil {
		log.Printf("Transaction error: %v", tx.Error)
		return errors.New("failed to begin transaction")
	}
	defer tx.Rollback()

	txRepo := s.tripRepo.WithTx(tx)

	globalDayOffset := 0

	for destIdx, aiDest := range aiResp.Plan.Destinations {
		// Create the destination record — pre-generate ID so we can reference it
		destID := uuid.New()
		dest := models.TripDestination{
			ID:            destID,
			TripID:        tripID,
			City:          aiDest.City,
			DaysAllocated: aiDest.Days,
			Category:      aiDest.Category,
			OrderInTrip:   destIdx + 1,
		}

		if err := txRepo.CreateDestinations([]models.TripDestination{dest}); err != nil {
			log.Printf("Destination save error: %v", err)
			return fmt.Errorf("failed to save destination %s", aiDest.City)
		}

		// Create days and activities for this destination
		for _, aiDay := range aiDest.DailySchedule {
			dayDate := startDate.AddDate(0, 0, globalDayOffset)

			tripDayID := uuid.New()
			tripDay := models.TripDay{
				ID:                tripDayID,
				TripID:            tripID,
				TripDestinationID: dest.ID,
				DayNumber:         aiDay.DayNumber,
				Date:              dayDate,
				HoursUsed:         aiDay.HoursUsed,
			}

			if err := txRepo.CreateTripDays([]models.TripDay{tripDay}); err != nil {
				log.Printf("Trip day save error: %v", err)
				return fmt.Errorf("failed to save day %d", aiDay.DayNumber)
			}

			var activities []models.TripActivity
			for i, aiAct := range aiDay.Activities {
				// Validate that the AI-generated place ID actually exists in our DB
				var validPlaceID *uint
				if aiAct.PlaceID > 0 {
					if _, err := s.placeRepo.FindByID(aiAct.PlaceID); err == nil {
						pid := aiAct.PlaceID
						validPlaceID = &pid
					} else {
						log.Printf("AI returned non-existent place_id %d for activity %q, skipping FK link", aiAct.PlaceID, aiAct.Name)
					}
				}
				activity := models.TripActivity{
					TripDayID:     tripDay.ID,
					PlaceID:       validPlaceID,
					ActivityName:  aiAct.Name,
					Category:      aiAct.Category,
					DurationHours: aiAct.DurationHours,
					Rating:        aiAct.Rating,
					OrderInDay:    i + 1,
					ActivityType:  aiAct.Category,
				}
				activities = append(activities, activity)
			}

			if len(activities) > 0 {
				if err := txRepo.CreateActivities(activities); err != nil {
					log.Printf("Activities save error: %v", err)
					return fmt.Errorf("failed to save activities for day %d", aiDay.DayNumber)
				}
			}

			globalDayOffset++
		}
	}

	return tx.Commit().Error
}

// ── Read ────────────────────────────────────────────────────────────────────

func (s *TripService) GetTrip(tripID uuid.UUID, userID uuid.UUID) (*models.Trip, error) {
	trip, err := s.tripRepo.FindByID(tripID)
	if err != nil {
		return nil, errors.New("trip not found")
	}
	if trip.UserID != userID {
		return nil, errors.New("access denied")
	}
	return trip, nil
}

func (s *TripService) ListTrips(userID uuid.UUID, page, perPage int) ([]models.Trip, int64, error) {
	return s.tripRepo.ListByUserID(userID, page, perPage)
}

// ── Update ──────────────────────────────────────────────────────────────────

func (s *TripService) UpdateTrip(tripID, userID uuid.UUID, input UpdateTripInput) (*models.Trip, error) {
	trip, err := s.tripRepo.FindByID(tripID)
	if err != nil {
		return nil, errors.New("trip not found")
	}
	if trip.UserID != userID {
		return nil, errors.New("access denied")
	}

	if input.Title != "" {
		trip.Title = input.Title
	}
	if input.TravelersCount > 0 {
		trip.TravelersCount = input.TravelersCount
	}
	if input.Status != "" {
		trip.Status = input.Status
	}
	if input.StartDate != "" {
		if t, err := time.Parse("2006-01-02", input.StartDate); err == nil {
			trip.StartDate = t
		}
	}
	if input.EndDate != "" {
		if t, err := time.Parse("2006-01-02", input.EndDate); err == nil {
			trip.EndDate = t
		}
	}

	if err := s.tripRepo.Update(trip); err != nil {
		return nil, errors.New("failed to update trip")
	}
	return s.tripRepo.FindByID(tripID)
}

// ── Delete ──────────────────────────────────────────────────────────────────

func (s *TripService) DeleteTrip(tripID, userID uuid.UUID) error {
	trip, err := s.tripRepo.FindByID(tripID)
	if err != nil {
		return errors.New("trip not found")
	}
	if trip.UserID != userID {
		return errors.New("access denied")
	}

	// Delete destinations (cascades to days and activities)
	_ = s.tripRepo.DeleteDestinations(tripID)
	return s.tripRepo.Delete(tripID)
}

// ── Regenerate ──────────────────────────────────────────────────────────────

func (s *TripService) RegenerateItinerary(tripID, userID uuid.UUID) (*models.Trip, error) {
	trip, err := s.tripRepo.FindByID(tripID)
	if err != nil {
		return nil, errors.New("trip not found")
	}
	if trip.UserID != userID {
		return nil, errors.New("access denied")
	}

	// Delete old itinerary (destinations → days → activities)
	_ = s.tripRepo.DeleteDestinations(tripID)

	numDays := int(trip.EndDate.Sub(trip.StartDate).Hours()/24) + 1

	aiReq := RecommendRequest{
		Interests:        trip.Preferences.Interests,
		TravelCompanion:  trip.Preferences.TravelCompanion,
		Budget:           trip.Preferences.Budget,
		AgeGroup:         trip.Preferences.AgeGroup,
		CrowdPreference:  trip.Preferences.CrowdPreference,
		Season:           trip.Preferences.Season,
		TripDurationDays: numDays,
	}

	aiResp, err := s.aiClient.GetRecommendation(aiReq)
	if err != nil {
		log.Printf("AI regeneration failed: %v", err)
		return nil, errors.New("AI service unavailable")
	}

	if err := s.saveItinerary(tripID, trip.StartDate, aiResp); err != nil {
		return nil, errors.New("failed to save new itinerary")
	}

	trip.Status = "planned"
	_ = s.tripRepo.Update(trip)

	return s.tripRepo.FindByID(tripID)
}

// ── Activity Editing ────────────────────────────────────────────────────────

func (s *TripService) UpdateActivity(activityID, userID uuid.UUID, input UpdateActivityInput) (*models.TripActivity, error) {
	activity, err := s.tripRepo.FindActivity(activityID)
	if err != nil {
		return nil, errors.New("activity not found")
	}

	// Verify ownership: activity → day → destination → trip → user
	if err := s.verifyActivityOwnership(activityID, userID); err != nil {
		return nil, err
	}

	if input.ActivityName != "" {
		activity.ActivityName = input.ActivityName
	}
	if input.Description != "" {
		activity.Description = input.Description
	}
	if input.StartTime != "" {
		activity.StartTime = input.StartTime
	}
	if input.EndTime != "" {
		activity.EndTime = input.EndTime
	}
	if input.DurationHours > 0 {
		activity.DurationHours = input.DurationHours
	}
	if input.EstimatedCost > 0 {
		activity.EstimatedCost = input.EstimatedCost
	}
	if input.ActivityType != "" {
		activity.ActivityType = input.ActivityType
	}

	if err := s.tripRepo.UpdateActivity(activity); err != nil {
		return nil, errors.New("failed to update activity")
	}
	return activity, nil
}

func (s *TripService) DeleteActivity(activityID, userID uuid.UUID) error {
	_, err := s.tripRepo.FindActivity(activityID)
	if err != nil {
		return errors.New("activity not found")
	}

	// Verify ownership: activity → day → destination → trip → user
	if err := s.verifyActivityOwnership(activityID, userID); err != nil {
		return err
	}

	return s.tripRepo.DeleteActivity(activityID)
}

// verifyActivityOwnership checks that an activity belongs to the given user
// by traversing the chain: activity → trip_day → trip_destination → trip.
func (s *TripService) verifyActivityOwnership(activityID, userID uuid.UUID) error {
	ownerID, err := s.tripRepo.FindActivityOwner(activityID)
	if err != nil {
		return errors.New("activity not found")
	}
	if ownerID != userID {
		return errors.New("access denied")
	}
	return nil
}
