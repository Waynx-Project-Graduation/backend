package repository

import (
	"errors"

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

// FindActivity finds a specific activity by ID
func (r *TripRepository) FindActivity(activityID uuid.UUID) (*models.TripActivity, error) {
	var activity models.TripActivity
	err := r.db.Preload("Place").Where("id = ?", activityID).First(&activity).Error
	if err != nil {
		return nil, err
	}
	return &activity, nil
}

// UpdateActivity updates a single activity
func (r *TripRepository) UpdateActivity(activity *models.TripActivity) error {
	return r.db.Save(activity).Error
}

// DeleteActivity deletes a single activity
func (r *TripRepository) DeleteActivity(activityID uuid.UUID) error {
	return r.db.Where("id = ?", activityID).Delete(&models.TripActivity{}).Error
}

// ValidateDayBelongsToTrip checks that a trip day belongs to the given trip.
func (r *TripRepository) ValidateDayBelongsToTrip(dayID, tripID uuid.UUID) error {
	var count int64
	r.db.Model(&models.TripDay{}).Where("id = ? AND trip_id = ?", dayID, tripID).Count(&count)
	if count == 0 {
		return errors.New("trip day not found in this trip")
	}
	return nil
}

// FindActivityOwner returns the user ID that owns the given activity,
// traversing the chain: activity → trip_day → trip_destination → trip.
func (r *TripRepository) FindActivityOwner(activityID uuid.UUID) (uuid.UUID, error) {
	var userID string
	err := r.db.Raw(`
		SELECT t.user_id
		FROM trip_activities ta
		JOIN trip_days td ON td.id = ta.trip_day_id
		JOIN trip_destinations dest ON dest.id = td.trip_destination_id
		JOIN trips t ON t.id = dest.trip_id
		WHERE ta.id = ?
	`, activityID).Scan(&userID).Error
	if err != nil || userID == "" {
		return uuid.Nil, errors.New("activity owner not found")
	}
	parsed, err := uuid.Parse(userID)
	if err != nil {
		return uuid.Nil, errors.New("invalid owner ID")
	}
	return parsed, nil
}
