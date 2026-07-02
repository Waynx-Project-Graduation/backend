package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Trip represents a user's travel plan. The AI picks destinations based on
// the user's preferences; the user does NOT choose a single city upfront.
type Trip struct {
	ID             uuid.UUID      `gorm:"type:text;primaryKey" json:"id"`
	UserID         uuid.UUID      `gorm:"type:text;index;not null" json:"user_id"`
	Title          string         `gorm:"type:text;not null" json:"title"`
	StartDate      time.Time      `gorm:"not null" json:"start_date"`
	EndDate        time.Time      `gorm:"not null" json:"end_date"`
	TravelersCount int            `gorm:"default:1" json:"travelers_count"`
	Preferences    Preferences    `gorm:"type:text" json:"preferences"` // same JSON shape as user preferences
	Status         string         `gorm:"type:text;default:'draft';index" json:"status"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`

	// Relations
	User         *User             `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Destinations []TripDestination `gorm:"foreignKey:TripID;constraint:OnDelete:CASCADE;" json:"destinations,omitempty"`
	Expenses     []TripExpense     `gorm:"foreignKey:TripID;constraint:OnDelete:CASCADE;" json:"expenses,omitempty"`
	Members      []TripMember      `gorm:"foreignKey:TripID;constraint:OnDelete:CASCADE;" json:"members,omitempty"`
}

func (t *Trip) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

func (Trip) TableName() string {
	return "trips"
}

// TripDestination groups days for a single city within a trip.
// A trip to Egypt might have destinations: Cairo (3 days), Luxor (2 days), Aswan (2 days).
type TripDestination struct {
	ID            uuid.UUID `gorm:"type:text;primaryKey" json:"id"`
	TripID        uuid.UUID `gorm:"type:text;index;not null" json:"trip_id"`
	City          string    `gorm:"type:text;not null" json:"city"`
	DaysAllocated int       `gorm:"not null" json:"days_allocated"`
	Category      string    `gorm:"type:text" json:"category"` // dominant category for this city
	OrderInTrip   int       `gorm:"default:0" json:"order_in_trip"`
	TravelHours   float64   `gorm:"default:0" json:"travel_hours"`

	// Relations
	Trip     *Trip      `gorm:"foreignKey:TripID" json:"trip,omitempty"`
	TripDays []TripDay `gorm:"foreignKey:TripDestinationID;constraint:OnDelete:CASCADE" json:"trip_days,omitempty"`
}

func (td *TripDestination) BeforeCreate(tx *gorm.DB) error {
	if td.ID == uuid.Nil {
		td.ID = uuid.New()
	}
	return nil
}

func (TripDestination) TableName() string {
	return "trip_destinations"
}

// TripDay represents a single day within a destination.
type TripDay struct {
	ID                  uuid.UUID `gorm:"type:text;primaryKey" json:"id"`
	TripID              uuid.UUID `gorm:"type:text;index;not null" json:"trip_id"`
	TripDestinationID   uuid.UUID `gorm:"type:text;index;not null" json:"trip_destination_id"`
	DayNumber           int       `gorm:"not null" json:"day_number"`
	Date                time.Time `gorm:"not null" json:"date"`
	HoursUsed           int       `gorm:"default:0" json:"hours_used"`
	FreeHours           int       `gorm:"default:0" json:"free_hours"`

	// Relations
	Trip            *Trip            `gorm:"foreignKey:TripID" json:"trip,omitempty"`
	TripDestination *TripDestination `gorm:"foreignKey:TripDestinationID" json:"trip_destination,omitempty"`
	Activities      []TripActivity  `gorm:"foreignKey:TripDayID;constraint:OnDelete:CASCADE" json:"activities,omitempty"`
}

func (td *TripDay) BeforeCreate(tx *gorm.DB) error {
	if td.ID == uuid.Nil {
		td.ID = uuid.New()
	}
	return nil
}

func (TripDay) TableName() string {
	return "trip_days"
}

// TripActivity represents a single activity/place visit within a day.
type TripActivity struct {
	ID            uuid.UUID `gorm:"type:text;primaryKey" json:"id"`
	TripDayID     uuid.UUID `gorm:"type:text;index;not null" json:"trip_day_id"`
	PlaceID       *uint     `gorm:"index" json:"place_id,omitempty"`
	ActivityName  string    `gorm:"type:text;not null" json:"activity_name"`
	Description   string    `gorm:"type:text" json:"description"`
	Category      string    `gorm:"type:text" json:"category"`
	DurationHours int       `gorm:"default:0" json:"duration_hours"`
	StartTime     string    `gorm:"type:text" json:"start_time,omitempty"`
	EndTime       string    `gorm:"type:text" json:"end_time,omitempty"`
	EstimatedCost float64   `gorm:"default:0" json:"estimated_cost"`
	Rating        float64   `gorm:"default:0" json:"rating"`
	OrderInDay    int       `gorm:"default:0" json:"order_in_day"`
	ActivityType  string    `gorm:"type:text" json:"activity_type"`
	MatchScore    float64   `gorm:"default:0" json:"match_score"`

	// Relations
	TripDay *TripDay `gorm:"foreignKey:TripDayID" json:"trip_day,omitempty"`
	Place   *Place  `gorm:"foreignKey:PlaceID" json:"place,omitempty"`
}

func (ta *TripActivity) BeforeCreate(tx *gorm.DB) error {
	if ta.ID == uuid.Nil {
		ta.ID = uuid.New()
	}
	return nil
}

func (TripActivity) TableName() string {
	return "trip_activities"
}
