package services

import (
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/kemit/trip-planner/internal/models"
	"github.com/kemit/trip-planner/internal/repository"
	"github.com/kemit/trip-planner/internal/utils"
	"golang.org/x/crypto/bcrypt"
)

// Sentinel errors so handlers can map to correct HTTP status codes instead of
// blindly returning 500.
var (
	ErrUserNotFound       = errors.New("user not found")
	ErrIncorrectPassword  = errors.New("incorrect current password")
	ErrPasswordNotAllowed = errors.New("password change is not available for social login accounts")
	ErrInvalidInput       = errors.New("invalid input")
)

type UserService struct {
	userRepo       *repository.UserRepository
	savedPlaceRepo *repository.SavedPlaceRepository
	chatRepo       *repository.ChatRepository
}

func NewUserService(userRepo *repository.UserRepository, savedPlaceRepo *repository.SavedPlaceRepository, chatRepo *repository.ChatRepository) *UserService {
	return &UserService{
		userRepo:       userRepo,
		savedPlaceRepo: savedPlaceRepo,
		chatRepo:       chatRepo,
	}
}

// UpdateProfileInput uses pointers so the caller can distinguish "not provided"
// (nil → leave unchanged) from "clear this field" (empty string → set to "").
type UpdateProfileInput struct {
	FullName  *string `json:"full_name" binding:"omitempty,min=2,max=100"`
	City      *string `json:"city" binding:"omitempty,max=100"`
	AvatarURL *string `json:"avatar_url"`
}

// UpdatePreferencesInput uses pointers so a partial update (e.g. only budget)
// leaves the other preferences intact instead of wiping them.
type UpdatePreferencesInput struct {
	Interests       *[]string `json:"interests"`
	TravelCompanion *string   `json:"travel_companion"`
	Budget          *string   `json:"budget"`
	AgeGroup        *string   `json:"age_group"`
	CrowdPreference *string   `json:"crowd_preference"`
	Season          *string   `json:"season"`
}

type ChangePasswordInput struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// UserStats represents profile statistics shown on the profile page
type UserStats struct {
	DestinationsVisited int64 `json:"destinations_visited"`
	AIPlansCreated      int64 `json:"ai_plans_created"`
	ExplorerPoints      int   `json:"explorer_points"`
	SavedPlacesCount    int64 `json:"saved_places_count"`
	ChatSessionsCount   int64 `json:"chat_sessions_count"`
}

func (s *UserService) UpdateProfile(userID uuid.UUID, input UpdateProfileInput) (*models.User, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	// nil → leave unchanged; non-nil → set (empty string clears the field).
	if input.FullName != nil {
		if strings.TrimSpace(*input.FullName) == "" {
			return nil, errors.New("full name cannot be empty")
		}
		user.FullName = strings.TrimSpace(*input.FullName)
	}
	if input.City != nil {
		user.City = strings.TrimSpace(*input.City)
	}
	if input.AvatarURL != nil {
		user.AvatarURL = *input.AvatarURL
	}

	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}
	return user, nil
}

// validateUserPreferences checks the enum fields against the allowed value sets.
// Unlike the trip-creation validator, empty fields are allowed here since a user
// may only partially fill their profile preferences.
func validateUserPreferences(p models.Preferences) error {
	if !utils.IsOneOf(p.Budget, utils.ValidBudgets) {
		return errors.New("invalid budget: must be low, medium, or high")
	}
	if !utils.IsOneOf(p.AgeGroup, utils.ValidAgeGroups) {
		return errors.New("invalid age_group: must be kid, teen, adult, or senior")
	}
	if !utils.IsOneOf(p.Season, utils.ValidSeasons) {
		return errors.New("invalid season")
	}
	if !utils.IsOneOf(p.CrowdPreference, utils.ValidCrowdPreferences) {
		return errors.New("invalid crowd_preference")
	}
	if !utils.IsOneOf(p.TravelCompanion, utils.ValidCompanions) {
		return errors.New("invalid travel_companion")
	}
	if !utils.AllOneOf(p.Interests, utils.ValidInterests) {
		return errors.New("invalid interests: one or more values are not recognized")
	}
	return nil
}

