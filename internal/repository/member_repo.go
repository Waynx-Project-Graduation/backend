package repository

import (
	"github.com/google/uuid"
	"github.com/kemit/trip-planner/internal/models"
	"gorm.io/gorm"
)

type MemberRepository struct {
	db *gorm.DB
}

func NewMemberRepository(db *gorm.DB) *MemberRepository {
	return &MemberRepository{db: db}
}

func (r *MemberRepository) Create(member *models.TripMember) error {
	return r.db.Create(member).Error
}

func (r *MemberRepository) Find(tripID, userID uuid.UUID) (*models.TripMember, error) {
	var member models.TripMember
	err := r.db.Where("trip_id = ? AND user_id = ?", tripID, userID).First(&member).Error
	return &member, err
}

func (r *MemberRepository) ListByTripID(tripID uuid.UUID) ([]models.TripMember, error) {
	var members []models.TripMember
	err := r.db.Preload("User").Where("trip_id = ?", tripID).Order("created_at ASC").Find(&members).Error
	return members, err
}

func (r *MemberRepository) UpdateRole(tripID, userID uuid.UUID, role string) error {
	return r.db.Model(&models.TripMember{}).
		Where("trip_id = ? AND user_id = ?", tripID, userID).
		Update("role", role).Error
}

func (r *MemberRepository) Delete(tripID, userID uuid.UUID) error {
	return r.db.Where("trip_id = ? AND user_id = ?", tripID, userID).Delete(&models.TripMember{}).Error
}
