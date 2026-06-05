package services

import (
	"errors"

	"github.com/google/uuid"
	"github.com/kemit/trip-planner/internal/models"
	"github.com/kemit/trip-planner/internal/repository"
	"golang.org/x/crypto/bcrypt"
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

type UpdateProfileInput struct {
	FullName  string `json:"full_name"`
	City      string `json:"city"`
	AvatarURL string `json:"avatar_url"`
}

type UpdatePreferencesInput struct {
	Interests       []string `json:"interests"`
	TravelCompanion string   `json:"travel_companion"`
	Budget          string   `json:"budget"`
	AgeGroup        string   `json:"age_group"`
	CrowdPreference string   `json:"crowd_preference"`
	Season          string   `json:"season"`
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
		return nil, err
	}

	if input.FullName != "" {
		user.FullName = input.FullName
	}
	if input.City != "" {
		user.City = input.City
	}
	if input.AvatarURL != "" {
		user.AvatarURL = input.AvatarURL
	}

	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserService) UpdatePreferences(userID uuid.UUID, input UpdatePreferencesInput) (*models.User, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	user.Preferences = models.Preferences{
		Interests:       input.Interests,
		TravelCompanion: input.TravelCompanion,
		Budget:          input.Budget,
		AgeGroup:        input.AgeGroup,
		CrowdPreference: input.CrowdPreference,
		Season:          input.Season,
	}

	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}
	return user, nil
}

// ChangePassword verifies the old password and updates to the new one
func (s *UserService) ChangePassword(userID uuid.UUID, input ChangePasswordInput) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return errors.New("user not found")
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.OldPassword)); err != nil {
		return errors.New("incorrect current password")
	}

	// Hash new password
	hash, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), 12)
	if err != nil {
		return errors.New("failed to hash password")
	}

	return s.userRepo.UpdatePasswordHash(userID, string(hash))
}

// GetStats returns profile statistics for the user
func (s *UserService) GetStats(userID uuid.UUID) (*UserStats, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
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
