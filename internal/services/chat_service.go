package services

import (
	"errors"
	"log"

	"github.com/google/uuid"
	"github.com/kemit/trip-planner/internal/models"
	"github.com/kemit/trip-planner/internal/repository"
)

type ChatService struct {
	chatRepo *repository.ChatRepository
	aiClient *AIClient
}

func NewChatService(chatRepo *repository.ChatRepository, aiClient *AIClient) *ChatService {
	return &ChatService{
		chatRepo: chatRepo,
		aiClient: aiClient,
	}
}

type SendMessageInput struct {
	SessionID *uuid.UUID `json:"session_id"`
	Message   string     `json:"message" binding:"required,min=1"`
}

type ChatMessageResponse struct {
	SessionID     uuid.UUID `json:"session_id"`
	UserMessage   MessageDetail `json:"user_message"`
	AIResponse    MessageDetail `json:"ai_response"`
}

type MessageDetail struct {
	ID            uuid.UUID `json:"id"`
	Role          string    `json:"role"`
	Content       string    `json:"content"`
	IsVerified    bool      `json:"is_verified"`
	RelatedPlaces []string  `json:"related_places,omitempty"`
}

// SendMessage sends a user message and returns the AI response
func (s *ChatService) SendMessage(userID uuid.UUID, input SendMessageInput) (*ChatMessageResponse, error) {
	var session *models.ChatSession

	// If no session ID provided, create a new session
	if input.SessionID == nil || *input.SessionID == uuid.Nil {
		session = &models.ChatSession{
			UserID: userID,
			Title:  truncateString(input.Message, 80),
		}
		if err := s.chatRepo.CreateSession(session); err != nil {
			return nil, errors.New("failed to create chat session")
		}
	} else {
		var err error
		session, err = s.chatRepo.FindSessionByID(*input.SessionID)
		if err != nil {
			return nil, errors.New("chat session not found")
		}
		if session.UserID != userID {
			return nil, errors.New("access denied")
		}
	}

	// Build conversation history for AI context BEFORE saving the new message
	var history []ChatHistory
	for _, msg := range session.Messages {
		history = append(history, ChatHistory{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}
	// Append the current user message to history for AI context
	history = append(history, ChatHistory{
		Role:    "user",
		Content: input.Message,
	})

	// Save the user message to DB
	userMsg := &models.ChatMessage{
		SessionID: session.ID,
		Role:      "user",
		Content:   input.Message,
	}
	if err := s.chatRepo.CreateMessage(userMsg); err != nil {
		return nil, errors.New("failed to save message")
	}

	// Call AI service
	aiReq := ChatRequest{
		Message: input.Message,
		History: history,
	}

	aiResp, err := s.aiClient.AskQuestion(aiReq)
	if err != nil {
		log.Printf("AI chat error: %v", err)
		// Save a fallback response
		aiMsg := &models.ChatMessage{
			SessionID:  session.ID,
			Role:       "assistant",
			Content:    "I'm sorry, I'm currently unable to process your request. Please try again later.",
			IsVerified: false,
		}
		_ = s.chatRepo.CreateMessage(aiMsg)
		return &ChatMessageResponse{
			SessionID:   session.ID,
			UserMessage: toMessageDetail(userMsg),
			AIResponse:  toMessageDetail(aiMsg),
		}, nil
	}

	// Save AI response
	aiMsg := &models.ChatMessage{
		SessionID:     session.ID,
		Role:          "assistant",
		Content:       aiResp.Answer,
		IsVerified:    aiResp.IsVerified,
		RelatedPlaces: models.StringSlice(aiResp.RelatedPlaces),
	}
	if err := s.chatRepo.CreateMessage(aiMsg); err != nil {
		return nil, errors.New("failed to save AI response")
	}

	// Update session title if AI suggested one
	if aiResp.SuggestedTitle != "" && len(session.Messages) == 0 {
		_ = s.chatRepo.UpdateSessionTitle(session.ID, aiResp.SuggestedTitle)
	}

	return &ChatMessageResponse{
		SessionID:   session.ID,
		UserMessage: toMessageDetail(userMsg),
		AIResponse:  toMessageDetail(aiMsg),
	}, nil
}

// ListSessions returns paginated chat sessions for a user
func (s *ChatService) ListSessions(userID uuid.UUID, page, perPage int) ([]models.ChatSession, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 50 {
		perPage = 10
	}
	return s.chatRepo.ListSessionsByUserID(userID, page, perPage)
}

// GetSession returns a specific chat session with all messages
func (s *ChatService) GetSession(sessionID, userID uuid.UUID) (*models.ChatSession, error) {
	session, err := s.chatRepo.FindSessionByID(sessionID)
	if err != nil {
		return nil, errors.New("chat session not found")
	}
	if session.UserID != userID {
		return nil, errors.New("access denied")
	}
	return session, nil
}

// DeleteSession deletes a chat session
func (s *ChatService) DeleteSession(sessionID, userID uuid.UUID) error {
	session, err := s.chatRepo.FindSessionByID(sessionID)
	if err != nil {
		return errors.New("chat session not found")
	}
	if session.UserID != userID {
		return errors.New("access denied")
	}
	return s.chatRepo.DeleteSession(sessionID)
}

func toMessageDetail(msg *models.ChatMessage) MessageDetail {
	return MessageDetail{
		ID:            msg.ID,
		Role:          msg.Role,
		Content:       msg.Content,
		IsVerified:    msg.IsVerified,
		RelatedPlaces: msg.RelatedPlaces,
	}
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

