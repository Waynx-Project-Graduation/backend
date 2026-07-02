package services

import (
	"errors"
	"fmt"
	"log"
	"strings"
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

// dayStartHour is the hour each itinerary day begins (09:00).
const dayStartHour = 9

// formatHour converts an hour count into a "HH:MM" clock string, wrapping
// past midnight (e.g. 25 -> "01:00") so times stay valid on long days.
func formatHour(hour int) string {
	h := ((hour % 24) + 24) % 24
	return fmt.Sprintf("%02d:00", h)
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
	StartDate      string             `json:"start_date" binding:"required"` // YYYY-MM-DD
	EndDate        string             `json:"end_date" binding:"required"`   // YYYY-MM-DD
	TravelersCount int                `json:"travelers_count" binding:"required,gte=1"`
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

type CreateActivityInput struct {
	TripDayID     uuid.UUID `json:"trip_day_id" binding:"required"`
	ActivityName  string    `json:"activity_name" binding:"required"`
	Description   string    `json:"description"`
	Category      string    `json:"category"`
	StartTime     string    `json:"start_time"`
	EndTime       string    `json:"end_time"`
	DurationHours int       `json:"duration_hours"`
	EstimatedCost float64   `json:"estimated_cost"`
	ActivityType  string    `json:"activity_type"`
	PlaceID       *uint     `json:"place_id"`
}

// ── Valid enum values (matching WAYNX AI engine dimensions) ─────────────────

var (
	validInterests  = map[string]bool{"history": true, "beach": true, "food": true, "wellness": true, "religious": true, "nature": true, "adventure": true}
	validCompanions = map[string]bool{"solo": true, "couple": true, "family": true, "friends": true}
	validBudgets    = map[string]bool{"low": true, "medium": true, "high": true}
	validAgeGroups  = map[string]bool{"teen": true, "adult": true, "senior": true}
	validCrowdPrefs = map[string]bool{"quiet": true, "moderate": true, "no_preference": true, "crowded": true}
	validSeasons    = map[string]bool{"winter": true, "spring": true, "summer": true, "autumn": true}
)

func mapKeys(m map[string]bool) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return strings.Join(keys, ", ")
}

// validatePreferences checks all 7 dimensions match the WAYNX AI engine's vocabulary.
func validatePreferences(p models.Preferences) error {
	if len(p.Interests) == 0 {
		return errors.New("at least one interest must be provided")
	}
	for _, interest := range p.Interests {
		if !validInterests[interest] {
			return fmt.Errorf("invalid interest %q, must be one of: %s", interest, mapKeys(validInterests))
		}
	}

	if p.TravelCompanion == "" {
		return errors.New("travel_companion must be specified")
	}
	if !validCompanions[p.TravelCompanion] {
		return fmt.Errorf("invalid travel_companion %q, must be one of: %s", p.TravelCompanion, mapKeys(validCompanions))
	}

	if p.Budget == "" {
		return errors.New("budget must be specified")
	}
	if !validBudgets[p.Budget] {
		return fmt.Errorf("invalid budget %q, must be one of: %s", p.Budget, mapKeys(validBudgets))
	}

	if p.AgeGroup == "" {
		return errors.New("age_group must be specified")
	}
	if !validAgeGroups[p.AgeGroup] {
		return fmt.Errorf("invalid age_group %q, must be one of: %s", p.AgeGroup, mapKeys(validAgeGroups))
	}

	if p.CrowdPreference == "" {
		return errors.New("crowd_preference must be specified")
	}
	if !validCrowdPrefs[p.CrowdPreference] {
		return fmt.Errorf("invalid crowd_preference %q, must be one of: %s", p.CrowdPreference, mapKeys(validCrowdPrefs))
	}

	if p.Season == "" {
		return errors.New("season must be specified")
	}
	if !validSeasons[p.Season] {
		return fmt.Errorf("invalid season %q, must be one of: %s", p.Season, mapKeys(validSeasons))
	}

	return nil
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

	// Validate all 7 AI dimensions against the WAYNX engine's vocabulary
	if err := validatePreferences(input.Preferences); err != nil {
		return nil, err
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

	// Call the WAYNX engine FIRST, before persisting anything. This avoids
	// leaving orphan "draft" trips in the database when the engine is down.
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
		log.Printf("WAYNX engine error (trip not created): %v", err)
		// Surface a clean error so the handler returns 503 instead of a
		// misleading success with an empty itinerary.
		return nil, errors.New("AI service unavailable")
	}
	if len(aiResp.Plan.Destinations) == 0 {
		log.Printf("WAYNX engine returned an empty plan for user %s", userID)
		return nil, errors.New("AI service unavailable")
	}

	// Engine succeeded — persist the trip as "planned"
	trip := &models.Trip{
		UserID:         userID,
		Title:          title,
		StartDate:      startDate,
		EndDate:        endDate,
		TravelersCount: input.TravelersCount,
		Preferences:    input.Preferences,
		Status:         "planned",
	}

	if err := s.tripRepo.Create(trip); err != nil {
		return nil, errors.New("failed to create trip")
	}

	// Save the AI-generated multi-city itinerary
	if err := s.saveItinerary(trip.ID, startDate, aiResp); err != nil {
		log.Printf("Failed to save itinerary for trip %s: %v", trip.ID, err)
		// Roll back the parent trip so we don't leave a planned trip with no plan.
		_ = s.tripRepo.Delete(trip.ID)
		return nil, errors.New("failed to save itinerary")
	}

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
			TravelHours:   aiDest.TravelHours,
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
				FreeHours:         aiDay.FreeHours,
			}

			if err := txRepo.CreateTripDays([]models.TripDay{tripDay}); err != nil {
				log.Printf("Trip day save error: %v", err)
				return fmt.Errorf("failed to save day %d", aiDay.DayNumber)
			}

			var activities []models.TripActivity
			// Activities are scheduled sequentially starting at 09:00, chaining
			// each activity by its duration so the frontend has a real timeline.
			clockHour := dayStartHour
			for i, aiAct := range aiDay.Activities {
				// Validate that the AI-generated place ID actually exists in our DB
				var validPlaceID *uint
				if aiAct.PlaceID > 0 {
					if err := txRepo.DB().Where("id = ?", aiAct.PlaceID).First(&models.Place{}).Error; err == nil {
						pid := aiAct.PlaceID
						validPlaceID = &pid
					} else {
						log.Printf("AI returned non-existent place_id %d for activity %q, skipping FK link", aiAct.PlaceID, aiAct.Name)
					}
				}

				startTime := formatHour(clockHour)
				clockHour += aiAct.DurationHours
				endTime := formatHour(clockHour)

				activity := models.TripActivity{
					TripDayID:     tripDay.ID,
					PlaceID:       validPlaceID,
					ActivityName:  aiAct.Name,
					Description:   aiAct.Description,
					Category:      aiAct.Category,
					DurationHours: aiAct.DurationHours,
					StartTime:     startTime,
					EndTime:       endTime,
					Rating:        aiAct.Rating,
					OrderInDay:    i + 1,
					ActivityType:  aiAct.Category,
					MatchScore:    aiAct.MatchScore,
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

func (s *TripService) CreateActivity(tripID, userID uuid.UUID, input CreateActivityInput) (*models.TripActivity, error) {
	trip, err := s.tripRepo.FindByID(tripID)
	if err != nil {
		return nil, errors.New("trip not found")
	}
	if trip.UserID != userID {
		return nil, errors.New("access denied")
	}

	// Validate the day belongs to this trip
	if err := s.tripRepo.ValidateDayBelongsToTrip(input.TripDayID, tripID); err != nil {
		return nil, errors.New("trip day not found in this trip")
	}

	activity := &models.TripActivity{
		TripDayID:     input.TripDayID,
		PlaceID:       input.PlaceID,
		ActivityName:  input.ActivityName,
		Description:   input.Description,
		Category:      input.Category,
		DurationHours: input.DurationHours,
		StartTime:     input.StartTime,
		EndTime:       input.EndTime,
		EstimatedCost: input.EstimatedCost,
		ActivityType:  input.ActivityType,
	}

	if err := s.tripRepo.CreateActivities([]models.TripActivity{*activity}); err != nil {
		return nil, errors.New("failed to create activity")
	}

	return activity, nil
}

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
