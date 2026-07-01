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

// ListNotifications godoc
// @Summary      List notifications
// @Description  Get a paginated list of notifications for the user
// @Tags         notifications
// @Produce      json
// @Security     BearerAuth
// @Param        page query int false "Page number" default(1)
// @Param        per_page query int false "Items per page" default(10)
// @Success      200  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /notifications [get]
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

// UnreadCount godoc
// @Summary      Get unread count
// @Description  Get the number of unread notifications for the user
// @Tags         notifications
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /notifications/unread-count [get]
func (h *NotificationHandler) UnreadCount(c *gin.Context) {
	userID := getUserID(c)

	count, err := h.notifService.UnreadCount(userID)
	if err != nil {
		utils.InternalError(c, "failed to get unread count")
		return
	}

	utils.Success(c, gin.H{"unread_count": count})
}

// MarkAsRead godoc
// @Summary      Mark notification as read
// @Description  Mark a specific notification as read
// @Tags         notifications
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Notification ID"
// @Success      200  {object}  utils.Response
// @Failure      400  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /notifications/{id}/read [put]
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

// MarkAllAsRead godoc
// @Summary      Mark all notifications as read
// @Description  Mark all unread notifications for the user as read
// @Tags         notifications
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /notifications/read-all [put]
func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	userID := getUserID(c)

	if err := h.notifService.MarkAllAsRead(userID); err != nil {
		utils.InternalError(c, "failed to mark all notifications as read")
		return
	}

	utils.Success(c, gin.H{"message": "all notifications marked as read"})
}
