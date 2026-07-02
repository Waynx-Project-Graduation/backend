package repository

import (
	"github.com/google/uuid"
	"github.com/kemit/trip-planner/internal/models"
	"gorm.io/gorm"
)

type NotificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) Create(notif *models.Notification) error {
	return r.db.Create(notif).Error
}

func (r *NotificationRepository) ListByUserID(userID uuid.UUID, page, perPage int) ([]models.Notification, int64, error) {
	var notifs []models.Notification
	var total int64

	query := r.db.Model(&models.Notification{}).Where("user_id = ?", userID)
	query.Count(&total)

	offset := (page - 1) * perPage
	err := query.Offset(offset).Limit(perPage).Order("created_at DESC").Find(&notifs).Error
	return notifs, total, err
}

func (r *NotificationRepository) MarkAsRead(id uint, userID uuid.UUID) error {
	return r.db.Model(&models.Notification{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("is_read", true).Error
}

func (r *NotificationRepository) MarkAllAsRead(userID uuid.UUID) error {
	return r.db.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = false", userID).
		Update("is_read", true).Error
}

func (r *NotificationRepository) UnreadCount(userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = false", userID).
		Count(&count).Error
	return count, err
}
