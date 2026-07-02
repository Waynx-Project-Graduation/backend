package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID             uuid.UUID      `gorm:"type:text;primaryKey" json:"id"`
	FullName       string         `gorm:"type:text;not null" json:"full_name"`
	Email          string         `gorm:"type:text;uniqueIndex;not null" json:"email"`
	PasswordHash   string         `gorm:"type:text" json:"-"`
	AuthProvider   string         `gorm:"type:text;default:'local'" json:"auth_provider"`
	ProviderID     string         `gorm:"type:text" json:"-"`
	AvatarURL      string         `gorm:"type:text" json:"avatar_url"`
	City           string         `gorm:"type:text" json:"city"`
	Role           string         `gorm:"type:text;default:'user';not null" json:"role"`
	ExplorerPoints int            `gorm:"default:0" json:"explorer_points"`
	BadgeType      string         `gorm:"type:text;default:'explorer'" json:"badge_type"`
	Preferences    Preferences    `gorm:"type:jsonb" json:"preferences"`
	LastLogin      *time.Time     `json:"last_login,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Trips         []Trip         `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;" json:"trips,omitempty"`
	SavedPlaces   []SavedPlace   `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;" json:"saved_places,omitempty"`
	ChatSessions  []ChatSession  `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;" json:"chat_sessions,omitempty"`
	Notifications []Notification `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;" json:"notifications,omitempty"`
	Reviews       []PlaceReview  `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;" json:"reviews,omitempty"`
	TripMembers   []TripMember   `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;" json:"trip_members,omitempty"`
}

// BeforeCreate generates a UUID before inserting (replaces gen_random_uuid())
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// Preferences stores user travel preferences as JSON in a text column.
// These fields mirror the trip creation inputs so the frontend can
// pre-fill the trip form with the user's saved defaults.
type Preferences struct {
	Interests       []string `json:"interests,omitempty"`        // ["history","adventure",…]
	TravelCompanion string   `json:"travel_companion,omitempty"` // solo / couple / family / friends
	Budget          string   `json:"budget,omitempty"`           // low / medium / high
	AgeGroup        string   `json:"age_group,omitempty"`        // teen / adult / senior
	CrowdPreference string   `json:"crowd_preference,omitempty"` // crowded / quiet / no_preference
	Season          string   `json:"season,omitempty"`           // winter / spring / summer / autumn
}

// Scan implements the sql.Scanner interface for reading JSON from SQLite
func (p *Preferences) Scan(value interface{}) error {
	if value == nil {
		*p = Preferences{}
		return nil
	}
	var bytes []byte
	switch v := value.(type) {
	case string:
		bytes = []byte(v)
	case []byte:
		bytes = v
	default:
		return fmt.Errorf("unsupported type for Preferences: %T", value)
	}
	if len(bytes) == 0 || string(bytes) == "{}" {
		*p = Preferences{}
		return nil
	}
	return json.Unmarshal(bytes, p)
}

// Value implements the driver.Valuer interface for writing JSON to SQLite
func (p Preferences) Value() (driver.Value, error) {
	b, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

func (User) TableName() string {
	return "users"
}
