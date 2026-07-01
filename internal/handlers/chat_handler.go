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

// SendMessage godoc
// @Summary      Send message
// @Description  Send a message to the AI and get a response
// @Tags         chat
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        input body services.SendMessageInput true "Message input"
// @Success      201  {object}  models.ChatMessage
// @Failure      400  {object}  utils.Response
// @Failure      403  {object}  utils.Response
// @Router       /chat [post]
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

// ListSessions godoc
// @Summary      List chat sessions
// @Description  Get a paginated list of chat sessions for the user
// @Tags         chat
// @Produce      json
// @Security     BearerAuth
// @Param        page query int false "Page number" default(1)
// @Param        per_page query int false "Items per page" default(10)
// @Success      200  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /chat/history [get]
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

// GetSession godoc
// @Summary      Get chat session
// @Description  Get a specific chat session with its messages
// @Tags         chat
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Session ID (UUID)"
// @Success      200  {object}  models.ChatSession
// @Failure      400  {object}  utils.Response
// @Failure      403  {object}  utils.Response
// @Failure      404  {object}  utils.Response
// @Router       /chat/{id} [get]
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

// RenameSession godoc
// @Summary      Rename chat session
// @Description  Rename an existing chat session
// @Tags         chat
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Session ID (UUID)"
// @Param        body body object{title=string} true "New title"
// @Success      200  {object}  utils.Response
// @Failure      400  {object}  utils.Response
// @Failure      403  {object}  utils.Response
// @Failure      404  {object}  utils.Response
// @Router       /chat/{id} [put]
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

// DeleteSession godoc
// @Summary      Delete chat session
// @Description  Delete a chat session
// @Tags         chat
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Session ID (UUID)"
// @Success      200  {object}  utils.Response
// @Failure      400  {object}  utils.Response
// @Failure      403  {object}  utils.Response
// @Failure      404  {object}  utils.Response
// @Router       /chat/{id} [delete]
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
