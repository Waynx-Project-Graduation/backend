package repository

import (
	"github.com/google/uuid"
	"github.com/kemit/trip-planner/internal/models"
	"gorm.io/gorm"
)

type SavedPlaceRepository struct {
	db *gorm.DB
}

func NewSavedPlaceRepository(db *gorm.DB) *SavedPlaceRepository {
	return &SavedPlaceRepository{db: db}
}

// Save adds a place to user's saved places
func (r *SavedPlaceRepository) Save(userID uuid.UUID, placeID uint) error {
	saved := &models.SavedPlace{
		UserID:  userID,
		PlaceID: placeID,
	}
	return r.db.Create(saved).Error
}

// Unsave removes a place from user's saved places
func (r *SavedPlaceRepository) Unsave(userID uuid.UUID, placeID uint) error {
	return r.db.Where("user_id = ? AND place_id = ?", userID, placeID).
		Delete(&models.SavedPlace{}).Error
}

// IsSaved checks if a user has saved a specific place
func (r *SavedPlaceRepository) IsSaved(userID uuid.UUID, placeID uint) (bool, error) {
	var count int64
	err := r.db.Model(&models.SavedPlace{}).
		Where("user_id = ? AND place_id = ?", userID, placeID).
		Count(&count).Error
	return count > 0, err
}

// ListByUserID returns all saved places for a user with place details
func (r *SavedPlaceRepository) ListByUserID(userID uuid.UUID, page, perPage int) ([]models.SavedPlace, int64, error) {
	var saved []models.SavedPlace
	var total int64

	query := r.db.Model(&models.SavedPlace{}).Where("user_id = ?", userID)
	query.Count(&total)

	offset := (page - 1) * perPage
	err := query.
		Preload("Place").
		Offset(offset).Limit(perPage).
		Order("created_at DESC").
		Find(&saved).Error

	return saved, total, err
}

// CountByUserID returns the number of saved places for a user
func (r *SavedPlaceRepository) CountByUserID(userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.SavedPlace{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}
