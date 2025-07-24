package usecase

import (
	"context"
	"time"

	"video-call/config"
	"video-call/internal/chat"
	"video-call/internal/models"
	"video-call/pkg/logger"
	"github.com/google/uuid"
)

// usecase implements the chat.UseCase interface.
type usecase struct {
	cfg       *config.Config
	repo      chat.Repository
	redisRepo chat.RedisRepository
	logger    logger.Logger
}

// NewUseCase is the constructor for the chat use case.
func NewUseCase(cfg *config.Config, repo chat.Repository, redisRepo chat.RedisRepository, logger logger.Logger) chat.UseCase {
	return &usecase{
		cfg:       cfg,
		repo:      repo,
		redisRepo: redisRepo,
		logger:    logger,
	}
}

// CreateConversation creates a new conversation.
func (u *usecase) CreateConversation(ctx context.Context, conversation *models.Conversation) error {
	u.logger.Infof(ctx, "Usecase CreateConversation: %+v", conversation)
	
	// Set default values if not provided
	if conversation.ID == "" {
		conversation.ID = uuid.New().String()
	}
	if conversation.CreatedAt.IsZero() {
		conversation.CreatedAt = time.Now()
	}
	
	// Create the conversation
	err := u.repo.CreateConversation(ctx, conversation)
	if err != nil {
		u.logger.Errorf(ctx, "Failed to create conversation: %v", err)
		return err
	}

	return nil
}

// GetConversationByID retrieves a conversation by its ID.
func (u *usecase) GetConversationByID(ctx context.Context, conversationID string) (*models.Conversation, error) {
	u.logger.Infof(ctx, "Usecase GetConversationByID: %s", conversationID)
	
	// Validate conversation ID
	if _, err := uuid.Parse(conversationID); err != nil {
		return nil, chat.ErrInvalidConversationID
	}
	
	return u.repo.GetConversationByID(ctx, conversationID)
}

// GetConversationsByUserID retrieves all conversations for a given user.
// Results are ordered by last activity (newest first).
func (u *usecase) GetConversationsByUserID(ctx context.Context, userID string) ([]*models.Conversation, error) {
	u.logger.Infof(ctx, "Usecase GetConversationsByUserID: %s", userID)
	
	// Validate user ID
	if _, err := uuid.Parse(userID); err != nil {
		return nil, chat.ErrInvalidUserID
	}
	
	return u.repo.GetConversationsByUserID(ctx, userID)
}

// UpdateConversation updates an existing conversation.
func (u *usecase) UpdateConversation(ctx context.Context, conversation *models.Conversation) error {
	u.logger.Infof(ctx, "Usecase UpdateConversation: %+v", conversation)
	return u.repo.UpdateConversation(ctx, conversation)
}

// DeleteConversation deletes a conversation by its ID.
func (u *usecase) DeleteConversation(ctx context.Context, conversationID string) error {
	u.logger.Infof(ctx, "Usecase DeleteConversation: %s", conversationID)
	
	// Validate conversation ID
	if _, err := uuid.Parse(conversationID); err != nil {
		return chat.ErrInvalidConversationID
	}
	
	return u.repo.DeleteConversation(ctx, conversationID)
}

// IsUserInConversation checks if a user is a participant in a conversation.
func (u *usecase) IsUserInConversation(ctx context.Context, userID, conversationID string) (bool, error) {
	u.logger.Infof(ctx, "Usecase IsUserInConversation: userID %s, conversationID %s", userID, conversationID)
	
	// Validate user ID and conversation ID
	if _, err := uuid.Parse(userID); err != nil {
		return false, chat.ErrInvalidUserID
	}
	if _, err := uuid.Parse(conversationID); err != nil {
		return false, chat.ErrInvalidConversationID
	}
	
	return u.repo.IsUserInConversation(ctx, userID, conversationID)
}

// CreateMessage creates a new message.
func (u *usecase) CreateMessage(ctx context.Context, message models.Message) error {
	u.logger.Infof(ctx, "Usecase CreateMessage: %+v", message)
	return u.repo.CreateMessage(ctx, message)
}

// GetMessages retrieves messages for a conversation with pagination.
// Returns messages in descending order by creation time (newest first).
func (u *usecase) GetMessages(ctx context.Context, conversationID string, limit, offset int) ([]*models.Message, error) {
	u.logger.Infof(ctx, "Usecase GetMessages: conversationID=%s, limit=%d, offset=%d", conversationID, limit, offset)
	
	// Validate conversation ID
	if _, err := uuid.Parse(conversationID); err != nil {
		return nil, chat.ErrInvalidConversationID
	}
	
	// Validate pagination parameters
	if limit <= 0 || limit > 100 {
		limit = 50 // Default limit
	}
	if offset < 0 {
		offset = 0
	}
	
	// Get messages from repository
	messages, err := u.repo.GetMessages(ctx, conversationID, limit, offset)
	if err != nil {
		u.logger.Errorf(ctx, "Failed to get messages: %v", err)
		return nil, err
	}

	return messages, nil
}
