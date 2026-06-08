package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kemit/trip-planner/internal/services"
	"github.com/kemit/trip-planner/internal/utils"
)

type NotificationHandler struct {
	notifService *services.NotificationService
}

func NewNotificationHandler(notifService *services.NotificationService) *NotificationHandler {
	return &NotificationHandler{notifService: notifService}
}

// GET /api/notifications
func (h *NotificationHandler) ListNotifications(c *gin.Context) {
	userID := getUserID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "10"))

	notifs, total, err := h.notifService.ListNotifications(userID, page, perPage)
	if err != nil {
		utils.InternalError(c, "failed to list notifications")
		return
	}

	utils.SuccessWithMeta(c, notifs, &utils.Meta{
		Page:    page,
		PerPage: perPage,
		Total:   total,
	})
}

// GET /api/notifications/unread-count
func (h *NotificationHandler) UnreadCount(c *gin.Context) {
	userID := getUserID(c)

	count, err := h.notifService.UnreadCount(userID)
	if err != nil {
		utils.InternalError(c, "failed to get unread count")
		return
	}

	utils.Success(c, gin.H{"unread_count": count})
}

// PUT /api/notifications/:id/read
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	userID := getUserID(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "invalid notification ID")
		return
	}

	if err := h.notifService.MarkAsRead(uint(id), userID); err != nil {
		utils.InternalError(c, "failed to mark notification as read")
		return
	}

	utils.Success(c, gin.H{"message": "notification marked as read"})
}

// PUT /api/notifications/read-all
func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	userID := getUserID(c)

	if err := h.notifService.MarkAllAsRead(userID); err != nil {
		utils.InternalError(c, "failed to mark all notifications as read")
		return
	}

	utils.Success(c, gin.H{"message": "all notifications marked as read"})
}
