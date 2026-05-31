package models

import (
	"time"
)

// Place represents an Egyptian tourist destination.
// Fields match the CSV data source (databaseFiles/kem_places.csv).
type Place struct {
	ID             uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name           string    `gorm:"type:text;not null" json:"name"`
	City           string    `gorm:"type:text;index" json:"city"`
	Category       string    `gorm:"type:text;index" json:"category"`
	BudgetLevel    string    `gorm:"type:text" json:"budget_level"`    // low, medium, high
	BestSeason     string    `gorm:"type:text" json:"best_season"`     // winter, summer, spring, autumn, any
	CrowdLevel     string    `gorm:"type:text" json:"crowd_level"`     // quiet, moderate, crowded
	SuitableFor    string    `gorm:"type:text" json:"suitable_for"`    // comma-separated: family,couple,solo,friends
	SuitableAge    string    `gorm:"type:text" json:"suitable_age"`    // comma-separated: kid,teen,adult,senior
	DurationNeeded int       `gorm:"default:0" json:"duration_needed"` // hours
	Rating         float64   `gorm:"type:real;default:0" json:"rating"`
	Description    string    `gorm:"type:text" json:"description"`
	ThumbnailURL   string    `gorm:"type:text" json:"thumbnail_url"` // externally hosted image URL
	CreatedAt      time.Time `json:"created_at"`

	// Relations
	TripActivities []TripActivity `gorm:"foreignKey:PlaceID" json:"trip_activities,omitempty"`
	SavedBy        []SavedPlace   `gorm:"foreignKey:PlaceID" json:"saved_by,omitempty"`
}

func (Place) TableName() string {
	return "places"
}
