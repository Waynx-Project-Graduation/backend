package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kemit/trip-planner/internal/services"
	"github.com/kemit/trip-planner/internal/utils"
)

type ChatHandler struct {
	chatService *services.ChatService
}

func NewChatHandler(chatService *services.ChatService) *ChatHandler {
	return &ChatHandler{chatService: chatService}
}

// POST /api/chat — Send a message and get AI response
func (h *ChatHandler) SendMessage(c *gin.Context) {
	userID := getUserID(c)

	var input services.SendMessageInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	resp, err := h.chatService.SendMessage(userID, input)
	if err != nil {
		if err.Error() == "access denied" {
			utils.Forbidden(c, "you don't have access to this chat session")
			return
		}
		utils.BadRequest(c, err.Error())
		return
	}

	utils.Created(c, resp)
}

// GET /api/chat/history — List chat sessions
func (h *ChatHandler) ListSessions(c *gin.Context) {
	userID := getUserID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "10"))

	sessions, total, err := h.chatService.ListSessions(userID, page, perPage)
	if err != nil {
		utils.InternalError(c, "failed to list chat sessions")
		return
	}

	utils.SuccessWithMeta(c, sessions, &utils.Meta{
		Page:    page,
		PerPage: perPage,
		Total:   total,
	})
}

// GET /api/chat/:id — Get a specific chat session with messages
func (h *ChatHandler) GetSession(c *gin.Context) {
	userID := getUserID(c)
	sessionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "invalid session ID")
		return
	}

	session, err := h.chatService.GetSession(sessionID, userID)
	if err != nil {
		if err.Error() == "access denied" {
			utils.Forbidden(c, "you don't have access to this chat session")
			return
		}
		utils.NotFound(c, "chat session not found")
		return
	}

	utils.Success(c, session)
}

// PUT /api/chat/:id — Rename a chat session
func (h *ChatHandler) RenameSession(c *gin.Context) {
	userID := getUserID(c)
	sessionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "invalid session ID")
		return
	}

	var body struct {
		Title string `json:"title" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	if err := h.chatService.RenameSession(sessionID, userID, body.Title); err != nil {
		if err.Error() == "access denied" {
			utils.Forbidden(c, err.Error())
			return
		}
		utils.NotFound(c, err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "session renamed successfully"})
}

// DELETE /api/chat/:id — Delete a chat session
func (h *ChatHandler) DeleteSession(c *gin.Context) {
	userID := getUserID(c)
	sessionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "invalid session ID")
		return
	}

	if err := h.chatService.DeleteSession(sessionID, userID); err != nil {
		if err.Error() == "access denied" {
			utils.Forbidden(c, "you don't have access to this chat session")
			return
		}
		utils.NotFound(c, err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "chat session deleted successfully"})
}
