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

// AddExplorerPoints atomically increments a user's explorer points by delta
// (which may be negative). Returns the new total.
func (r *UserRepository) AddExplorerPoints(userID uuid.UUID, delta int) (int, error) {
	if err := r.db.Model(&models.User{}).
		Where("id = ?", userID).
		UpdateColumn("explorer_points", gorm.Expr("explorer_points + ?", delta)).Error; err != nil {
		return 0, err
	}
	var user models.User
	if err := r.db.Select("explorer_points").Where("id = ?", userID).First(&user).Error; err != nil {
		return 0, err
	}
	return user.ExplorerPoints, nil
}

// UpdateBadge sets a user's badge type.
func (r *UserRepository) UpdateBadge(userID uuid.UUID, badge string) error {
	return r.db.Model(&models.User{}).Where("id = ?", userID).Update("badge_type", badge).Error
}

// BumpTokenVersion increments the user's token version, invalidating the ability
// of previously-issued tokens to be refreshed.
func (r *UserRepository) BumpTokenVersion(userID uuid.UUID) error {
	return r.db.Model(&models.User{}).
		Where("id = ?", userID).
		UpdateColumn("token_version", gorm.Expr("token_version + 1")).Error
}

// SoftDelete performs a soft delete on a user account and anonymizes the email
// so the unique index is freed and the person can re-register later. Runs both
// updates in a transaction for consistency.
func (r *UserRepository) SoftDelete(userID uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		anonymized := "deleted+" + userID.String() + "@deleted.waynx"
		if err := tx.Model(&models.User{}).
			Where("id = ?", userID).
			Updates(map[string]interface{}{
				"email":         anonymized,
				"token_version": gorm.Expr("token_version + 1"),
			}).Error; err != nil {
			return err
		}
		return tx.Where("id = ?", userID).Delete(&models.User{}).Error
	})
}

// RecentTripCities returns the distinct cities the user has recently traveled to,
// most-recent first, capped at `limit`. Used to personalize the chat assistant.
func (r *UserRepository) RecentTripCities(userID uuid.UUID, limit int) ([]string, error) {
	var cities []string
	err := r.db.Model(&models.TripDestination{}).
		Joins("JOIN trips ON trips.id = trip_destinations.trip_id AND trips.user_id = ?", userID).
		Where("trips.deleted_at IS NULL").
		Group("trip_destinations.city").
		Order("MAX(trips.created_at) DESC").
		Limit(limit).
		Pluck("trip_destinations.city", &cities).Error
	return cities, err
}
