package repository

import (
	"github.com/google/uuid"
	"github.com/kemit/trip-planner/internal/models"
	"gorm.io/gorm"
)

type TripRepository struct {
	db *gorm.DB
}

func (r *TripRepository) DB() *gorm.DB {
	return r.db
}

func NewTripRepository(db *gorm.DB) *TripRepository {
	return &TripRepository{db: db}
}

// WithTx returns a new repository instance that uses the provided transaction.
func (r *TripRepository) WithTx(tx *gorm.DB) *TripRepository {
	return &TripRepository{db: tx}
}

func (r *TripRepository) Create(trip *models.Trip) error {
	return r.db.Create(trip).Error
}

// FindByID loads a trip with the full nested hierarchy:
// Trip → Destinations → TripDays → Activities → Place
func (r *TripRepository) FindByID(id uuid.UUID) (*models.Trip, error) {
	var trip models.Trip
	err := r.db.
		Preload("Destinations", func(db *gorm.DB) *gorm.DB {
			return db.Order("order_in_trip ASC")
		}).
		Preload("Destinations.TripDays", func(db *gorm.DB) *gorm.DB {
			return db.Order("day_number ASC")
		}).
		Preload("Destinations.TripDays.Activities", func(db *gorm.DB) *gorm.DB {
			return db.Order("order_in_day ASC")
		}).
		Preload("Destinations.TripDays.Activities.Place").
		Where("id = ?", id).
		First(&trip).Error
	if err != nil {
		return nil, err
	}
	return &trip, nil
}

func (r *TripRepository) ListByUserID(userID uuid.UUID, page, perPage int) ([]models.Trip, int64, error) {
	var trips []models.Trip
	var total int64

	query := r.db.Model(&models.Trip{}).Where("user_id = ?", userID)
	query.Count(&total)

	offset := (page - 1) * perPage
	err := query.
		Offset(offset).Limit(perPage).
		Order("created_at DESC").
		Find(&trips).Error

	return trips, total, err
}

func (r *TripRepository) Update(trip *models.Trip) error {
	return r.db.Save(trip).Error
}

func (r *TripRepository) Delete(id uuid.UUID) error {
	return r.db.Where("id = ?", id).Delete(&models.Trip{}).Error
}

// ── Destination Methods ─────────────────────────────────────────────────────

// CreateDestinations creates trip destinations in bulk
func (r *TripRepository) CreateDestinations(destinations []models.TripDestination) error {
	return r.db.Create(&destinations).Error
}

// DeleteDestinations removes all destinations (cascading to days and activities) for a trip
func (r *TripRepository) DeleteDestinations(tripID uuid.UUID) error {
	// Get all destination IDs for this trip
	var destIDs []string
	r.db.Model(&models.TripDestination{}).Where("trip_id = ?", tripID).Pluck("id", &destIDs)

	if len(destIDs) > 0 {
		// Get all day IDs for these destinations
		var dayIDs []string
		r.db.Model(&models.TripDay{}).Where("trip_destination_id IN ?", destIDs).Pluck("id", &dayIDs)

		if len(dayIDs) > 0 {
			// Delete activities first
			r.db.Where("trip_day_id IN ?", dayIDs).Delete(&models.TripActivity{})
		}
		// Delete days
		r.db.Where("trip_destination_id IN ?", destIDs).Delete(&models.TripDay{})
	}

	// Delete destinations
	return r.db.Where("trip_id = ?", tripID).Delete(&models.TripDestination{}).Error
}

// ── Day Methods ─────────────────────────────────────────────────────────────

// CreateTripDays creates trip days in bulk
func (r *TripRepository) CreateTripDays(days []models.TripDay) error {
	return r.db.Create(&days).Error
}

// ── Activity Methods ────────────────────────────────────────────────────────

// CreateActivities creates activities in bulk
func (r *TripRepository) CreateActivities(activities []models.TripActivity) error {
	return r.db.Create(&activities).Error
}
