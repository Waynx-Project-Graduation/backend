package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ChatSession represents a conversation between a user and the AI assistant
type ChatSession struct {
	ID        uuid.UUID      `gorm:"type:text;primaryKey" json:"id"`
	UserID    uuid.UUID      `gorm:"type:text;index;not null" json:"user_id"`
	Title     string         `gorm:"type:text" json:"title"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Messages []ChatMessage `gorm:"foreignKey:SessionID;constraint:OnDelete:CASCADE" json:"messages,omitempty"`
}

func (cs *ChatSession) BeforeCreate(tx *gorm.DB) error {
	if cs.ID == uuid.Nil {
		cs.ID = uuid.New()
	}
	return nil
}

func (ChatSession) TableName() string {
	return "chat_sessions"
}

// ChatMessage represents a single message in a chat session
type ChatMessage struct {
	ID            uuid.UUID   `gorm:"type:text;primaryKey" json:"id"`
	SessionID     uuid.UUID   `gorm:"type:text;index;not null" json:"session_id"`
	Role          string      `gorm:"type:text;not null" json:"role"` // "user" or "assistant"
	Content       string      `gorm:"type:text;not null" json:"content"`
	IsVerified    bool        `gorm:"default:false" json:"is_verified"`
	RelatedPlaces StringSlice `gorm:"type:text" json:"related_places"` // JSON-encoded string slice
	CreatedAt     time.Time   `json:"created_at"`
}

func (cm *ChatMessage) BeforeCreate(tx *gorm.DB) error {
	if cm.ID == uuid.Nil {
		cm.ID = uuid.New()
	}
	return nil
}

func (ChatMessage) TableName() string {
	return "chat_messages"
}

// StringSlice is a custom type that stores a []string as JSON text in SQLite.
// Replaces pq.StringArray from the PostgreSQL driver.
type StringSlice []string

// Scan implements sql.Scanner for reading JSON from SQLite
func (s *StringSlice) Scan(value interface{}) error {
	if value == nil {
		*s = nil
		return nil
	}
	var bytes []byte
	switch v := value.(type) {
	case string:
		bytes = []byte(v)
	case []byte:
		bytes = v
	default:
		return fmt.Errorf("unsupported type for StringSlice: %T", value)
	}
	if len(bytes) == 0 {
		*s = nil
		return nil
	}
	return json.Unmarshal(bytes, s)
}

// Value implements driver.Valuer for writing JSON to SQLite
func (s StringSlice) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}
	b, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}
