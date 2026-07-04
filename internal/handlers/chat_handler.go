package handlers

import (
	"encoding/json"
	"fmt"
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

	resp, err := h.chatService.SendMessage(c.Request.Context(), userID, input)
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

// POST /api/chat/stream — Send a message and stream the AI response via SSE.
func (h *ChatHandler) StreamMessage(c *gin.Context) {
	userID := getUserID(c)

	var input services.SendMessageInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	// SSE headers.
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := c.Writer.(interface{ Flush() })
	if !ok {
		utils.InternalError(c, "streaming unsupported")
		return
	}

	writeEvent := func(event, data string) {
		fmt.Fprintf(c.Writer, "event: %s\ndata: %s\n\n", event, data)
		flusher.Flush()
	}

	resp, err := h.chatService.StreamMessage(c.Request.Context(), userID, input, func(chunk string) {
		b, _ := json.Marshal(gin.H{"text": chunk})
		writeEvent("chunk", string(b))
	})
	if err != nil {
		b, _ := json.Marshal(gin.H{"message": err.Error()})
		writeEvent("error", string(b))
		return
	}

	done, _ := json.Marshal(gin.H{
		"session_id": resp.SessionID,
		"message_id": resp.AIResponse.ID,
	})
	writeEvent("done", string(done))
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

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "50"))

	session, messages, total, err := h.chatService.GetSession(sessionID, userID, page, perPage)
	if err != nil {
		if err.Error() == "access denied" {
			utils.Forbidden(c, "you don't have access to this chat session")
			return
		}
		utils.NotFound(c, "chat session not found")
		return
	}

	utils.SuccessWithMeta(c, gin.H{
		"session":  session,
		"messages": messages,
	}, &utils.Meta{Page: page, PerPage: perPage, Total: total})
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
