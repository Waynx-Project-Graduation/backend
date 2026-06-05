package models

import (
	"time"

	"github.com/google/uuid"
)

type TripExpense struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	TripID      uuid.UUID `gorm:"type:uuid;not null;index" json:"trip_id"`
	Amount      float64   `gorm:"not null" json:"amount"`
	Currency    string    `gorm:"size:3;not null;default:'EGP'" json:"currency"`
	Category    string    `gorm:"size:50" json:"category"`
	Description string    `gorm:"size:255" json:"description"`
	Date        time.Time `json:"date"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
