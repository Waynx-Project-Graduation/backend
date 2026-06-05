package models

import (
	"time"

	"github.com/google/uuid"
)

// SavedPlace represents a user's bookmarked/favorited place
type SavedPlace struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uuid.UUID `gorm:"type:text;index;not null;uniqueIndex:idx_user_place" json:"user_id"`
	PlaceID   uint      `gorm:"index;not null;uniqueIndex:idx_user_place" json:"place_id"`
	CreatedAt time.Time `json:"created_at"`

	// Relations
	User  User  `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Place Place `gorm:"foreignKey:PlaceID" json:"place,omitempty"`
}

func (SavedPlace) TableName() string {
	return "saved_places"
}
