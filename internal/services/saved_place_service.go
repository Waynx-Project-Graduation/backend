package services

import (
	"errors"

	"github.com/google/uuid"
	"github.com/kemit/trip-planner/internal/models"
	"github.com/kemit/trip-planner/internal/repository"
)

type SavedPlaceService struct {
	savedPlaceRepo *repository.SavedPlaceRepository
	placeRepo      *repository.PlaceRepository
	awarder        PointsAwarder
}

func NewSavedPlaceService(savedPlaceRepo *repository.SavedPlaceRepository, placeRepo *repository.PlaceRepository) *SavedPlaceService {
	return &SavedPlaceService{
		savedPlaceRepo: savedPlaceRepo,
		placeRepo:      placeRepo,
	}
}

// SetPointsAwarder wires the gamification awarder (optional).
func (s *SavedPlaceService) SetPointsAwarder(a PointsAwarder) {
	s.awarder = a
}

// SavePlace adds a place to user's favorites
func (s *SavedPlaceService) SavePlace(userID uuid.UUID, placeID uint) error {
	// Verify place exists
	_, err := s.placeRepo.FindByID(placeID)
	if err != nil {
		return errors.New("place not found")
	}

	// Check if already saved
	isSaved, _ := s.savedPlaceRepo.IsSaved(userID, placeID)
	if isSaved {
		return errors.New("place already saved")
	}

	if err := s.savedPlaceRepo.Save(userID, placeID); err != nil {
		return err
	}

	if s.awarder != nil {
		_ = s.awarder.AwardPoints(userID, PointsPerSave)
	}
	return nil
}

// UnsavePlace removes a place from user's favorites
func (s *SavedPlaceService) UnsavePlace(userID uuid.UUID, placeID uint) error {
	isSaved, _ := s.savedPlaceRepo.IsSaved(userID, placeID)
	if !isSaved {
		return errors.New("place not saved")
	}

	return s.savedPlaceRepo.Unsave(userID, placeID)
}

// ListSavedPlaces returns paginated saved places for a user
func (s *SavedPlaceService) ListSavedPlaces(userID uuid.UUID, page, perPage int) ([]models.SavedPlace, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 50 {
		perPage = 10
	}
	return s.savedPlaceRepo.ListByUserID(userID, page, perPage)
}

// IsSaved checks if a user has saved a specific place
func (s *SavedPlaceService) IsSaved(userID uuid.UUID, placeID uint) (bool, error) {
	return s.savedPlaceRepo.IsSaved(userID, placeID)
}
