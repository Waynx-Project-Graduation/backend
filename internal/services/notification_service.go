package services

import (
	"github.com/google/uuid"
	"github.com/kemit/trip-planner/internal/models"
	"github.com/kemit/trip-planner/internal/repository"
)

type NotificationService struct {
	notifRepo *repository.NotificationRepository
}

func NewNotificationService(notifRepo *repository.NotificationRepository) *NotificationService {
	return &NotificationService{notifRepo: notifRepo}
}

func (s *NotificationService) CreateNotification(userID uuid.UUID, notifType, title, body string) error {
	notif := &models.Notification{
		UserID: userID,
		Type:   notifType,
		Title:  title,
		Body:   body,
	}
	return s.notifRepo.Create(notif)
}

func (s *NotificationService) ListNotifications(userID uuid.UUID, page, perPage int) ([]models.Notification, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 50 {
		perPage = 10
	}
	return s.notifRepo.ListByUserID(userID, page, perPage)
}

func (s *NotificationService) MarkAsRead(id uint, userID uuid.UUID) error {
	return s.notifRepo.MarkAsRead(id, userID)
}

func (s *NotificationService) MarkAllAsRead(userID uuid.UUID) error {
	return s.notifRepo.MarkAllAsRead(userID)
}
