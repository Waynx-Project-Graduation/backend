package repository

import (
	"github.com/google/uuid"
	"github.com/kemit/trip-planner/internal/models"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

func (r *UserRepository) FindByProviderID(provider, providerID string) (*models.User, error) {
	var user models.User
	err := r.db.Where("auth_provider = ? AND provider_id = ?", provider, providerID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdatePasswordHash directly updates the password hash for a user
func (r *UserRepository) UpdatePasswordHash(userID uuid.UUID, hash string) error {
	return r.db.Model(&models.User{}).Where("id = ?", userID).Update("password_hash", hash).Error
}

// CountTrips returns the number of trips for a user
func (r *UserRepository) CountTrips(userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.Trip{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}

// CountDistinctDestinations returns the number of unique destination cities visited
func (r *UserRepository) CountDistinctDestinations(userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.TripDestination{}).
		Joins("JOIN trips ON trips.id = trip_destinations.trip_id AND trips.user_id = ?", userID).
		Where("trips.deleted_at IS NULL").
		Distinct("trip_destinations.city").
		Count(&count).Error
	return count, err
}

// UpdateExplorerPoints updates the explorer points for a user
func (r *UserRepository) UpdateExplorerPoints(userID uuid.UUID, points int) error {
	return r.db.Model(&models.User{}).Where("id = ?", userID).Update("explorer_points", points).Error
}
