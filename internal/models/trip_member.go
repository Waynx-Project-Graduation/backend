package models

import (
	"time"

	"github.com/google/uuid"
)

// TripMember represents a user collaborating on a trip
type TripMember struct {
	TripID    uuid.UUID `gorm:"type:uuid;primaryKey" json:"trip_id"`
	UserID    uuid.UUID `gorm:"type:uuid;primaryKey" json:"user_id"`
	Role      string    `gorm:"type:text;not null;default:'viewer'" json:"role"` // owner, editor, viewer
	CreatedAt time.Time `json:"created_at"`

	// Relations
	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;" json:"user,omitempty"`
	Trip Trip `gorm:"foreignKey:TripID;constraint:OnDelete:CASCADE;" json:"trip,omitempty"`
}
