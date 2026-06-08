package repository

import (
	"github.com/google/uuid"
	"github.com/kemit/trip-planner/internal/models"
	"gorm.io/gorm"
)

type ReviewRepository struct {
	db *gorm.DB
}

func NewReviewRepository(db *gorm.DB) *ReviewRepository {
	return &ReviewRepository{db: db}
}

func (r *ReviewRepository) Create(review *models.PlaceReview) error {
	return r.db.Create(review).Error
}

func (r *ReviewRepository) FindByID(id uuid.UUID) (*models.PlaceReview, error) {
	var review models.PlaceReview
	err := r.db.Preload("User").Where("id = ?", id).First(&review).Error
	return &review, err
}

func (r *ReviewRepository) FindByUserAndPlace(userID uuid.UUID, placeID uint) (*models.PlaceReview, error) {
	var review models.PlaceReview
	err := r.db.Where("user_id = ? AND place_id = ?", userID, placeID).First(&review).Error
	return &review, err
}

func (r *ReviewRepository) ListByPlaceID(placeID uint, page, perPage int) ([]models.PlaceReview, int64, error) {
	var reviews []models.PlaceReview
	var total int64

	query := r.db.Model(&models.PlaceReview{}).Where("place_id = ?", placeID)
	query.Count(&total)

	offset := (page - 1) * perPage
	err := query.Preload("User").Offset(offset).Limit(perPage).Order("created_at DESC").Find(&reviews).Error
	return reviews, total, err
}

func (r *ReviewRepository) Update(review *models.PlaceReview) error {
	return r.db.Save(review).Error
}

func (r *ReviewRepository) Delete(id uuid.UUID) error {
	return r.db.Where("id = ?", id).Delete(&models.PlaceReview{}).Error
}
