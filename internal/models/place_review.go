package models

import (
	"time"

	"github.com/google/uuid"
)

type PlaceReview struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	PlaceID   uint      `gorm:"not null;index" json:"place_id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	Rating    int       `gorm:"not null" json:"rating"` // 1-5
	Comment   string    `gorm:"type:text" json:"comment"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relations
	User  User  `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;" json:"user,omitempty"`
	Place Place `gorm:"foreignKey:PlaceID;constraint:OnDelete:CASCADE;" json:"place,omitempty"`
}