func (s *UserService) UpdatePreferences(userID uuid.UUID, input UpdatePreferencesInput) (*models.User, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	// Start from existing preferences so unspecified fields are preserved.
	prefs := user.Preferences
	if input.Interests != nil {
		prefs.Interests = *input.Interests
	}
	if input.TravelCompanion != nil {
		prefs.TravelCompanion = *input.TravelCompanion
	}
	if input.Budget != nil {
		prefs.Budget = *input.Budget
	}
	if input.AgeGroup != nil {
		prefs.AgeGroup = *input.AgeGroup
	}
	if input.CrowdPreference != nil {
		prefs.CrowdPreference = *input.CrowdPreference
	}
	if input.Season != nil {
		prefs.Season = *input.Season
	}

	if err := validateUserPreferences(prefs); err != nil {
		return nil, err
	}

	user.Preferences = prefs
	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}
	return user, nil
}

// ChangePassword verifies the old password and updates to the new one. It also
// bumps the user's token version so other active sessions can no longer refresh.
func (s *UserService) ChangePassword(userID uuid.UUID, input ChangePasswordInput) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return ErrUserNotFound
	}

	// Social-login accounts have no local password to change.
	if user.AuthProvider != "local" || user.PasswordHash == "" {
		return ErrPasswordNotAllowed
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.OldPassword)); err != nil {
		return ErrIncorrectPassword
	}

	// Hash new password
	hash, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), 12)
	if err != nil {
		return errors.New("failed to hash password")
	}

	if err := s.userRepo.UpdatePasswordHash(userID, string(hash)); err != nil {
		return err
	}
	// Invalidate other sessions' ability to refresh.
	_ = s.userRepo.BumpTokenVersion(userID)
	return nil
}

// GetProfile returns the full user record for the authenticated user.
func (s *UserService) GetProfile(userID uuid.UUID) (*models.User, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// GetStats returns profile statistics for the user
func (s *UserService) GetStats(userID uuid.UUID) (*UserStats, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	destinations, _ := s.userRepo.CountDistinctDestinations(userID)
	trips, _ := s.userRepo.CountTrips(userID)
	savedPlaces, _ := s.savedPlaceRepo.CountByUserID(userID)
	chatSessions, _ := s.chatRepo.CountSessionsByUserID(userID)

	return &UserStats{
		DestinationsVisited: destinations,
		AIPlansCreated:      trips,
		ExplorerPoints:      user.ExplorerPoints,
		SavedPlacesCount:    savedPlaces,
		ChatSessionsCount:   chatSessions,
	}, nil
}

// DeleteAccount soft-deletes the user account (anonymizing the email so it can
// be reused for a future registration).
func (s *UserService) DeleteAccount(userID uuid.UUID) error {
	_, err := s.userRepo.FindByID(userID)
	if err != nil {
		return ErrUserNotFound
	}
	return s.userRepo.SoftDelete(userID)
}

// Explorer points awarded for various actions. Kept together as a single source
// of truth for the gamification system.
const (
	PointsPerTrip   = 50
	PointsPerReview = 10
	PointsPerSave   = 2
)

// PointsAwarder is the minimal interface other services depend on to award
// explorer points, avoiding a hard dependency on the whole UserService.
type PointsAwarder interface {
	AwardPoints(userID uuid.UUID, points int) error
}

// AwardPoints adds points to a user and recomputes their badge tier. Errors are
// intentionally swallowed by callers (gamification is non-critical).
func (s *UserService) AwardPoints(userID uuid.UUID, points int) error {
	total, err := s.userRepo.AddExplorerPoints(userID, points)
	if err != nil {
		return err
	}
	return s.userRepo.UpdateBadge(userID, badgeForPoints(total))
}

// badgeForPoints maps an explorer-points total to a badge tier.
func badgeForPoints(points int) string {
	switch {
	case points >= 1000:
		return "legend"
	case points >= 500:
		return "voyager"
	case points >= 200:
		return "adventurer"
	case points >= 50:
		return "traveler"
	default:
		return "explorer"
	}
}
