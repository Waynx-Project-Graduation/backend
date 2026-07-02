package models

import (
	"time"

	"github.com/google/uuid"
)

type Notification struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uuid.UUID `gorm:"type:text;index;not null" json:"user_id"`
	Type      string    `gorm:"type:text;not null" json:"type"`
	Title     string    `gorm:"type:text;not null" json:"title"`
	Body      string    `gorm:"type:text" json:"body"`
	IsRead    bool      `gorm:"default:false" json:"is_read"`
	CreatedAt time.Time `json:"created_at"`

	// Relations
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (Notification) TableName() string {
	return "notifications"
}
