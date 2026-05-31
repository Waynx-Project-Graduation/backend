package repository

import (
	"github.com/google/uuid"
	"github.com/kemit/trip-planner/internal/models"
	"gorm.io/gorm"
)

type ChatRepository struct {
	db *gorm.DB
}

func NewChatRepository(db *gorm.DB) *ChatRepository {
	return &ChatRepository{db: db}
}

// CreateSession creates a new chat session
func (r *ChatRepository) CreateSession(session *models.ChatSession) error {
	return r.db.Create(session).Error
}

// FindSessionByID returns a session with all its messages
func (r *ChatRepository) FindSessionByID(id uuid.UUID) (*models.ChatSession, error) {
	var session models.ChatSession
	err := r.db.
		Preload("Messages", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC")
		}).
		Where("id = ?", id).
		First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// ListSessionsByUserID returns paginated chat sessions for a user (without messages, just metadata)
func (r *ChatRepository) ListSessionsByUserID(userID uuid.UUID, page, perPage int) ([]models.ChatSession, int64, error) {
	var sessions []models.ChatSession
	var total int64

	query := r.db.Model(&models.ChatSession{}).Where("user_id = ?", userID)
	query.Count(&total)

	offset := (page - 1) * perPage
	err := query.
		Offset(offset).Limit(perPage).
		Order("updated_at DESC").
		Find(&sessions).Error

	return sessions, total, err
}

// CreateMessage adds a message to a session
func (r *ChatRepository) CreateMessage(message *models.ChatMessage) error {
	return r.db.Create(message).Error
}

// DeleteSession soft-deletes a chat session (messages cascade)
func (r *ChatRepository) DeleteSession(id uuid.UUID) error {
	// Delete messages first
	r.db.Where("session_id = ?", id).Delete(&models.ChatMessage{})
	return r.db.Where("id = ?", id).Delete(&models.ChatSession{}).Error
}

// UpdateSessionTitle updates the title of a session
func (r *ChatRepository) UpdateSessionTitle(id uuid.UUID, title string) error {
	return r.db.Model(&models.ChatSession{}).Where("id = ?", id).Update("title", title).Error
}

// CountSessionsByUserID returns the total number of chat sessions for a user
func (r *ChatRepository) CountSessionsByUserID(userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.ChatSession{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}
