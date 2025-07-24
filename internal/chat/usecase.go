//go:generate mockgen -source usecase.go -destination mocks/usecase_mock.go -package mock
package chat

import (
	"context"

	"video-call/internal/models"
)

// UseCase defines the interface for chat-related business logic.
type UseCase interface {
	// CreateConversation creates a new conversation with the given details
	// The conversation ID will be generated if not provided
	CreateConversation(ctx context.Context, conversation *models.Conversation) error

	// GetConversationByID retrieves a conversation by its ID
	GetConversationByID(ctx context.Context, conversationID string) (*models.Conversation, error)

	// GetConversationsByUserID retrieves all conversations for a given user
	// Results are ordered by last activity (newest first)
	GetConversationsByUserID(ctx context.Context, userID string) ([]*models.Conversation, error)

	// UpdateConversation updates an existing conversation
	UpdateConversation(ctx context.Context, conversation *models.Conversation) error

	// DeleteConversation deletes a conversation by its ID
	DeleteConversation(ctx context.Context, conversationID string) error

	// IsUserInConversation checks if a user is a participant in a conversation
	IsUserInConversation(ctx context.Context, userID, conversationID string) (bool, error)

	// CreateMessage creates a new message in a conversation
	CreateMessage(ctx context.Context, message models.Message) error

	// GetMessages retrieves messages for a conversation with pagination
	// Returns messages in descending order by creation time (newest first)
	GetMessages(ctx context.Context, conversationID string, limit, offset int) ([]*models.Message, error)
}
