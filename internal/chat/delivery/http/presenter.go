package http

import (
	"time"

	"video-call/internal/models"
)

// Request and Response models

type (
	// CreateConversationRequest represents the request body for creating a conversation
	CreateConversationRequest struct {
		Name      string   `json:"name"`
		IsGroup   bool     `json:"is_group"`
		MemberIDs []string `json:"member_ids"`
	}

	// UpdateConversationRequest represents the request body for updating a conversation
	UpdateConversationRequest struct {
		Name string `json:"name"`
	}

	// SendMessageRequest represents the request body for sending a message
	SendMessageRequest struct {
		Content string `json:"content" binding:"required"`
	}

	// ConversationResponse represents the API response for a conversation
	ConversationResponse struct {
		ID          string    `json:"id"`
		Name        string    `json:
ame"`
		IsGroup     bool      `json:"is_group"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
		MemberCount int       `json:"member_count"`
	}

	// MessageResponse represents the API response for a message
	MessageResponse struct {
		ID           string    `json:"id"`
		Content      string    `json:"content"`
		SenderID     string    `json:"sender_id"`
		ConversationID string  `json:"conversation_id"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
	}

	// ConversationListResponse represents a paginated list of conversations
	ConversationListResponse struct {
		Items      []ConversationResponse `json:"items"`
		TotalCount int                    `json:"total_count"`
		Page       int                    `json:"page"`
		PageSize   int                    `json:"page_size"`
	}

	// MessageListResponse represents a paginated list of messages
	MessageListResponse struct {
		Items      []MessageResponse `json:"items"`
		TotalCount int               `json:"total_count"`
		Page       int               `json:"page"`
		PageSize   int               `json:"page_size"`
	}
)

// Helper functions to convert between domain models and API responses

func toConversationResponse(conv *models.Conversation) ConversationResponse {
	name := ""
	if conv.Name != nil {
		name = *conv.Name
	}

	return ConversationResponse{
		ID:          conv.ID,
		Name:        name,
		IsGroup:     conv.IsGroup,
		CreatedAt:   conv.CreatedAt,
		MemberCount: 1, // TODO: Get actual participant count
	}
}

func toMessageResponse(msg *models.Message) MessageResponse {
	return MessageResponse{
		ID:            msg.ID,
		Content:       msg.Content,
		SenderID:      msg.SenderID,
		ConversationID: msg.ConversationID,
		CreatedAt:     msg.CreatedAt,
	}
}

func toConversationListResponse(convs []*models.Conversation, total, page, pageSize int) ConversationListResponse {
	items := make([]ConversationResponse, len(convs))
	for i, conv := range convs {
		items[i] = toConversationResponse(conv)
	}

	return ConversationListResponse{
		Items:      items,
		TotalCount: total,
		Page:       page,
		PageSize:   pageSize,
	}
}

func toMessageListResponse(msgs []*models.Message, total, page, pageSize int) MessageListResponse {
	items := make([]MessageResponse, len(msgs))
	for i, msg := range msgs {
		items[i] = toMessageResponse(msg)
	}

	return MessageListResponse{
		Items:      items,
		TotalCount: total,
		Page:       page,
		PageSize:   pageSize,
	}
}
